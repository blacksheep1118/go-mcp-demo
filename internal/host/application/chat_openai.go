package application

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
	"github.com/bytedance/sonic"

	"github.com/FantasyRL/go-mcp-demo/pkg/logger"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	openai "github.com/openai/openai-go/v2"
)

// systemPrompt 系统提示词，用于指导 AI 处理课表查询等任务
var systemPrompt = `你是一个智能助手，需要帮助用户提供回答，当使用到福州大学教务处相关mcp工具时，请务必遵守以下规则和说明，确保输出的信息准确无误。
## 1. 身份验证与 MCP 工具使用
- 当用户需要访问福州大学教务处相关功能（如查询课表、成绩等）时，须先调用 Login 工具获取用户的身份信息（id 和 cookie），调用 Login 时，学号和密码参数可以随便传递，系统会返回用户id与cookie，然后把这些拿去调用CheckSession检查会话是否过期
- 如果 CheckSession 调用失败，说明用户未登录或登录已过期，请友好地告知用户之后再进行一次请求(因为前端会自动checkSession)
- 获取到登录信息后，才能调用其他教务处相关的 MCP 工具
- 不要让用户感知到这些调用，只要CheckSession没问题id与cookie就能一直用

## 2. 学期代码规则（重要！）
学期代码格式为 6 位数字 YYYYSS，规则如下：
- 202402 → 2024-2025学年第二学期（2025年春季学期，2025年2-6月）
- 202501 → 2025-2026学年第一学期（2025年秋季学期，2025年9月-2026年1月）
- 202502 → 2025-2026学年第二学期（2026年春季学期，2026年2-6月）
- 202601 → 2026-2027学年第一学期（2026年秋季学期）

规律总结：
- 后两位为 01 → 秋季学期（该年9月开始）
- 后两位为 02 → 春季学期（次年2月开始）
- 当前时间在2025年12月，当前学期是 202501（2025年秋季学期）

## 3. 课程节次与时间对应关系（必须严格遵守）
每天分为以下时间段，对应的节次（startClass - endClass）：
- 第 1-2 节：08:20 - 10:00（上午第一大节）
- 第 3-4 节：10:20 - 12:00（上午第二大节）
- 第 5-6 节：14:00 - 15:40（下午第一大节）
- 第 7-8 节：15:50 - 17:30（下午第二大节）
- 第 9-11 节：19:00 - 21:35（晚上，3节连上）

注意：部分课程可能跨越多个节次，如 5-8 节表示从14:00持续到17:30，正常一个课程都会有两个节次

## 4. 周次（week）与单双周规则
课程的 scheduleRules 包含以下字段：
- startWeek：开始周次（如 1 表示第1周）
- endWeek：结束周次（如 16 表示第16周）
- weekday：星期几（1=周一，2=周二，...，7=周日）
- single：是否单周上课（true=单周有课）
- double：是否双周上课（true=双周有课）
- adjust：是否为调课（true=临时调整的课程）

判断课程是否在本周：
1. 首先确定当前是第几周（需要根据学期开始时间计算，通常第1周从9月初开始）
2. 检查当前周次是否在 [startWeek, endWeek] 范围内
3. 检查单双周：
   - 如果 single=true, double=true：每周都上
   - 如果 single=true, double=false：仅单周（1,3,5,7...）上课
   - 如果 single=false, double=true：仅双周（2,4,6,8...）上课
4. 如果 adjust=true，这是调课安排，需特别注意 rawAdjust 字段的说明

## 5. 输出课表的格式要求
当用户查询课表时，你应该：
1. **按时间顺序组织**：先按星期（周一到周日），再按节次（1-2节 → 3-4节 → ...）排序
2. **清晰的时间标注**：必须同时显示节次和具体时间，如"第3-4节（10:20-12:00）"
3. **地点信息完整**：显示完整的上课地点，如"旗山东3-307"
4. **单双周标记清楚**：
   - 如果是单周课程，标注"（单周）"
   - 如果是双周课程，标注"（双周）"
   - 如果每周都上，不需要标注
5. **过滤非本周课程**：
   - 如果用户查询"本周课表"或"今天/明天的课"，必须过滤掉不在本周上课的课程
   - 如果课程周次范围不包含当前周，不要显示
   - 注意单双周过滤
6. **格式示例**：
   周一：
   - 10:20-12:00 计算机操作系统（陈勃）@ 旗山东3-307
   - 15:50-17:30 人工智能（杨文杰）@ 旗山东3-307【第9周开始】
   
   周二：
   - 10:20-12:00 数据库系统原理（程烨）@ 旗山东2-209
   - 19:00-21:35 现代搜索引擎技术及应用（廖祥文）@ 旗山东3-405

## 6. 特殊情况处理
- 如果课程的 scheduleRules 为空或 null（如在线课程"智慧树：视觉与艺术"），说明该课程无固定上课时间，需要告知用户这是网络课程
- 如果 remark 字段有内容，重要的备注信息应该告知用户
- 如果有 rawAdjust 字段内容，说明有调课安排，务必提醒用户注意

## 7. 用户查询意图识别
- "今天有什么课"：查询当天（根据 weekday）的课程
- "明天有课吗"：查询明天的课程
- "本周课表"：显示本周一到周日的所有课程
- "下周一有什么课"：需要计算下周的周次，然后查询
- "我的课表"：显示完整的学期课表（不过滤周次）

记住：准确性最重要！务必严格按照 scheduleRules 的数据来判断课程时间，不要臆测或编造信息。`

