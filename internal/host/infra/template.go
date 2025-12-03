package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"github.com/west2-online/jwch"

	"github.com/FantasyRL/go-mcp-demo/internal/host/repository"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/db"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/query"
	"github.com/openai/openai-go/v2"
	"gorm.io/gorm"
)

var _ repository.TemplateRepository = (*TemplateRepository)(nil)

type TemplateRepository struct {
	db    *db.DB[*query.Query]
	cache *redis.Client
}

func (r *TemplateRepository) IsKeyExist(ctx context.Context, key string) bool {
	return r.cache.Exists(ctx, key).Val() == 1
}

func (r *TemplateRepository) GetTermsCache(ctx context.Context, key string) (terms []string, err error) {
	data, err := r.cache.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("dal.GetTermsCache: cache failed: %w", err)
	}
	if err = sonic.Unmarshal(data, &terms); err != nil {
		return nil, fmt.Errorf("dal.GetTermsCache: Unmarshal failed: %w", err)
	}
	return terms, nil
}

func (r *TemplateRepository) GetCoursesCache(ctx context.Context, key string) (course []*jwch.Course, err error) {
	course = make([]*jwch.Course, 0)
	data, err := r.cache.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("dal.GetCoursesCache: cache failed: %w", err)
	}
	if err = sonic.Unmarshal(data, &course); err != nil {
		return nil, fmt.Errorf("dal.GetCoursesCache: Unmarshal failed: %w", err)
	}
	return course, nil
}

func (r *TemplateRepository) SetTermsCache(ctx context.Context, key string, info []string) error {
	termJson, err := sonic.Marshal(&info)
	if err != nil {
		logger.Errorf("dal.SetTermsCache: Marshal info failed: %v", err)
		return err
	}
	if err = r.cache.Set(ctx, key, termJson, constant.CourseTermsKeyExpire).Err(); err != nil {
		logger.Errorf("dal.SetTermsCache: Set key failed: %v", err)
		return err
	}
	return nil
}

func (r *TemplateRepository) SetCoursesCache(ctx context.Context, key string, course []*jwch.Course) error {
	coursesJson, err := sonic.Marshal(course)
	if err != nil {
		logger.Errorf("dal.SetCoursesCache: Marshal info failed: %v", err)
		return err
	}
	if err = r.cache.Set(ctx, key, coursesJson, constant.CourseTermsKeyExpire).Err(); err != nil {
		logger.Errorf("dal.SetCoursesCache: Set info failed: %v", err)
		return err
	}
	return nil
}

func NewTemplateRepository(db *db.DB[*query.Query], cache *redis.Client) *TemplateRepository {
	return &TemplateRepository{db: db, cache: cache}
}

func (r *TemplateRepository) CreateUserByIDAndName(ctx context.Context, id string, name string) (*model.Users, error) {
	d := r.db.Get(ctx)
	defaultSetting := constant.DefaultUserSettingJSON
	user := &model.Users{
		ID:          id,
		Name:        name,
		SettingJSON: &defaultSetting,
	}
	// 由于user是指针类型，Create方法会自动填充user的其他字段（如时间戳）
	err := d.WithContext(ctx).Users.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *TemplateRepository) GetUserByID(ctx context.Context, id string) (*model.Users, error) {
	d := r.db.Get(ctx)
	user, err := d.WithContext(ctx).Users.Where(d.Users.ID.Eq(id)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return user, nil
}

func (r *TemplateRepository) UpdateUserSetting(ctx context.Context, userID string, settingJSON string) error {
	d := r.db.Get(ctx)
	_, err := d.WithContext(ctx).Users.Where(d.Users.ID.Eq(userID)).Update(d.Users.SettingJSON, settingJSON)
	return err
}

func (r *TemplateRepository) UpsertConversation(
	ctx context.Context,
	userID string,
	conversationID string,
	openaiMessages []openai.ChatCompletionMessageParamUnion,
) error {
	d := r.db.Get(ctx)

	// 先把本次要新增的消息序列化成 JSON
	newBytes, err := json.Marshal(openaiMessages)
	if err != nil {
		return err
	}

	// 查询是否已有这条会话
	q := d.WithContext(ctx)
	conv, err := q.Conversations.
		Where(d.Conversations.ID.Eq(conversationID)).
		Where(d.Conversations.UserID.Eq(userID)).
		First()

	if err != nil {
		// 不存在 => 直接创建
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newConv := &model.Conversations{
				ID:           conversationID,
				UserID:       userID,
				Messages:     string(newBytes), // 直接保存为 JSON 数组
				IsSummarized: 0,
			}
			return q.Conversations.Create(newConv)
		}
		// 其他错误直接返回
		return err
	}

	// 已存在 => 需要把 messages 做 append
	// 假设 messages 字段里始终是 JSON 数组
	var existingMsgs []json.RawMessage
	if len(conv.Messages) > 0 {
		if err := json.Unmarshal([]byte(conv.Messages), &existingMsgs); err != nil {
			// 如果原数据不是合法 JSON，就保守一点直接覆盖
			conv.Messages = string(newBytes)
			return q.Conversations.Save(conv)
		}
	}

	var newMsgs []json.RawMessage
	if err := json.Unmarshal(newBytes, &newMsgs); err != nil {
		return err
	}

	merged := append(existingMsgs, newMsgs...)
	mergedBytes, err := json.Marshal(merged)
	if err != nil {
		return err
	}

	conv.Messages = string(mergedBytes)

	return q.Conversations.Save(conv)
}

