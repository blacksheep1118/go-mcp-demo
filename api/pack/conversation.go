package pack

import (
	"encoding/json"
	"fmt"

	"github.com/FantasyRL/go-mcp-demo/api/model/api"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
)

// BuildConversationItem 将数据库模型转换为API响应
func BuildConversationItem(conversation *model.Conversations) *api.ConversationItem {
	item := &api.ConversationItem{
		ID:        conversation.ID,
		CreatedAt: conversation.CreatedAt.UnixMilli(),
		UpdatedAt: conversation.UpdatedAt.UnixMilli(),
	}

	// 如果有标题则使用，否则生成一个mock标题
	if conversation.Title != nil && *conversation.Title != "" {
		item.Title = *conversation.Title
	} else {
		// 尝试从第一条消息生成标题
		title := generateTitleFromMessages(conversation.Messages)
		item.Title = title
	}

	return item
}

// generateTitleFromMessages 从消息中生成标题
func generateTitleFromMessages(messagesJSON string) string {
	var messages []map[string]interface{}
	if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
		return "新对话"
	}

	// 查找第一条用户消息
	for _, msg := range messages {
		if role, ok := msg["role"].(string); ok && role == "user" {
			if content, ok := msg["content"].(string); ok && content != "" {
				// 取前30个字符作为标题
				if len([]rune(content)) > 30 {
					return string([]rune(content)[:30]) + "..."
				}
				return content
			}
		}
	}

	// 如果没有找到用户消息，返回默认标题
	return fmt.Sprintf("对话 %s", messages[0]["role"])
}

// BuildConversationList 构建对话列表
func BuildConversationList(conversations []*model.Conversations) []*api.ConversationItem {
	items := make([]*api.ConversationItem, 0, len(conversations))
	for _, conversation := range conversations {
		items = append(items, BuildConversationItem(conversation))
	}
	return items
}