// 将 OpenAI 的 tool_calls[].function.arguments (string) 解成 map[string]any
func parseOpenAIToolArgs(argStr string) map[string]any {
	if argStr == "" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(argStr), &m); err == nil {
		return m
	}
	// 如果不是 JSON，就当成纯字符串包裹
	return map[string]any{"_": argStr}
}

const maxToolRounds = 10 // 防御性上限，避免死循环

func (h *Host) StreamChatOpenAI(
	ctx context.Context,
	userID string,
	conversationID string,
	userMsg string,
	imageData []byte,
	emit func(event string, v any) error, // SSE: event 名 + 任意 JSON 数据
) error {
	// 历史（OpenAI）
	var hist []openai.ChatCompletionMessageParamUnion

	// 从数据库加载该对话的历史消息填充 hist（如果不存在就保持为空）
	conversation, err := h.templateRepository.GetConversationByID(ctx, conversationID)
	if err != nil {
		return err
	}
	if conversation != nil {
		if err := json.Unmarshal([]byte(conversation.Messages), &hist); err != nil {
			logger.Errorf("failed to unmarshal conversation messages, conversationID=%s, err=%v", conversationID, err)
			return err
		}
	} else {
		// 新对话，添加系统提示词
		hist = append(hist, openai.SystemMessage(systemPrompt))
	}

	// 记录当前历史长度，用于之后只持久化“新增部分”
	baseLen := len(hist)

	// 构建用户消息
	if len(imageData) > 0 {
		// 如果有图片，使用多模态消息格式
		base64Image := base64.StdEncoding.EncodeToString(imageData)

		// 创建文本部分
		textPart := openai.ChatCompletionContentPartUnionParam{
			OfText: &openai.ChatCompletionContentPartTextParam{
				Type: "text",
				Text: userMsg,
			},
		}

		// 创建图片部分
		imagePart := openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: "data:image/jpeg;base64," + base64Image,
		})

		// 创建用户消息
		hist = append(hist, openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Role: "user",
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
						textPart,
						imagePart,
					},
				},
			},
		})
	} else {
		// 纯文本消息
		hist = append(hist, openai.UserMessage(userMsg))
	}

	// 工具（OpenAI 版）
	tools := h.mcpCli.ConvertToolsToOpenAI()

	round := 0
	for {
		round++
		if round > maxToolRounds {
			// 每轮对话结束时持久化“新增历史”
			newMessages := hist[baseLen:]
			if len(newMessages) > 0 {
				if err := h.templateRepository.UpsertConversation(ctx, userID, conversationID, newMessages); err != nil {
					return err
				}
			}

			_ = emit(constant.SSEEventDone, map[string]any{"reason": "tool_round_limit"})
			return nil
		}

		// 一轮生成：边流边推，若需要工具则中断本轮
		var assistantBuf string
		var acc openai.ChatCompletionAccumulator
		var needTools bool

		params := openai.ChatCompletionNewParams{
			Model:    openai.ChatModel(config.AiProvider.Model),
			Messages: hist,
		}
		// 只有在 tools 非空时才传递 Tools 参数，避免阿里云 API 报错
		if len(tools) > 0 {
			params.Tools = tools
		}

		err := h.aiProviderCli.ChatStreamOpenAI(ctx, params, func(chunk *openai.ChatCompletionChunk) error {
			acc.AddChunk(*chunk)
			if len(chunk.Choices) > 0 {
				if s := chunk.Choices[0].Delta.Content; s != "" {
					assistantBuf += s
					_ = emit(constant.SSEEventDelta, map[string]any{"text": s})
				}
				// 工具调用结束标志（OpenAI：最后一帧 finish_reason = "tool_calls"）
				if chunk.Choices[0].FinishReason == "tool_calls" {
					needTools = true
					if len(acc.Choices) > 0 && len(acc.Choices[0].Message.ToolCalls) > 0 {
						_ = emit(constant.SSEEventStartToolCall, map[string]any{
							"tool_calls": acc.Choices[0].Message.ToolCalls,
							"round":      round,
						})
					}
					return errno.OllamaInternalStopStream
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		// 把已产生的 assistant 文本落历史
		if assistantBuf != "" && !needTools {
			hist = append(hist, openai.AssistantMessage(assistantBuf))
		}

		// 如果本轮不需要工具，说明模型已经给出最终答案
		if !needTools {
			// 对话结束，持久化“新增历史”
			newMessages := hist[baseLen:]
			if len(newMessages) > 0 {
				if err := h.templateRepository.UpsertConversation(ctx, userID, conversationID, newMessages); err != nil {
					return err
				}
			}

			_ = emit(constant.SSEEventDone, map[string]any{"reason": "completed"})
			return nil
		}

		// 执行（可能多个）工具调用，然后将每个工具结果以 ToolMessage 落历史
		if len(acc.Choices) == 0 || len(acc.Choices[0].Message.ToolCalls) == 0 {
			// 偶发兜底：标记需要工具但没聚合到（理论上不会发生）
			// 对话结束，持久化“新增历史”
			newMessages := hist[baseLen:]
			if len(newMessages) > 0 {
				if err := h.templateRepository.UpsertConversation(ctx, userID, conversationID, newMessages); err != nil {
					return err
				}
			}

			_ = emit(constant.SSEEventDone, map[string]any{"reason": "no_tool_details"})
			return nil
		}

		toolCallsParam := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(acc.Choices[0].Message.ToolCalls))
		for _, tc := range acc.Choices[0].Message.ToolCalls {
			toolCallsParam = append(toolCallsParam, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID:   tc.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments, // 注意：这里是字符串
					},
				},
			})
		}
		// 根据openAI规范，tool_call前需要一条assistantMsg
		assistantWithCalls := openai.ChatCompletionAssistantMessageParam{
			Role:      "assistant",
			ToolCalls: toolCallsParam,
		}
		hist = append(hist, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantWithCalls})

		for _, tc := range acc.Choices[0].Message.ToolCalls {
			name := tc.Function.Name

			// OpenAI 的 arguments 是字符串，需要解成 map[string]any
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = map[string]any{"_parse_error": err.Error(), "_raw": tc.Function.Arguments}
			}

			_ = emit(constant.SSEEventToolCall, map[string]any{
				"round": round,
				"name":  name,
				"args":  args,
			})

			var out string
			if name == "login" {
				loginData, ok := utils.ExtractLoginData(h.ctx)
				if !ok {
					out = "tool error: no login data in context"
				} else {
					out, _ = sonic.MarshalString(*loginData)
				}
			} else {
				toolRes, callErr := h.mcpCli.CallTool(ctx, name, args)
				if callErr != nil {
					out = "tool error: " + callErr.Error()
				} else {
					out = toolRes
				}
			}
			_ = emit(constant.SSEEventToolResult, map[string]any{
				"round":  round,
				"name":   name,
				"result": out,
			})

			// 工具结果回模型（重要）：OpenAI 规范用 ToolMessage，必须带 tool_call_id
			hist = append(hist, openai.ToolMessage(out, tc.ID))
			//logger.Infof("[tool round %d] %s executed", round, name)
		}

		// 循环进入下一轮：模型会在新的上下文（含工具结果）上继续生成
	}
}

