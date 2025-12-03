package application

import (
	"time"

	"github.com/FantasyRL/go-mcp-demo/api/model/api"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
)

func (h *Host) TemplateLogic(req *api.TemplateRequest) (*model.Users, error) {
	// 这里编排你的业务逻辑
	return h.templateRepository.CreateUserByIDAndName(h.ctx, req.TemplateId, "test")
}

// CreateTodoLogic 创建待办事项
func (h *Host) CreateTodoLogic(req *api.CreateTodoRequest, userID string) (string, error) {
	// 将unix毫秒时间戳转换为time.Time
	startTime := time.UnixMilli(req.StartTime)
	endTime := time.UnixMilli(req.EndTime)

	// 构造待办事项对象
	todo := &model.Todolists{
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		StartTime: startTime,
		EndTime:   endTime,
		IsAllDay:  0,            // 默认不是全天
		Status:    0,            // 默认未完成
		Priority:  req.Priority, // 必填字段
	}

	// 设置可选字段
	if req.IsAllDay != nil {
		todo.IsAllDay = *req.IsAllDay
	}

	if req.RemindAt != nil {
		remindTime := time.UnixMilli(*req.RemindAt)
		todo.RemindAt = &remindTime
	}

	if req.Category != nil {
		todo.Category = req.Category
	}

	// 创建待办事项
	err := h.templateRepository.CreateTodo(h.ctx, todo)
	if err != nil {
		return "", err
	}

	return todo.ID, nil
}

// GetTodoLogic 获取待办事项详情
func (h *Host) GetTodoLogic(id string, userID string) (*model.Todolists, error) {
	todo, err := h.templateRepository.GetTodoByID(h.ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if todo == nil {
		return nil, errno.NewErrNo(errno.BizNotExist, "待办事项不存在")
	}

	return todo, nil
}

// ListTodoLogic 获取用户的所有待办事项列表
func (h *Host) ListTodoLogic(userID string) ([]*model.Todolists, error) {
	todos, err := h.templateRepository.ListTodosByUserID(h.ctx, userID)
	if err != nil {
		return nil, err
	}
	return todos, nil
}

// SearchTodoLogic 搜索待办事项（支持多条件筛选）
func (h *Host) SearchTodoLogic(userID string, status *int16, priority *int16, category *string) ([]*model.Todolists, error) {
	// 有筛选条件时使用filters方法
	todos, err := h.templateRepository.ListTodosByFilters(h.ctx, userID, status, priority, category)
	if err != nil {
		return nil, err
	}

	return todos, nil
}

// UpdateTodoLogic 更新待办事项
func (h *Host) UpdateTodoLogic(req *api.UpdateTodoRequest, userID string) error {
	// 先查询待办事项是否存在
	todo, err := h.templateRepository.GetTodoByID(h.ctx, req.ID, userID)
	if err != nil {
		return err
	}

	if todo == nil {
		return errno.NewErrNo(errno.BizNotExist, "待办事项不存在")
	}

	// 更新字段
	if req.Title != nil {
		todo.Title = *req.Title
	}

	if req.Content != nil {
		todo.Content = *req.Content
	}

	if req.StartTime != nil {
		startTime := time.UnixMilli(*req.StartTime)
		todo.StartTime = startTime
	}

	if req.EndTime != nil {
		endTime := time.UnixMilli(*req.EndTime)
		todo.EndTime = endTime
	}

	if req.IsAllDay != nil {
		todo.IsAllDay = *req.IsAllDay
	}

	if req.Status != nil {
		todo.Status = *req.Status
	}

	if req.Priority != nil {
		todo.Priority = *req.Priority
	}

	if req.RemindAt != nil {
		remindTime := time.UnixMilli(*req.RemindAt)
		todo.RemindAt = &remindTime
	}

	if req.Category != nil {
		todo.Category = req.Category
	}

	// 更新待办事项
	return h.templateRepository.UpdateTodo(h.ctx, todo)
}

// DeleteTodoLogic 删除待办事项
func (h *Host) DeleteTodoLogic(id string, userID string) error {
	return h.templateRepository.DeleteTodo(h.ctx, id, userID)
}
