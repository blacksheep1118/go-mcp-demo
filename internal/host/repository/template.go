package repository

import (
	"context"

	"github.com/west2-online/jwch"

	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
	"github.com/openai/openai-go/v2"
)

// TemplateRepository 根据实际需求定义上层访问的接口，在下层infra做具体方法的实现
type TemplateRepository interface {
	/*
		database related methods
	*/

	// CreateUserByIDAndName 这里定义一些示例方法，这样就可以先编排逻辑了，后面再去下层实现接口
	CreateUserByIDAndName(ctx context.Context, id string, name string) (*model.Users, error)
	// GetUserByID 通过ID获取用户信息
	GetUserByID(ctx context.Context, id string) (*model.Users, error)
	// UpdateUserSetting 更新用户设置JSON
	UpdateUserSetting(ctx context.Context, userID string, settingJSON string) error
	// UpsertConversation 插入或更新对话记录
	UpsertConversation(ctx context.Context, userID string, conversationID string, openaiMessages []openai.ChatCompletionMessageParamUnion) error
	// GetConversationByID 通过ID获取对话记录
	GetConversationByID(ctx context.Context, id string) (*model.Conversations, error)
	// ListConversationsByUserID 获取用户的所有对话列表
	ListConversationsByUserID(ctx context.Context, userID string) ([]*model.Conversations, error)

	// CreateTodo 创建待办事项
	CreateTodo(ctx context.Context, todo *model.Todolists) error
	// GetTodoByID 通过ID获取待办事项
	GetTodoByID(ctx context.Context, id string, userID string) (*model.Todolists, error)
	// ListTodosByUserID 获取用户的所有待办事项列表
	ListTodosByUserID(ctx context.Context, userID string) ([]*model.Todolists, error)
	// ListTodosByStatus 根据状态获取待办事项列表
	ListTodosByStatus(ctx context.Context, userID string, status int16) ([]*model.Todolists, error)
	// ListTodosByPriority 根据优先级获取待办事项列表
	ListTodosByPriority(ctx context.Context, userID string, priority int16) ([]*model.Todolists, error)
	// ListTodosByCategory 根据分类获取待办事项列表
	ListTodosByCategory(ctx context.Context, userID string, category string) ([]*model.Todolists, error)
	// ListTodosByFilters 根据多个条件筛选获取待办事项列表
	ListTodosByFilters(ctx context.Context, userID string, status *int16, priority *int16, category *string) ([]*model.Todolists, error)
	// UpdateTodo 更新待办事项
	UpdateTodo(ctx context.Context, todo *model.Todolists) error
	// DeleteTodo 删除待办事项
	DeleteTodo(ctx context.Context, id string, userID string) error

	// CreateSummary 创建摘要
	CreateSummary(ctx context.Context, summary *model.Summaries) error
	// GetSummaryByID 通过ID获取摘要
	GetSummaryByID(ctx context.Context, id string) (*model.Summaries, error)
	// ListSummariesByUserID 获取用户的所有摘要列表（通过conversation关联）
	ListSummariesByUserID(ctx context.Context, userID string) ([]*model.Summaries, error)
	// UpdateSummary 更新摘要
	UpdateSummary(ctx context.Context, summary *model.Summaries) error
	// DeleteSummary 删除摘要
	DeleteSummary(ctx context.Context, id string) error

	/*
		redis related methods
	*/
	// IsKeyExist 判断缓存中是否存在某个 key
	IsKeyExist(ctx context.Context, key string) bool
	// GetTermsCache 获取学期列表缓存
	GetTermsCache(ctx context.Context, key string) (terms []string, err error)
	// GetCoursesCache 获取课程列表缓存
	GetCoursesCache(ctx context.Context, key string) (course []*jwch.Course, err error)
	// SetCoursesCache 设置课程列表缓存
	SetCoursesCache(ctx context.Context, key string, course []*jwch.Course) error
	// SetTermsCache 设置学期列表缓存
	SetTermsCache(ctx context.Context, key string, info []string) error
}
