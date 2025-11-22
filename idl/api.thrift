namespace go api
include "model.thrift"
include "openapi.thrift"
struct ChatRequest{
    1: string message(api.body="message", openapi.property='{
        title: "用户消息",
        description: "用户发送的消息内容",
        type: "string"
    }')
    2: optional binary image(api.form="image", api.file_name="image", openapi.property='{
        title: "图片文件",
        description: "可选的图片文件，支持上传图片给AI分析",
        type: "string",
        format: "binary"
    }')
}(
    openapi.schema='{
        title: "聊天请求",
        description: "包含用户消息的聊天请求",
        required: ["message"]
    }'
)

struct ChatResponse{
    1: string response(api.body="response", openapi.property='{
        title: "AI回复",
        description: "AI生成的回复内容",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "聊天响应",
        description: "包含AI回复的聊天响应",
        required: ["response"]
    }'
)

struct ChatSSEHandlerRequest{
    1: string message(api.query="message",openapi.property='{
        title: "用户消息",
        description: "用户发送的消息内容",
        type: "string"
    }')
    2: optional binary image(api.form="image", api.file_name="image", openapi.property='{
        title: "图片文件",
        description: "可选的图片文件，支持上传图片给AI分析",
        type: "file",
    }')
}(
     openapi.schema='{
         title: "流式聊天请求",
         description: "包含用户消息的流式聊天请求",
         required: ["message"]
     }'
)

struct ChatSSEHandlerResponse{
    1: string response(api.body="response", openapi.property='{
        title: "AI回复片段",
        description: "AI生成的回复片段",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "流式聊天响应",
        description: "包含AI回复片段的流式聊天响应",
        required: ["response"]
    }'
)

struct TemplateRequest{
    1: string templateId(api.body="templateId", openapi.property='{
        title: "示范用param",
        description: "示范用param",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "示例请求",
        description: "示例请求",
        required: ["templateId"]
    }'
)

struct TemplateResponse{
    1: model.User user(api.body="user", openapi.property='{
        title: "示范用返回值",
        description: "示范用返回值",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "示例响应",
        description: "示例响应",
        required: ["user"]
    }'
)

struct SummarizeConversationRequest{
    1: string conversation_id(api.body="conversation_id", openapi.property='{
        title: "会话ID",
        description: "需要总结的会话ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "总结会话请求",
        description: "请求总结指定会话的内容",
        required: ["conversation_id"]
    }'
)

struct SummarizeConversationResponse{
    1: string summary(api.body="summary", openapi.property='{
        title: "会话总结",
        description: "会话内容的总结",
        type: "string"
    }')
    2: list<string> tags(api.body="tags", openapi.property='{
        title: "标签列表",
        description: "会话相关的标签",
        type: "array",
        items: {type: "string"}
    }')
    3: string tool_calls_json(api.body="tool_calls_json", openapi.property='{
        title: "工具调用JSON",
        description: "工具调用的JSON字符串",
        type: "string"
    }')
    4: map<string,string> notes(api.body="notes", openapi.property='{
        title: "笔记",
        description: "AI 或用户针对总结写的任意键值笔记，包含文件路径等信息",
        type: "object",
        additionalProperties: { "type": "string" }
    }')
}(
    openapi.schema='{
        title: "总结会话响应",
        description: "包含会话总结、标签、工具调用和笔记的响应",
        required: ["summary", "tags", "tool_calls_json", "notes"]
    }'
)

service ApiService {
    // 非流式对话
    ChatResponse Chat(1: ChatRequest req)(api.post="/api/v1/chat")
    // 流式对话
    ChatSSEHandlerResponse ChatSSE(1: ChatSSEHandlerRequest req)(api.post="/api/v1/chat/sse")
    // 示例接口 idl写好后运行make hertz-gen-api生成脚手架
    TemplateResponse Template(1: TemplateRequest req)(api.post="/api/v1/template")
    // 总结会话
    SummarizeConversationResponse SummarizeConversation(1: SummarizeConversationRequest req)(api.post="/api/v1/conversation/summarize")
}