namespace go api
include "model.thrift"
include "openapi.thrift"
struct ChatRequest{
    1: string message(api.body="message", openapi.property='{
        title: "用户消息",
        description: "用户发送的消息内容",
        type: "string"
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

service ApiService {
    // 非流式对话
    ChatResponse Chat(1: ChatRequest req)(api.post="/api/v1/chat")
    // 流式对话
    ChatSSEHandlerResponse ChatSSE(1: ChatSSEHandlerRequest req)(api.get="/api/v1/chat/sse")
    // 示例接口 idl写好后运行make hertz-gen-api生成脚手架
    TemplateResponse Template(1: TemplateRequest req)(api.post="/api/v1/template")
}