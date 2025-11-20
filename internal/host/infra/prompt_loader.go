package infra

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
)
const (
	// promptsDir 提示词模板目录
	promptsDir = "internal/host/infra/prompts"
)
// PromptRepository 负责加载和渲染提示词模板（如 summarize.txt）
type PromptRepository struct {
	baseDir string // 模板根目录，例如 "internal/host/infra"
}

// NewPromptRepository 创建提示词仓库
func NewPromptRepository(baseDir string) *PromptRepository {
	return &PromptRepository{baseDir: baseDir}
}

// Render 渲染指定模板文件（相对 baseDir），使用 data 作为模板变量
func (r *PromptRepository) Render(ctx context.Context, tmplPath string, data any) (string, error) {
	if tmplPath == "" {
		return "", fmt.Errorf("template path is required")
	}

	// 支持 ctx 预留扩展（例如后面做超时/日志等）
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	path := tmplPath
	if r.baseDir != "" && !filepath.IsAbs(tmplPath) {
		path = filepath.Join(r.baseDir, tmplPath)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read template %s: %w", path, err)
	}

	tmpl, err := template.New(filepath.Base(path)).
		Option("missingkey=error").
		Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", path, err)
	}

	if data == nil {
		data = struct{}{}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", path, err)
	}

	return buf.String(), nil
}

// 也可以加一个针对 summarize 的封装，方便上层调用
const SummarizePromptPath = "prompts/summarize.txt"

func (r *PromptRepository) RenderSummarizePrompt(ctx context.Context, data any) (string, error) {
	return r.Render(ctx, SummarizePromptPath, data)
}

// LoadPrompt 从 prompts 目录读取模板文本
// name 参数对应文件名（不含扩展名），如 "summarize" 对应 "summarize.txt"
func LoadPrompt(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("prompt name is required")
	}

	// 构建文件路径：name -> prompts/name.txt
	filename := name + ".txt"
	filePath := filepath.Join(promptsDir, filename)

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		logger.Errorf("failed to load prompt template %s: %v", filePath, err)
		return "", fmt.Errorf("load prompt template %s: %w", name, err)
	}

	logger.Infof("loaded prompt template: %s", filePath)
	return string(content), nil
}

// RenderPrompt 使用 text/template 渲染模板，替换变量占位符
// tpl 是模板文本，data 是用于填充模板的数据（可以是 map、struct 等）
func RenderPrompt(tpl string, data any) (string, error) {
	if tpl == "" {
		return "", fmt.Errorf("template is empty")
	}

	// 创建模板并设置选项：缺失变量时报错
	tmpl, err := template.New("prompt").
		Option("missingkey=error").
		Parse(tpl)
	if err != nil {
		logger.Errorf("failed to parse prompt template: %v", err)
		return "", fmt.Errorf("parse template: %w", err)
	}

	// 如果 data 为 nil，使用空结构体
	if data == nil {
		data = struct{}{}
	}

	// 执行模板渲染
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		logger.Errorf("failed to render prompt template: %v", err)
		return "", fmt.Errorf("render template: %w", err)
	}

	result := buf.String()
	logger.Infof("rendered prompt template successfully, length: %d", len(result))
	return result, nil
}

// LoadAndRenderPrompt 组合操作：加载并渲染提示词模板
// 这是一个便捷函数，将 LoadPrompt 和 RenderPrompt 组合在一起
func LoadAndRenderPrompt(name string, data any) (string, error) {
	tpl, err := LoadPrompt(name)
	if err != nil {
		return "", err
	}
	return RenderPrompt(tpl, data)
}