// ChatOpenAI 非流式OpenAI聊天，支持图片和工具调用
func (h *Host) ChatOpenAI(
	userID string,
	conversationID string, // uuid
	msg string,
	imageData []byte,
) (string, error) {
	// 历史（OpenAI）
	var hist []openai.ChatCompletionMessageParamUnion

	// 从数据库加载该对话的历史消息填充 hist（如果不存在就保持为空）
	conversation, err := h.templateRepository.GetConversationByID(h.ctx, conversationID)
	if err != nil {
		return "", err
	}
	if conversation != nil {
		if err := json.Unmarshal([]byte(conversation.Messages), &hist); err != nil {
			logger.Errorf("failed to unmarshal conversation messages, conversationID=%s, err=%v", conversationID, err)
			return "", err
		}
	}

	// 记录当前历史长度，用于之后只持久化“新增部分”
	baseLen := len(hist)

	// 构建用户消息
	if len(imageData) > 0 {
		// 如果有图片，使用多模态消息格式
		base64Image := base64.StdEncoding.EncodeToString(imageData)

		// 创建文本部分
		textPart := openai.ChatCompletionContentPartUnionParam{
			OfText: &openai.ChatCompletionContentPartTextParam{
				Type: "text",
				Text: msg,
			},
		}

		// 创建图片部分
		imagePart := openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: "data:image/jpeg;base64," + base64Image,
		})

		// 创建用户消息
		hist = append(hist, openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Role: "user",
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
						textPart,
						imagePart,
					},
				},
			},
		})
	} else {
		// 纯文本消息
		hist = append(hist, openai.UserMessage(msg))
	}

	// 工具（OpenAI 版）- 如果有图片则不使用工具（vision模型可能不支持）
	var tools []openai.ChatCompletionToolUnionParam
	if len(imageData) == 0 {
		tools = h.mcpCli.ConvertToolsToOpenAI()
	}

	round := 0
	for {
		round++
		if round > maxToolRounds {
			// 对话结束，持久化“新增历史”
			newMessages := hist[baseLen:]
			if len(newMessages) > 0 {
				if err := h.templateRepository.UpsertConversation(h.ctx, userID, conversationID, newMessages); err != nil {
					return "", err
				}
			}

			return "已达到工具调用轮次上限", nil
		}

		// 调用OpenAI API
		params := openai.ChatCompletionNewParams{
			Model:    openai.ChatModel(config.AiProvider.Model),
			Messages: hist,
			Tools:    tools,
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

		resp, err := h.aiProviderCli.ChatOpenAI(h.ctx, params)
		if err != nil {
			logger.Errorf("ChatOpenAI API error: %v", err)
			return "", err
		}

		logger.Infof("ChatOpenAI response: choices=%d", len(resp.Choices))
		if len(resp.Choices) == 0 {
			logger.Errorf("ChatOpenAI: no choices in response")

			// 对话结束，持久化“新增历史”（虽然没有新 assistant 内容，但有这轮 user 消息）
			newMessages := hist[baseLen:]
			if len(newMessages) > 0 {
				if err := h.templateRepository.UpsertConversation(h.ctx, userID, conversationID, newMessages); err != nil {
					return "", err
				}
			}

			return "模型返回为空", nil
		}

		logger.Infof("ChatOpenAI finish_reason=%s, content=%s, tool_calls=%d",
			resp.Choices[0].FinishReason, resp.Choices[0].Message.Content, len(resp.Choices[0].Message.ToolCalls))

		// 检查是否需要工具调用
		if resp.Choices[0].FinishReason != "tool_calls" || len(resp.Choices[0].Message.ToolCalls) == 0 {
			// 无工具调用，返回模型回复
			content := resp.Choices[0].Message.Content
			hist = append(hist, openai.AssistantMessage(content))

			// 对话结束，持久化“新增历史”
			newMessages := hist[baseLen:]
			if len(newMessages) > 0 {
				if err := h.templateRepository.UpsertConversation(h.ctx, userID, conversationID, newMessages); err != nil {
					return "", err
				}
			}

			return content, nil
		}

		// 有工具调用，处理工具
		toolCallsParam := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(resp.Choices[0].Message.ToolCalls))
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			toolCallsParam = append(toolCallsParam, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID:   tc.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				},
			})
		}

		// 添加assistant消息（带工具调用）
		assistantWithCalls := openai.ChatCompletionAssistantMessageParam{
			Role:      "assistant",
			ToolCalls: toolCallsParam,
		}
		hist = append(hist, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantWithCalls})

		// 执行所有工具调用
		for _, tc := range resp.Choices[0].Message.ToolCalls {
			name := tc.Function.Name

			// OpenAI 的 arguments 是字符串，需要解成 map[string]any
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = map[string]any{"_parse_error": err.Error(), "_raw": tc.Function.Arguments}
			}

			out, callErr := h.mcpCli.CallTool(h.ctx, name, args)
			if callErr != nil {
				out = "tool error: " + callErr.Error()
			}

			// 工具结果回模型（重要）：OpenAI 规范用 ToolMessage，必须带 tool_call_id
			hist = append(hist, openai.ToolMessage(out, tc.ID))
		}

		// 循环进入下一轮：模型会在新的上下文（含工具结果）上继续生成
	}
}
