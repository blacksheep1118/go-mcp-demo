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
	SumID         string
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
		Summary   string          `json:"summary"`
		Tags      []string        `json:"tags"`
		ToolCalls json.RawMessage `json:"tool_calls"`
		Notes     json.RawMessage `json:"notes"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Summary = aux.Summary
	p.Tags = aux.Tags
	p.ToolCalls = aux.ToolCalls
	p.NotesRaw = aux.Notes
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

	// 获取当前对话的现有summary（如果有的话）
	existingSummary, err := h.templateRepository.GetSummaryByConversationID(ctx, conversationID)
	if err != nil {
		logger.Warnf("get existing summary failed: %v", err)
		existingSummary = nil
	}

	// 构建现有summary的信息
	summaryInfo := h.buildExistingSummaryInfo(existingSummary)
	logger.Infof("buildSummarizePrompt: conversationID=%s, has existing summary=%v",
		conversationID, existingSummary != nil)

	templateData := map[string]any{
		"conversation_id":      conversationID,
		"conversation_history": strings.Join(history, "\n"),
		"existing_summary":     summaryInfo,
		"generated_at":         time.Now().Format(time.RFC3339),
	}

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
			openai.UserMessage("请严格输出 JSON，字段：summary、tags、tool_calls、notes。"),
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

// buildExistingSummaryInfo 构建现有summary的信息字符串
func (h *Host) buildExistingSummaryInfo(summary *model.Summaries) string {
	if summary == nil {
		return "本对话暂无已有总结"
	}

	var builder strings.Builder
	builder.WriteString("本对话已有总结：\n\n")

	// 解析tags
	var tags []string
	if err := json.Unmarshal([]byte(summary.Tags), &tags); err == nil {
		builder.WriteString(fmt.Sprintf("总结内容: %s\n", summary.SummaryText))
		builder.WriteString(fmt.Sprintf("标签: %v\n", tags))
		builder.WriteString(fmt.Sprintf("创建时间: %s\n", summary.CreatedAt.Format("2006-01-02 15:04:05")))
		builder.WriteString(fmt.Sprintf("更新时间: %s\n", summary.UpdatedAt.Format("2006-01-02 15:04:05")))
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

// persistConversationSummary 保存或更新对话总结
// 一个对话只对应一个总结，如果已存在则更新，否则创建
func (h *Host) persistConversationSummary(ctx context.Context, conversationID string, payload *conversationSummaryPayload) (string, error) {
	logger.Infof("persistConversationSummary conversation_id=%s summary_len=%d",
		conversationID, len(payload.Summary))
	logger.Debugf("notes jsonb=%s", string(payload.NotesRaw))

	// 将标签转换为JSON
	tagsJSON, err := json.Marshal(payload.Tags)
	if err != nil {
		return "", fmt.Errorf("marshal tags failed: %w", err)
	}

	// 查询该对话是否已有总结
	existingSummary, err := h.templateRepository.GetSummaryByConversationID(ctx, conversationID)
	if err != nil {
		return "", fmt.Errorf("get existing summary failed: %w", err)
	}

	if existingSummary != nil {
		// 已有总结，更新它
		logger.Infof("对话已有总结，将更新: summary_id=%s", existingSummary.ID)
		existingSummary.SummaryText = payload.Summary
		existingSummary.Tags = string(tagsJSON)
		existingSummary.ToolCalls = string(payload.ToolCalls)
		existingSummary.Notes = string(payload.NotesRaw)

		if err := h.templateRepository.UpdateSummary(ctx, existingSummary); err != nil {
			return "", err
		}
		return existingSummary.ID, nil
	}

	// 没有总结，创建新的
	logger.Infof("对话无总结，创建新的")
	summary := &model.Summaries{
		ConversationID: conversationID,
		SummaryText:    payload.Summary,
		Tags:           string(tagsJSON),
		ToolCalls:      string(payload.ToolCalls),
		Notes:          string(payload.NotesRaw),
	}

	if err := h.templateRepository.CreateSummary(ctx, summary); err != nil {
		return "", err
	}

	return summary.ID, nil
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

	// 解析 AI 返回结果
	payload, err := parseConversationSummary(raw)
	if err != nil {
		return nil, err
	}

	sumID, err := h.persistConversationSummary(h.ctx, conversationID, payload)
	if err != nil {
		return nil, err
	}

	return &SummarizeResult{
		SumID:         sumID,
		Summary:       payload.Summary,
		Tags:          payload.Tags,
		ToolCallsJSON: string(payload.ToolCalls),
		NotesJSON:     payload.NotesRaw,
		Notes:         jsonToStringMap(payload.NotesRaw),
	}, nil
}
