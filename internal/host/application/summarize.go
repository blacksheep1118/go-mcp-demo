package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/internal/host/infra"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/model"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	openai "github.com/openai/openai-go/v2"
)

const (
	summarizePromptName = "summarize"
)

type SummarizeResult struct {
	Summary       string
	Tags          []string
	ToolCallsJSON string
	// NotesJSON 直接透传原始 JSON bytes（用于写入 jsonb 和对外返回）
	NotesJSON json.RawMessage
	// 便于 handler 层直接使用的结构化 notes
	Notes map[string]string
}

type conversationSummaryPayload struct {
	Summary   string          `json:"summary"`
	Tags      []string        `json:"tags"`
	ToolCalls json.RawMessage `json:"tool_calls"`
	// 接收原始 notes 数据用于直接透传
	NotesRaw json.RawMessage `json:"notes"`
	// 相关的已有summary ID，如果为空则创建新的
	RelatedSummaryID string `json:"related_summary_id,omitempty"`
}

var atNotationRe = regexp.MustCompile(`^@?\{\s*(.*)\s*\}$`)

func parseAtNotation(input string) (map[string]string, bool) {
	match := atNotationRe.FindStringSubmatch(strings.TrimSpace(input))
	if len(match) < 2 {
		return nil, false
	}
	body := match[1]
	res := make(map[string]string)
	for _, part := range strings.Split(body, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		key := strings.TrimSpace(kv[0])
		val := ""
		if len(kv) > 1 {
			val = strings.TrimSpace(kv[1])
		}
		if key != "" {
			res[key] = val
		}
	}
	return res, true
}

func jsonToStringMap(b json.RawMessage) map[string]string {
	res := map[string]string{}
	if len(b) == 0 || string(bytes.TrimSpace(b)) == "null" {
		return res
	}
	// 优先 map[string]string
	var ms map[string]string
	if err := json.Unmarshal(b, &ms); err == nil {
		return ms
	}
	// map[string]any
	var ma map[string]any
	if err := json.Unmarshal(b, &ma); err == nil {
		for k, v := range ma {
			bb, _ := json.Marshal(v)
			res[k] = strings.Trim(string(bb), "\"")
		}
		return res
	}
	// 解析字符串（含 @{...} 兼容）
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if parsed, ok := parseAtNotation(s); ok {
			return parsed
		}
		res["value"] = s
		return res
	}
	// 兜底保存原始文本
	res["raw"] = string(b)
	return res
}