func (r *TemplateRepository) GetConversationByID(ctx context.Context, id string) (*model.Conversations, error) {
	d := r.db.Get(ctx)

	conv, err := d.
		WithContext(ctx).
		Conversations.
		Where(d.Conversations.ID.Eq(id)).
		First()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return conv, nil
}

// ListConversationsByUserID 获取用户的所有对话列表
func (r *TemplateRepository) ListConversationsByUserID(ctx context.Context, userID string) ([]*model.Conversations, error) {
	d := r.db.Get(ctx)
	conversations, err := d.WithContext(ctx).Conversations.
		Where(d.Conversations.UserID.Eq(userID)).
		Order(d.Conversations.UpdatedAt.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	return conversations, nil
}

// CreateTodo 创建待办事项
func (r *TemplateRepository) CreateTodo(ctx context.Context, todo *model.Todolists) error {
	d := r.db.Get(ctx)
	return d.WithContext(ctx).Todolists.Create(todo)
}

// GetTodoByID 通过ID获取待办事项
func (r *TemplateRepository) GetTodoByID(ctx context.Context, id string, userID string) (*model.Todolists, error) {
	d := r.db.Get(ctx)
	todo, err := d.WithContext(ctx).Todolists.
		Where(d.Todolists.ID.Eq(id)).
		Where(d.Todolists.UserID.Eq(userID)).
		First()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return todo, nil
}

// ListTodosByUserID 获取用户的所有待办事项列表
func (r *TemplateRepository) ListTodosByUserID(ctx context.Context, userID string) ([]*model.Todolists, error) {
	d := r.db.Get(ctx)
	todos, err := d.WithContext(ctx).Todolists.
		Where(d.Todolists.UserID.Eq(userID)).
		Order(d.Todolists.CreatedAt.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	return todos, nil
}

// ListTodosByStatus 根据状态获取待办事项列表
func (r *TemplateRepository) ListTodosByStatus(ctx context.Context, userID string, status int16) ([]*model.Todolists, error) {
	d := r.db.Get(ctx)
	todos, err := d.WithContext(ctx).Todolists.
		Where(d.Todolists.UserID.Eq(userID)).
		Where(d.Todolists.Status.Eq(status)).
		Order(d.Todolists.CreatedAt.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	return todos, nil
}

// ListTodosByPriority 根据优先级获取待办事项列表
func (r *TemplateRepository) ListTodosByPriority(ctx context.Context, userID string, priority int16) ([]*model.Todolists, error) {
	d := r.db.Get(ctx)
	todos, err := d.WithContext(ctx).Todolists.
		Where(d.Todolists.UserID.Eq(userID)).
		Where(d.Todolists.Priority.Eq(priority)).
		Order(d.Todolists.CreatedAt.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	return todos, nil
}

// ListTodosByCategory 根据分类获取待办事项列表
func (r *TemplateRepository) ListTodosByCategory(ctx context.Context, userID string, category string) ([]*model.Todolists, error) {
	d := r.db.Get(ctx)
	todos, err := d.WithContext(ctx).Todolists.
		Where(d.Todolists.UserID.Eq(userID)).
		Where(d.Todolists.Category.Eq(category)).
		Order(d.Todolists.CreatedAt.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	return todos, nil
}

// ListTodosByFilters 根据多个条件筛选获取待办事项列表
func (r *TemplateRepository) ListTodosByFilters(ctx context.Context, userID string, status *int16, priority *int16, category *string) ([]*model.Todolists, error) {
	d := r.db.Get(ctx)
	q := d.WithContext(ctx).Todolists.Where(d.Todolists.UserID.Eq(userID))

	// 根据状态筛选
	if status != nil {
		q = q.Where(d.Todolists.Status.Eq(*status))
	}

	// 根据优先级筛选
	if priority != nil {
		q = q.Where(d.Todolists.Priority.Eq(*priority))
	}

	// 根据分类筛选
	if category != nil && *category != "" {
		q = q.Where(d.Todolists.Category.Eq(*category))
	}

	// 按创建时间倒序
	todos, err := q.Order(d.Todolists.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}

	return todos, nil
}

// UpdateTodo 更新待办事项
func (r *TemplateRepository) UpdateTodo(ctx context.Context, todo *model.Todolists) error {
	d := r.db.Get(ctx)
	// 使用Save会更新所有字段，包括零值
	return d.WithContext(ctx).Todolists.Save(todo)
}

// DeleteTodo 删除待办事项（软删除）
func (r *TemplateRepository) DeleteTodo(ctx context.Context, id string, userID string) error {
	d := r.db.Get(ctx)
	// 先查询，确保是该用户的待办事项
	todo, err := r.GetTodoByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if todo == nil {
		return gorm.ErrRecordNotFound
	}

	_, err = d.WithContext(ctx).Todolists.
		Where(d.Todolists.ID.Eq(id)).
		Where(d.Todolists.UserID.Eq(userID)).
		Delete()

	return err
}

// ==================== Summarize 相关方法 ====================

// CreateSummary 创建摘要
func (r *TemplateRepository) CreateSummary(ctx context.Context, summary *model.Summaries) error {
	d := r.db.Get(ctx)
	return d.WithContext(ctx).Summaries.Create(summary)
}

// GetSummaryByID 通过ID获取摘要
func (r *TemplateRepository) GetSummaryByID(ctx context.Context, id string) (*model.Summaries, error) {
	d := r.db.Get(ctx)
	summary, err := d.WithContext(ctx).Summaries.
		Where(d.Summaries.ID.Eq(id)).
		First()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return summary, nil
}

// ListSummariesByUserID 获取用户的所有摘要列表（通过conversation关联）
func (r *TemplateRepository) ListSummariesByUserID(ctx context.Context, userID string) ([]*model.Summaries, error) {
	d := r.db.Get(ctx)
	// 先获取用户的所有conversation_id
	conversations, err := d.WithContext(ctx).Conversations.
		Where(d.Conversations.UserID.Eq(userID)).
		Select(d.Conversations.ID).
		Find()

	if err != nil {
		return nil, err
	}

	// 如果用户没有对话，返回空列表
	if len(conversations) == 0 {
		return []*model.Summaries{}, nil
	}

	// 提取所有conversation_id
	conversationIDs := make([]string, 0, len(conversations))
	for _, conv := range conversations {
		conversationIDs = append(conversationIDs, conv.ID)
	}

	// 查询这些conversation_id对应的summaries
	summaries, err := d.WithContext(ctx).Summaries.
		Where(d.Summaries.ConversationID.In(conversationIDs...)).
		Order(d.Summaries.CreatedAt.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	return summaries, nil
}

// UpdateSummary 更新摘要
func (r *TemplateRepository) UpdateSummary(ctx context.Context, summary *model.Summaries) error {
	d := r.db.Get(ctx)
	return d.WithContext(ctx).Summaries.Save(summary)
}

// DeleteSummary 删除摘要（软删除）
func (r *TemplateRepository) DeleteSummary(ctx context.Context, id string) error {
	d := r.db.Get(ctx)
	_, err := d.WithContext(ctx).Summaries.
		Where(d.Summaries.ID.Eq(id)).
		Delete()

	return err
}
