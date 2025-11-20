package application

import (
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
    "[User] 记得把 tags、tool_calls、file_paths 都写入 summaries 表。",
}

type SummarizeResult struct {
    Summary       string
    Tags          []string
    ToolCallsJSON string
    FilePaths     []string
}

type conversationSummaryPayload struct {
    Summary   string          `json:"summary"`
    Tags      []string        `json:"tags"`
    ToolCalls json.RawMessage `json:"tool_calls"`
    FilePaths []string        `json:"file_paths"`
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
            openai.UserMessage("请严格输出 JSON，字段：summary、tags、tool_calls、file_paths。"),
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

func (h *Host) persistConversationSummary(ctx context.Context, conversationID string, payload *conversationSummaryPayload) error {
    // TODO: 使用 summaries 查询对象完成 upsert（计划第 5 步会接入实际存储）
    logger.Infof("persistConversationSummary conversation_id=%s summary_len=%d", conversationID, len(payload.Summary))
    return nil
}

// SummarizeConversation .
// @router /api/v1/conversation/summarize [POST]
func SummarizeConversation(ctx context.Context, c *app.RequestContext) {
	var err error
	var req api.SummarizeConversationRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(api.SummarizeConversationResponse)

	c.JSON(consts.StatusOK, resp)
}

// summarizeConversation 承载实际的总结流程，由 Host.SummarizeConversation 调用。
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
        FilePaths:     payload.FilePaths,
    }, nil
}

