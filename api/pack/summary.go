package pack

import (
	"encoding/json"

	"github.com/FantasyRL/go-mcp-demo/api/model/api"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
)

// BuildSummaryItem 将数据库模型转换为API响应
func BuildSummaryItem(summary *model.Summaries) *api.SummaryItem {
	item := &api.SummaryItem{
		ID:             summary.ID,
		ConversationID: summary.ConversationID,
		SummaryText:    summary.SummaryText,
		CreatedAt:      summary.CreatedAt.UnixMilli(),
		UpdatedAt:      summary.UpdatedAt.UnixMilli(),
	}

	// 解析tags
	var tags []string
	if err := json.Unmarshal([]byte(summary.Tags), &tags); err == nil {
		item.Tags = tags
	} else {
		item.Tags = []string{}
	}

	// 转换tool_calls为字符串
	item.ToolCallsJSON = summary.ToolCalls

	// 解析notes
	var notes map[string]string
	if err := json.Unmarshal([]byte(summary.Notes), &notes); err == nil {
		item.Notes = notes
	} else {
		item.Notes = make(map[string]string)
	}

	return item
}

// BuildSummaryList 构建摘要列表
func BuildSummaryList(summaries []*model.Summaries) []*api.SummaryItem {
	items := make([]*api.SummaryItem, 0, len(summaries))
	for _, summary := range summaries {
		items = append(items, BuildSummaryItem(summary))
	}
	return items
}
