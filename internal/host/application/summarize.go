package application

import (
	"regexp"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	api "github.com/FantasyRL/go-mcp-demo/api/model/api"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/internal/host/infra"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/cloudwego/hertz/pkg/app"
	consts "github.com/cloudwego/hertz/pkg/protocol/consts"
	openai "github.com/openai/openai-go/v2"
)

const (
	mockConversationID  = "00000000-0000-0000-0000-000000000001"
	summarizePromptName = "summarize"
)

var mockConversationHistory = []string{
	"[User] 今天的任务是为 go-mcp 项目增加对话总结能力，已经完成 IDL 更新。",
	"[Assistant] 已记录，后续需要补上提示词和数据库落盘。",
	"[Tool] file.write -> internal/host/infra/prompts/summarize.txt",
	"[User] 记得把 tags、tool_calls、notes 都写入 summaries 表。",
}

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

func (h *Host) buildSummarizePrompt(conversationID string) (string, error) {
	tpl, err := infra.LoadPrompt(summarizePromptName)
	if err != nil {
		return "", fmt.Errorf("load summarize prompt: %w", err)
	}

	history := selectConversationHistory(conversationID)
	rendered, err := infra.RenderPrompt(tpl, map[string]any{
		"conversation_id":      conversationID,
		"conversation_history": strings.Join(history, "\n"),
		"generated_at":         time.Now().Format(time.RFC3339),
	})
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

func selectConversationHistory(conversationID string) []string {
	if conversationID == mockConversationID {
		return mockConversationHistory
	}
	return []string{
		fmt.Sprintf("[System] conversation_id=%s 未命中 mock 数据，返回空历史。", conversationID),
	}
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
	logger.Infof("persistConversationSummary conversation_id=%s summary_len=%d", conversationID, len(payload.Summary))
	logger.Debugf("notes jsonb=%s", string(payload.NotesRaw))
	// TODO: 将 payload.NotesRaw 作为 datatypes.JSON 或 json.RawMessage 写入 summaries 表
	return nil
}

// summarizeConversation 承载实际的总结流程，由 Host.SummarizeConversation 调用。
// summarizeConversation 返回 SummarizeResult 时直接返回 NotesJSON
func (h *Host) summarizeConversation(conversationID string) (*SummarizeResult, error) {
	logger.Infof("SummarizeConversation start conversation_id=%s", conversationID)

	prompt, err := h.buildSummarizePrompt(conversationID)
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