// 简化 Unmarshal：只接收原始 bytes，后续校验在 parseConversationSummary 做
func (p *conversationSummaryPayload) UnmarshalJSON(data []byte) error {
	var aux struct {
		Summary          string          `json:"summary"`
		Tags             []string        `json:"tags"`
		ToolCalls        json.RawMessage `json:"tool_calls"`
		Notes            json.RawMessage `json:"notes"`
		RelatedSummaryID string          `json:"related_summary_id"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Summary = aux.Summary
	p.Tags = aux.Tags
	p.ToolCalls = aux.ToolCalls
	p.NotesRaw = aux.Notes
	p.RelatedSummaryID = aux.RelatedSummaryID
	return nil
}

func (h *Host) buildSummarizePrompt(ctx context.Context, conversationID string, userID string) (string, error) {
	tpl, err := infra.LoadPrompt(summarizePromptName)
	if err != nil {
		return "", fmt.Errorf("load summarize prompt: %w", err)
	}

	history, err := h.selectConversationHistory(ctx, conversationID)
	if err != nil {
		return "", fmt.Errorf("get conversation history: %w", err)
	}

	// 获取用户的现有summaries
	existingSummaries, err := h.templateRepository.ListSummariesByUserID(ctx, userID)
	if err != nil {
		logger.Warnf("get existing summaries failed: %v", err)
		existingSummaries = []*model.Summaries{}
	}

	// 构建现有summaries的信息
	summariesInfo := h.buildExistingSummariesInfo(existingSummaries)
	logger.Infof("buildSummarizePrompt: userID=%s, existingSummaries count=%d, summariesInfo length=%d",
		userID, len(existingSummaries), len(summariesInfo))

	templateData := map[string]any{
		"conversation_id":      conversationID,
		"conversation_history": strings.Join(history, "\n"),
		"existing_summaries":   summariesInfo,
		"generated_at":         time.Now().Format(time.RFC3339),
	}
	logger.Debugf("template data keys: %v", func() []string {
		keys := make([]string, 0, len(templateData))
		for k := range templateData {
			keys = append(keys, k)
		}
		return keys
	}())

	rendered, err := infra.RenderPrompt(tpl, templateData)
	if err != nil {
		return "", fmt.Errorf("render summarize prompt: %w", err)
	}
	return rendered, nil
}

func (h *Host) invokeSummarizeModel(ctx context.Context, prompt string) (string, error) {
	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(config.AiProvider.Model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(prompt),
			openai.UserMessage("请严格输出 JSON，字段：summary、tags、tool_calls、notes、related_summary_id(如果与现有知识库相关则返回id，否则为空字符串)。"),
		},
	}

	if config.AiProvider.Options.MaxTokens != nil {
		params.MaxTokens = openai.Int(int64(*config.AiProvider.Options.MaxTokens))
	}
	if config.AiProvider.Options.Temperature != nil {
		params.Temperature = openai.Float(*config.AiProvider.Options.Temperature)
	}
	if config.AiProvider.Options.TopP != nil {
		params.TopP = openai.Float(*config.AiProvider.Options.TopP)
	}

	resp, err := h.aiProviderCli.ChatOpenAI(ctx, params)
	if err != nil {
		return "", fmt.Errorf("call openai summarize: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("openai summarize: empty choices")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// buildExistingSummariesInfo 构建现有summaries的信息字符串
func (h *Host) buildExistingSummariesInfo(summaries []*model.Summaries) string {
	if len(summaries) == 0 {
		return "暂无已有知识库"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("用户已有 %d 个知识库：\n\n", len(summaries)))

	for i, summary := range summaries {
		// 解析tags
		var tags []string
		if err := json.Unmarshal([]byte(summary.Tags), &tags); err == nil {
			builder.WriteString(fmt.Sprintf("%d. ID: %s\n", i+1, summary.ID))
			builder.WriteString(fmt.Sprintf("   标题: %s\n", summary.SummaryText[:min(50, len(summary.SummaryText))]))
			builder.WriteString(fmt.Sprintf("   标签: %v\n", tags))
			builder.WriteString(fmt.Sprintf("   创建时间: %s\n\n", summary.CreatedAt.Format("2006-01-02 15:04:05")))
		}
	}

	return builder.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *Host) selectConversationHistory(ctx context.Context, conversationID string) ([]string, error) {
	// 从数据库获取对话记录
	conversation, err := h.templateRepository.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("get conversation failed: %w", err)
	}
	if conversation == nil {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}

	// 解析messages字段（jsonb格式）
	// Messages格式示例：[{"role":"user","content":"..."},{"role":"assistant","content":"..."}]
	var messages []map[string]interface{}
	if err := json.Unmarshal([]byte(conversation.Messages), &messages); err != nil {
		logger.Errorf("unmarshal conversation messages failed: %v", err)
		return nil, fmt.Errorf("parse conversation messages failed: %w", err)
	}

	// 转换为字符串数组格式
	history := make([]string, 0, len(messages))
	for _, msg := range messages {
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)

		// 格式化为易读的格式
		var roleLabel string
		switch role {
		case "user":
			roleLabel = "User"
		case "assistant":
			roleLabel = "Assistant"
		case "system":
			roleLabel = "System"
		case "tool":
			roleLabel = "Tool"
		default:
			roleLabel = strings.Title(role)
		}

		history = append(history, fmt.Sprintf("[%s] %s", roleLabel, content))
	}

	return history, nil
}

func parseConversationSummary(raw string) (*conversationSummaryPayload, error) {
	clean := sanitizeJSONBlock(raw)
	payload := new(conversationSummaryPayload)
	if err := json.Unmarshal([]byte(clean), payload); err != nil {
		logger.Errorf("parseConversationSummary raw=%s error=%v", raw, err)
		return nil, fmt.Errorf("parse summarize response: %w", err)
	}
	if payload.Summary == "" {
		return nil, errors.New("AI 返回缺少 summary 字段")
	}
	if payload.ToolCalls == nil {
		payload.ToolCalls = json.RawMessage("[]")
	}

	// Notes: 强制为 JSON object（jsonb 最稳妥）
	if len(payload.NotesRaw) == 0 || string(bytes.TrimSpace(payload.NotesRaw)) == "null" {
		// 为避免 DB NOT NULL 问题，使用空对象 {}
		payload.NotesRaw = json.RawMessage("{}")
	} else {
		b := bytes.TrimSpace(payload.NotesRaw)
		if len(b) == 0 || b[0] != '{' {
			return nil, errors.New("AI 返回的 notes 必须是 JSON 对象（object），请修正提示词或模型输出")
		}
		// 可选：进一步校验为有效 JSON
		var tmp map[string]any
		if err := json.Unmarshal(b, &tmp); err != nil {
			return nil, fmt.Errorf("notes 不是合法 JSON object: %w", err)
		}
	}

	return payload, nil
}

func sanitizeJSONBlock(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```JSON")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

// 在 persistConversationSummary 中直接使用 payload.NotesRaw 透传给 DB（jsonb）
func (h *Host) persistConversationSummary(ctx context.Context, conversationID string, payload *conversationSummaryPayload) error {
	logger.Infof("persistConversationSummary conversation_id=%s summary_len=%d related_summary_id=%s",
		conversationID, len(payload.Summary), payload.RelatedSummaryID)
	logger.Debugf("notes jsonb=%s", string(payload.NotesRaw))

	// 将标签转换为JSON
	tagsJSON, err := json.Marshal(payload.Tags)
	if err != nil {
		return fmt.Errorf("marshal tags failed: %w", err)
	}

	// 如果AI识别到相关的summary，则更新它
	if payload.RelatedSummaryID != "" && payload.RelatedSummaryID != "null" {
		logger.Infof("检测到相关summary，将更新: %s", payload.RelatedSummaryID)

		// 获取现有summary
		existingSummary, err := h.templateRepository.GetSummaryByID(ctx, payload.RelatedSummaryID)
		if err != nil {
			return fmt.Errorf("get existing summary failed: %w", err)
		}
		if existingSummary == nil {
			logger.Warnf("相关summary不存在，将创建新的: %s", payload.RelatedSummaryID)
			// 如果summary不存在，跳转到创建逻辑
		} else {
			// 更新现有summary
			existingSummary.SummaryText = payload.Summary
			existingSummary.Tags = string(tagsJSON)
			existingSummary.ToolCalls = string(payload.ToolCalls)
			existingSummary.Notes = string(payload.NotesRaw)
			// 不更新conversation_id，保持原有关联

			return h.templateRepository.UpdateSummary(ctx, existingSummary)
		}
	}

	// 创建新摘要记录
	summary := &model.Summaries{
		ConversationID: conversationID,
		SummaryText:    payload.Summary,
		Tags:           string(tagsJSON),
		ToolCalls:      string(payload.ToolCalls),
		Notes:          string(payload.NotesRaw),
	}

	return h.templateRepository.CreateSummary(ctx, summary)
}

// summarizeConversation 承载实际的总结流程，由 Host.SummarizeConversation 调用。
// summarizeConversation 返回 SummarizeResult 时直接返回 NotesJSON
func (h *Host) summarizeConversation(conversationID string, userID string) (*SummarizeResult, error) {
	logger.Infof("SummarizeConversation start conversation_id=%s user_id=%s", conversationID, userID)

	prompt, err := h.buildSummarizePrompt(h.ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}

	raw, err := h.invokeSummarizeModel(h.ctx, prompt)
	if err != nil {
		return nil, err
	}

	payload, err := parseConversationSummary(raw)
	if err != nil {
		return nil, err
	}

	if err := h.persistConversationSummary(h.ctx, conversationID, payload); err != nil {
		return nil, err
	}

	return &SummarizeResult{
		Summary:       payload.Summary,
		Tags:          payload.Tags,
		ToolCallsJSON: string(payload.ToolCalls),
		NotesJSON:     payload.NotesRaw,
		Notes:         jsonToStringMap(payload.NotesRaw),
	}, nil
}
