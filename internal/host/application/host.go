package application

import (
	"context"
	"errors"

	"github.com/FantasyRL/go-mcp-demo/internal/host/infra"
	"github.com/FantasyRL/go-mcp-demo/internal/host/repository"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/db"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_client"
	"github.com/FantasyRL/go-mcp-demo/pkg/gorm-gen/query"
	"github.com/openai/openai-go/v2"
)

// 简单的内存存储用户对话历史
var history = make(map[int64][]ai_provider.Message)
var historyOpenAI = make(map[int64][]openai.ChatCompletionMessageParamUnion)

type Host struct {
	ctx           context.Context
	mcpCli        mcp_client.ToolClient
	aiProviderCli *ai_provider.Client
	// 添加需要的连接
	templateRepository repository.TemplateRepository
}

func NewHost(ctx context.Context, clientSet *base.ClientSet) *Host {
	return &Host{
		ctx:                ctx,
		mcpCli:             clientSet.MCPCli,
		aiProviderCli:      clientSet.AiProviderCli,
		templateRepository: infra.NewTemplateRepository(db.NewDBWithQuery(clientSet.ActualDB, query.Use), clientSet.Cache),
	}
}

// SummarizeConversation 暴露给 Handler/Service 的入口，负责做一些入参校验并
// 委托给 summarize.go 中的核心编排逻辑，保持 Host 结构的职责清晰。
func (h *Host) SummarizeConversation(conversationID string, userID string) (*SummarizeResult, error) {
	if h == nil {
		return nil, errors.New("host is nil")
	}
	if conversationID == "" {
		return nil, errors.New("conversation_id is required")
	}
	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	// 将真正的总结流程留在 summarize.go，便于测试与复用。
	return h.summarizeConversation(conversationID, userID)
}
