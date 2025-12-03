package application

import (
	"errors"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/api/model/api"
	"github.com/FantasyRL/go-mcp-demo/api/model/model"
	"github.com/FantasyRL/go-mcp-demo/api/pack"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
	"github.com/west2-online/jwch"
	"slices"
	"sort"
	"strings"
)

// GetTermsList 会返回当前用户含有课表的学期信息
func (h *Host) GetTermsList() ([]string, error) {
	loginData, e := utils.ExtractLoginData(h.ctx)
	if !e {
		return nil, errno.ParamError
	}

	key := fmt.Sprintf("terms:%s", loginData.ID)
	if h.templateRepository.IsKeyExist(h.ctx, key) {
		terms, err := h.templateRepository.GetTermsCache(h.ctx, key)
		if err = base.HandleJwchError(err); err != nil {
			return nil, fmt.Errorf("service.GetTermList: Get terms cache fail: %w", err)
		}
		return terms, nil
	}

	stu := jwch.NewStudent().WithLoginData(loginData.ID, utils.ParseCookies(loginData.Cookie))
	terms, err := stu.GetTerms()
	if err = base.HandleJwchError(err); err != nil {
		return nil, fmt.Errorf("service.GetTermList: Get terms fail: %w", err)
	}
	go func() {
		err = h.templateRepository.SetTermsCache(h.ctx, loginData.ID, terms.Terms)
		if err = base.HandleJwchError(err); err != nil {
			logger.Errorf("service.GetTermList: set cache fail: %v", err)
		}
	}()

	return terms.Terms, nil
}

func (h *Host) GetCourseList(req *api.CourseListRequest) ([]*model.Course, error) {
	loginData, e := utils.ExtractLoginData(h.ctx)
	if !e {
		return nil, errno.ParamError
	}
	termKey := fmt.Sprintf("terms:%s", loginData.ID)
	courseKey := fmt.Sprintf("course:%s:%s", loginData.ID, req.Term)
	terms := new(jwch.Term)
	// 学期缓存存在
	isRefresh := false
	if req.IsRefresh != nil {
		isRefresh = *req.IsRefresh
	}
	if !isRefresh && h.templateRepository.IsKeyExist(h.ctx, termKey) {
		termsList, err := h.templateRepository.GetTermsCache(h.ctx, termKey)
		if err != nil {
			return nil, fmt.Errorf("service.GetCourseList: Get term fail: %w", err)
		}
		terms.Terms = termsList
		courses, err := h.templateRepository.GetCoursesCache(h.ctx, courseKey)
		if err != nil {
			return nil, fmt.Errorf("service.GetCourseList: Get courses fail: %w", err)
		}
		return h.removeDuplicateCourses(pack.BuildCourse(courses)), nil
	}

	stu := jwch.NewStudent().WithLoginData(loginData.ID, utils.ParseCookies(loginData.Cookie))

	terms, err := stu.GetTerms()
	if err = base.HandleJwchError(err); err != nil {
		return nil, fmt.Errorf("service.GetCourseList: Get terms failed: %w", err)
	}

	// validate term
	if !slices.Contains(terms.Terms, req.Term) {
		return nil, errors.New("service.GetCourseList: Invalid term")
	}

	courses, err := stu.GetSemesterCourses(req.Term, terms.ViewState, terms.EventValidation)
	if err = base.HandleJwchError(err); err != nil {
		return nil, fmt.Errorf("service.GetCourseList: Get semester courses failed: %w", err)
	}

	err = h.templateRepository.SetCoursesCache(h.ctx, courseKey, courses)
	if err != nil {
		return nil, fmt.Errorf("service.GetCourseList: Set courses cache fail: %w", err)
	}
	err = h.templateRepository.SetTermsCache(h.ctx, courseKey, terms.Terms)
	if err != nil {
		return nil, fmt.Errorf("service.GetCourseList: Set terms cache fail: %w", err)
	}

	return h.removeDuplicateCourses(pack.BuildCourse(courses)), nil
}

func (h *Host) removeDuplicateCourses(courses []*model.Course) []*model.Course {
	seen := make(map[string]struct{})
	var result []*model.Course

	for _, c := range courses {
		srIDs := make([]string, 0, len(c.ScheduleRules))
		for _, rule := range c.ScheduleRules {
			part := fmt.Sprintf("%d-%d-%d-%d",
				rule.StartClass, rule.EndClass,
				rule.StartWeek, rule.EndWeek)
			srIDs = append(srIDs, part)
		}
		sort.Strings(srIDs)

		// 把“课程名 + 教师 + 排课信息”拼成一个全局唯一的 key
		identifier := fmt.Sprintf("%s-%s-%s", c.Name, c.Teacher, strings.Join(srIDs, "|"))

		// 如果 map 里还没出现过这个标识，那就是新课程
		if _, exists := seen[identifier]; !exists {
			seen[identifier] = struct{}{}
			result = append(result, c)
		}
	}

	return result
}
