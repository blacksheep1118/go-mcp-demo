package pack

import (
	"github.com/FantasyRL/go-mcp-demo/api/model/model"
	"github.com/west2-online/jwch"
	"strings"
)

func BuildCourse(courses []*jwch.Course) []*model.Course {
	var courseList []*model.Course
	for _, course := range courses {
		courseList = append(courseList, &model.Course{
			Name:             course.Name,
			Syllabus:         course.Syllabus,
			Lessonplan:       course.LessonPlan,
			Teacher:          course.Teacher,
			ScheduleRules:    buildScheduleRules(course.ScheduleRules),
			RawScheduleRules: course.RawScheduleRules,
			RawAdjust:        course.RawAdjust,
			Remark:           course.Remark,
			ExamType:         course.ExamType,
		})
	}
	return courseList
}

func buildScheduleRules(scheduleRules []jwch.CourseScheduleRule) []*model.CourseScheduleRule {
	var res []*model.CourseScheduleRule
	for _, scheduleRule := range scheduleRules {
		res = append(res, buildScheduleRule(scheduleRule))
	}
	return res
}

func buildScheduleRule(scheduleRule jwch.CourseScheduleRule) *model.CourseScheduleRule {
	return &model.CourseScheduleRule{
		Location:   normalizeCourseLocation(scheduleRule.Location),
		StartClass: int64(scheduleRule.StartClass),
		EndClass:   int64(scheduleRule.EndClass),
		StartWeek:  int64(scheduleRule.StartWeek),
		EndWeek:    int64(scheduleRule.EndWeek),
		Weekday:    int64(scheduleRule.Weekday),
		Single:     scheduleRule.Single,
		Double:     scheduleRule.Double,
		Adjust:     scheduleRule.Adjust,
	}
}

func normalizeCourseLocation(location string) string {
	if location == "旗山物理实验教学中心" || location == "铜盘教学楼" {
		return location
	}

	// 去除 {铜盘,旗山} 前缀
	location = strings.TrimPrefix(location, "铜盘")
	location = strings.TrimPrefix(location, "旗山")

	return location
}
