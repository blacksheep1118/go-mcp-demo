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
        title:"图片文件",
        description:"可选的图片文件，支持上传图片给AI分析",
        type:"string",
        format:"binary"
    }')
    3: string conversation_id(api.body="conversation_id", openapi.property='{
        title:"对话ID",
        description:"前端生成的UUID，多轮会话唯一标识",
        type:"string"
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
        title:"AI回复",
        description:"AI生成的回复内容",
        type:"string"
    }')
    2: optional string conversation_id(api.body="conversation_id", openapi.property='{
        title:"对话ID",
        description:"回显本轮所属的对话UUID",
        type:"string"
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
        title:"图片文件",
        description:"可选的图片文件，支持上传图片给AI分析",
        type:"file"
    }')
    3: string conversation_id(api.query="conversation_id", openapi.property='{
        title:"对话ID",
        description:"前端生成的UUID，多轮会话标识",
        type:"string"
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
        title:"AI回复片段",
        description:"AI生成的回复片段",
        type:"string"
    }')
    2: optional string conversation_id(api.body="conversation_id", openapi.property='{
        title:"对话ID UUID",
        type:"string"
    }')
}(
    openapi.schema='{
        title:"流式聊天响应",
        description:"包含AI回复片段的流式聊天响应",
        required:["response"]
    }'
)

struct GetConversationHistoryRequest {
    1: string conversation_id(api.query="conversation_id", openapi.property='{
        title:"对话ID",
        description:"要获取的对话UUID",
        type:"string"
    }')
}(
    openapi.schema='{
        title:"获取历史请求",
        description:"按UUID获取完整对话历史",
        required:["conversation_id"]
    }'
)

struct GetConversationHistoryResponse {
    1: string conversation_id(api.body="conversation_id", openapi.property='{
        title:"对话ID",
        type:"string"
    }')
    2: string messages(api.body="messages", openapi.property='{
        title:"json消息",
        type:"strig"
    }')
}(
    openapi.schema='{
        title:"获取历史响应",
        description:"返回对话的全部消息"
        required:["conversation_id","messages"]
    }'
)

struct ConversationItem {
    1: string id(api.body="id", openapi.property='{
        title: "对话ID",
        type: "string"
    }')
    2: string title(api.body="title", openapi.property='{
        title: "对话标题",
        type: "string"
    }')
    3: i64 created_at(api.body="created_at", openapi.property='{
        title: "创建时间",
        type: "integer",
        format: "int64"
    }')
    4: i64 updated_at(api.body="updated_at", openapi.property='{
        title: "更新时间",
        type: "integer",
        format: "int64"
    }')
}(
    openapi.schema='{
        title: "对话信息",
        description: "对话的基本信息"
    }'
)

struct ListConversationsRequest {
}(
    openapi.schema='{
        title: "获取对话列表请求",
        description: "获取用户的所有对话列表"
    }'
)

struct ListConversationsResponse {
    1: list<ConversationItem> conversations(api.body="conversations", openapi.property='{
        title: "对话列表",
        type: "array"
    }')
}(
    openapi.schema='{
        title: "对话列表响应",
        description: "返回对话列表",
        required: ["conversations"]
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
        type: "array"
    }')
    3: string tool_calls_json(api.body="tool_calls_json", openapi.property='{
        title: "工具调用JSON",
        description: "工具调用的JSON字符串",
        type: "string"
    }')
    4: map<string,string> notes(api.body="notes", openapi.property='{
        title: "笔记",
        description: "AI 或用户针对总结写的任意键值笔记，包含文件路径等信息",
        type: "object"
    }')
}(
    openapi.schema='{
        title: "总结会话响应",
        description: "包含会话总结、标签、工具调用和笔记的响应",
        required: ["summary", "tags", "tool_calls_json", "notes"]
    }'
)

struct GetLoginDataRequest{
    1: string stu_id(api.body="stu_id", openapi.property='{
        title: "学号",
        description: "福州大学学号",
        type: "string"
    }')
    2: string password(api.body="password", openapi.property='{
        title: "密码",
        description: "用户的登录密码",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "登录请求",
        description: "包含学号和密码的登录请求",
        required: ["stu_id", "password"]
    }'
)

struct GetLoginDataResponse{
    1: string identifier(api.body="identifier", openapi.property='{
        title: "用户ID",
        description: "登录成功后返回的用户唯一标识符",
        type: "string"
    }')
    2: string cookie(api.body="cookie", openapi.property='{
        title: "会话Cookie",
        description: "登录成功后返回的会话Cookie",
        type: "string"
    }')
    3: string access_token(api.body="access_token", openapi.property='{
        title: "访问令牌",
        description: "登录成功后返回的访问令牌",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "登录响应",
        description: "包含用户ID和会话Cookie的登录响应",
        required: ["identifier", "cookie","access_token"]
    }'
)

struct GetUserInfoRequest {
}(
    openapi.schema='{
        title: "用户信息请求",
        description: "请求用户的基本信息"
    }'
)

struct GetUserInfoResponse {
    1: string user_id(api.body="user_id", openapi.property='{
        title: "用户ID",
        description: "用户的唯一标识符",
        type: "string"
    }')
    2: string username(api.body="username", openapi.property='{
        title: "用户名",
        description: "用户的登录名",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "用户信息响应",
        description: "包含用户ID和用户名的响应",
        required: ["user_id", "username"]
    }'
)

struct CreateTodoRequest {
    1: string title(api.body="title", openapi.property='{
        title: "待办事项标题",
        description: "待办事项的标题",
        type: "string"
    }')
    2: string content(api.body="content", openapi.property='{
        title: "待办事项内容",
        description: "待办事项的详细内容",
        type: "string"
    }')
    3: i64 start_time(api.body="start_time", openapi.property='{
        title: "开始时间",
        description: "待办事项开始时间，unix毫秒时间戳",
        type: "integer",
        format: "int64"
    }')
    4: i64 end_time(api.body="end_time", openapi.property='{
        title: "结束时间",
        description: "待办事项结束时间，unix毫秒时间戳",
        type: "integer",
        format: "int64"
    }')
    5: optional i16 is_all_day(api.body="is_all_day", openapi.property='{
        title: "是否全天",
        description: "是否为全天事项，0-否，1-是",
        type: "integer"
    }')
    6: i16 priority(api.body="priority", openapi.property='{
        title: "优先级",
        description: "1-紧急且重要，2-重要不紧急，3-紧急不重要，4-不重要不紧急",
        type: "integer"
    }')
    7: optional i64 remind_at(api.body="remind_at", openapi.property='{
        title: "提醒时间",
        description: "待办事项提醒时间，unix毫秒时间戳",
        type: "integer",
        format: "int64"
    }')
    8: optional string category(api.body="category", openapi.property='{
        title: "分类",
        description: "待办事项分类",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "创建待办事项请求",
        description: "创建待办事项的请求参数",
        required: ["title", "content", "start_time", "end_time", "priority"]
    }'
)

struct CreateTodoResponse {
    1: string id(api.body="id", openapi.property='{
        title: "待办事项ID",
        description: "创建成功的待办事项ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "创建待办事项响应",
        description: "返回创建成功的待办事项ID",
        required: ["id"]
    }'
)

struct GetTodoRequest {
    1: string id(api.query="id", openapi.property='{
        title: "待办事项ID",
        description: "要查询的待办事项ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "获取待办事项请求",
        description: "根据ID获取待办事项",
        required: ["id"]
    }'
)

struct TodoItem {
    1: string id(api.body="id", openapi.property='{
        title: "待办事项ID",
        type: "string"
    }')
    2: string title(api.body="title", openapi.property='{
        title: "标题",
        type: "string"
    }')
    3: string content(api.body="content", openapi.property='{
        title: "内容",
        type: "string"
    }')
    4: i64 start_time(api.body="start_time", openapi.property='{
        title: "开始时间",
        type: "integer",
        format: "int64"
    }')
    5: i64 end_time(api.body="end_time", openapi.property='{
        title: "结束时间",
        type: "integer",
        format: "int64"
    }')
    6: i16 is_all_day(api.body="is_all_day", openapi.property='{
        title: "是否全天",
        type: "integer"
    }')
    7: i16 status(api.body="status", openapi.property='{
        title: "状态",
        description: "0-未完成，1-已完成",
        type: "integer"
    }')
    8: i16 priority(api.body="priority", openapi.property='{
        title: "优先级",
        description: "1-紧急且重要，2-重要不紧急，3-紧急不重要，4-不重要不紧急",
        type: "integer"
    }')
    9: optional i64 remind_at(api.body="remind_at", openapi.property='{
        title: "提醒时间",
        type: "integer",
        format: "int64"
    }')
    10: optional string category(api.body="category", openapi.property='{
        title: "分类",
        type: "string"
    }')
    11: i64 created_at(api.body="created_at", openapi.property='{
        title: "创建时间",
        type: "integer",
        format: "int64"
    }')
    12: i64 updated_at(api.body="updated_at", openapi.property='{
        title: "更新时间",
        type: "integer",
        format: "int64"
    }')
}(
    openapi.schema='{
        title: "待办事项",
        description: "待办事项详细信息"
    }'
)

struct GetTodoResponse {
    1: TodoItem todo(api.body="todo", openapi.property='{
        title: "待办事项",
        type: "object"
    }')
}(
    openapi.schema='{
        title: "获取待办事项响应",
        description: "返回待办事项详细信息",
        required: ["todo"]
    }'
)

struct ListTodoRequest {
}(
    openapi.schema='{
        title: "待办事项列表请求",
        description: "获取用户的所有待办事项"
    }'
)

struct ListTodoResponse {
    1: list<TodoItem> todos(api.body="todos", openapi.property='{
        title: "待办事项列表",
        type: "array"
    }')
}(
    openapi.schema='{
        title: "待办事项列表响应",
        description: "返回待办事项列表",
        required: ["todos"]
    }'
)

struct SearchTodoRequest {
    1: optional i16 status(api.query="status", openapi.property='{
        title: "状态筛选",
        description: "0-未完成，1-已完成，不传则返回全部",
        type: "integer"
    }')
    2: optional i16 priority(api.query="priority", openapi.property='{
        title: "优先级筛选",
        description: "1-紧急且重要，2-重要不紧急，3-紧急不重要，4-不重要不紧急",
        type: "integer"
    }')
    3: optional string category(api.query="category", openapi.property='{
        title: "分类筛选",
        description: "按分类筛选待办事项",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "搜索待办事项请求",
        description: "根据条件搜索待办事项，支持按状态、优先级和分类筛选"
    }'
)

struct SearchTodoResponse {
    1: list<TodoItem> todos(api.body="todos", openapi.property='{
        title: "待办事项列表",
        type: "array"
    }')
}(
    openapi.schema='{
        title: "待办事项列表响应",
        description: "返回待办事项列表",
        required: ["todos"]
    }'
)

struct UpdateTodoRequest {
    1: string id(api.body="id", openapi.property='{
        title: "待办事项ID",
        description: "要更新的待办事项ID",
        type: "string"
    }')
    2: optional string title(api.body="title", openapi.property='{
        title: "标题",
        type: "string"
    }')
    3: optional string content(api.body="content", openapi.property='{
        title: "内容",
        type: "string"
    }')
    4: optional i64 start_time(api.body="start_time", openapi.property='{
        title: "开始时间",
        type: "integer",
        format: "int64"
    }')
    5: optional i64 end_time(api.body="end_time", openapi.property='{
        title: "结束时间",
        type: "integer",
        format: "int64"
    }')
    6: optional i16 is_all_day(api.body="is_all_day", openapi.property='{
        title: "是否全天",
        type: "integer"
    }')
    7: optional i16 status(api.body="status", openapi.property='{
        title: "状态",
        description: "0-未完成，1-已完成",
        type: "integer"
    }')
    8: optional i16 priority(api.body="priority", openapi.property='{
        title: "优先级",
        description: "1-紧急且重要，2-重要不紧急，3-紧急不重要，4-不重要不紧急",
        type: "integer"
    }')
    9: optional i64 remind_at(api.body="remind_at", openapi.property='{
        title: "提醒时间",
        type: "integer",
        format: "int64"
    }')
    10: optional string category(api.body="category", openapi.property='{
        title: "分类",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "更新待办事项请求",
        description: "更新待办事项信息，只传需要更新的字段",
        required: ["id"]
    }'
)

struct UpdateTodoResponse {
    1: string id(api.body="id", openapi.property='{
        title: "待办事项ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "更新待办事项响应",
        description: "返回更新成功的待办事项ID",
        required: ["id"]
    }'
)

struct DeleteTodoRequest {
    1: string id(api.query="id", openapi.property='{
        title: "待办事项ID",
        description: "要删除的待办事项ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "删除待办事项请求",
        description: "根据ID删除待办事项",
        required: ["id"]
    }'
)

struct DeleteTodoResponse {
    1: string id(api.body="id", openapi.property='{
        title: "待办事项ID",
        description: "已删除的待办事项ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "删除待办事项响应",
        description: "返回已删除的待办事项ID",
        required: ["id"]
    }'
)

// ==================== 知识库(Summarize)相关接口 ====================

struct GetSummaryRequest {
    1: string id(api.query="id", openapi.property='{
        title: "摘要ID",
        description: "要获取的摘要ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "获取摘要详情请求",
        description: "根据ID获取摘要详情",
        required: ["id"]
    }'
)

struct SummaryItem {
    1: string id(api.body="id", openapi.property='{
        title: "摘要ID",
        type: "string"
    }')
    2: string conversation_id(api.body="conversation_id", openapi.property='{
        title: "对话ID",
        type: "string"
    }')
    3: string summary_text(api.body="summary_text", openapi.property='{
        title: "摘要内容",
        type: "string"
    }')
    4: list<string> tags(api.body="tags", openapi.property='{
        title: "标签列表",
        type: "array"
    }')
    5: string tool_calls_json(api.body="tool_calls_json", openapi.property='{
        title: "工具调用JSON",
        type: "string"
    }')
    6: map<string,string> notes(api.body="notes", openapi.property='{
        title: "笔记",
        type: "object"
    }')
    7: i64 created_at(api.body="created_at", openapi.property='{
        title: "创建时间",
        type: "integer",
        format: "int64"
    }')
    8: i64 updated_at(api.body="updated_at", openapi.property='{
        title: "更新时间",
        type: "integer",
        format: "int64"
    }')
}(
    openapi.schema='{
        title: "摘要详细信息",
        description: "摘要的详细信息"
    }'
)

struct GetSummaryResponse {
    1: SummaryItem summary(api.body="summary", openapi.property='{
        title: "摘要信息",
        type: "object"
    }')
}(
    openapi.schema='{
        title: "获取摘要详情响应",
        description: "返回摘要详细信息",
        required: ["summary"]
    }'
)

struct ListSummaryRequest {
}(
    openapi.schema='{
        title: "获取摘要列表请求",
        description: "获取用户的所有摘要"
    }'
)

struct ListSummaryResponse {
    1: list<SummaryItem> summaries(api.body="summaries", openapi.property='{
        title: "摘要列表",
        type: "array"
    }')
}(
    openapi.schema='{
        title: "摘要列表响应",
        description: "返回摘要列表",
        required: ["summaries"]
    }'
)

struct UpdateSummaryRequest {
    1: string id(api.body="id", openapi.property='{
        title: "摘要ID",
        description: "要更新的摘要ID",
        type: "string"
    }')
    2: optional string summary_text(api.body="summary_text", openapi.property='{
        title: "摘要内容",
        type: "string"
    }')
    3: optional list<string> tags(api.body="tags", openapi.property='{
        title: "标签列表",
        type: "array"
    }')
    4: optional string tool_calls_json(api.body="tool_calls_json", openapi.property='{
        title: "工具调用JSON",
        type: "string"
    }')
    5: optional map<string,string> notes(api.body="notes", openapi.property='{
        title: "笔记",
        type: "object"
    }')
}(
    openapi.schema='{
        title: "更新摘要请求",
        description: "更新摘要信息，只传需要更新的字段",
        required: ["id"]
    }'
)

struct UpdateSummaryResponse {
    1: string id(api.body="id", openapi.property='{
        title: "摘要ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "更新摘要响应",
        description: "返回更新成功的摘要ID",
        required: ["id"]
    }'
)

struct DeleteSummaryRequest {
    1: string id(api.query="id", openapi.property='{
        title: "摘要ID",
        description: "要删除的摘要ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "删除摘要请求",
        description: "根据ID删除摘要",
        required: ["id"]
    }'
)

struct DeleteSummaryResponse {
    1: string id(api.body="id", openapi.property='{
        title: "摘要ID",
        description: "已删除的摘要ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "删除摘要响应",
        description: "返回已删除的摘要ID",
        required: ["id"]
    }'
)

struct CourseListRequest {
    1: required string term(api.query="term", openapi.property='{
        title: "学期",
        description: "查询的学期，例如202501",
        type: "string"
    }')
    2: optional bool is_refresh(api.query="is_refresh", openapi.property='{
        title: "是否刷新",
        description: "是否强制刷新课程数据，默认为false",
        type: "boolean"
    }')
}(
    openapi.schema='{
        title: "课程列表请求",
        description: "获取指定学期的课程列表",
        required: ["term"]
    }'
)

struct CourseListResponse {
    1: required model.BaseResp base(api.body="base", openapi.property='{
        title: "基础响应",
        description: "响应的基础信息",
        type: "object"
    }')
    2: required list<model.Course> data(api.body="data", openapi.property='{
        title: "课程列表",
        description: "指定学期的课程列表",
        type: "array"
    }')
}(
    openapi.schema='{
        title: "课程列表响应",
        description: "包含课程列表的响应",
        required: ["base", "data"]
    }'
)

struct CourseTermListRequest{}(
    openapi.schema='{
        title: "学期列表请求",
        description: "获取可用的学期列表"
    }')

struct CourseTermListResponse{
    1: required model.BaseResp base(api.body="base", openapi.property='{
        title: "基础响应",
        description: "响应的基础信息",
        type: "object"
    }')
    2: required list<string> data(api.body="data", openapi.property='{
        title: "学期列表",
        description: "可用的学期列表",
        type: "array"
    }')
}

struct UpdateUserSettingRequest {
    1: string setting_json(api.body="setting_json", openapi.property='{
        title: "用户设置JSON",
        description: "用户设置JSON字符串",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "更新用户设置请求",
        description: "更新用户个性化设置",
        required: ["setting_json"]
    }'
)

struct UpdateUserSettingResponse {
    1: string user_id(api.body="user_id", openapi.property='{
        title: "用户ID",
        description: "更新成功的用户ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "更新用户设置响应",
        description: "返回更新结果",
        required: ["user_id"]
    }'
)

struct TermListRequest {}(
    openapi.schema='{
        title: "获取学期列表请求",
        description: "获取所有可用的学期列表"
    }'
)

struct TermListResponse {
    1: required model.BaseResp base(api.body="base", openapi.property='{
        title: "基础响应",
        description: "响应的基础信息",
        type: "object"
    }')
    2: required model.TermList term_lists(api.body="term_lists", openapi.property='{
        title: "学期列表",
        description: "包含所有可用学期的列表",
        type: "object"
    }')
}

// 学期信息
struct TermRequest {
    1: required string term(api.query="term", openapi.property='{
        title: "学期标识",
        description: "要查询的学期标识，例如202501",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "获取学期详情请求",
        description: "根据学期标识获取学期详情",
        required: ["term"]
    }'
)

struct TermResponse {
    1: required model.BaseResp base(api.body="base", openapi.property='{
        title: "基础响应",
        description: "响应的基础信息",
        type: "object"
    }')
    2: required model.TermInfo term_info(api.body="term_info", openapi.property='{
        title: "学期详情",
        description: "指定学期的详细信息",
        type: "object"
    }')
}


service ApiService {
    // 非流式对话
    ChatResponse Chat(1: ChatRequest req)(api.post="/api/v1/chat")
    // 流式对话
    ChatSSEHandlerResponse ChatSSE(1: ChatSSEHandlerRequest req)(api.post="/api/v1/chat/sse")
    // 示例接口 idl写好后运行make hertz-gen-api生成脚手架
    TemplateResponse Template(1: TemplateRequest req)(api.post="/api/v1/template")
    // 获取会话历史
    GetConversationHistoryResponse GetConversationHistory(1: GetConversationHistoryRequest req)(api.get="/api/v1/conversation/history")
    // 获取对话列表
    ListConversationsResponse ListConversations(1: ListConversationsRequest req)(api.get="/api/v1/conversation/list")
    // 会话总结
    SummarizeConversationResponse SummarizeConversation(1: SummarizeConversationRequest req)(api.post="/api/v1/conversation/summarize")
    // 获取jwch登录数据
    GetLoginDataResponse GetLoginData(1: GetLoginDataRequest req)(api.post="/api/v1/user/login")
    // 获取用户信息
    GetUserInfoResponse GetUserInfo(1: GetUserInfoRequest req)(api.get="/api/v1/user/info")
    // 更新用户设置
    UpdateUserSettingResponse UpdateUserSetting(1: UpdateUserSettingRequest req)(api.put="/api/v1/user/setting")
    
    // 待办事项管理
    // 创建待办事项
    CreateTodoResponse CreateTodo(1: CreateTodoRequest req)(api.post="/api/v1/todo/create")
    // 获取待办事项详情
    GetTodoResponse GetTodo(1: GetTodoRequest req)(api.get="/api/v1/todo/detail")
    // 获取所有待办事项列表
    ListTodoResponse ListTodo(1: ListTodoRequest req)(api.get="/api/v1/todo/list")
    // 搜索待办事项
    SearchTodoResponse SearchTodo(1: SearchTodoRequest req)(api.get="/api/v1/todo/search")
    // 更新待办事项
    UpdateTodoResponse UpdateTodo(1: UpdateTodoRequest req)(api.put="/api/v1/todo/update")
    // 删除待办事项
    DeleteTodoResponse DeleteTodo(1: DeleteTodoRequest req)(api.delete="/api/v1/todo/delete")

    // 知识库(摘要)管理
    // 获取摘要详情
    GetSummaryResponse GetSummary(1: GetSummaryRequest req)(api.get="/api/v1/summary/detail")
    // 获取所有摘要列表
    ListSummaryResponse ListSummary(1: ListSummaryRequest req)(api.get="/api/v1/summary/list")
    // 更新摘要
    UpdateSummaryResponse UpdateSummary(1: UpdateSummaryRequest req)(api.put="/api/v1/summary/update")
    // 删除摘要
    DeleteSummaryResponse DeleteSummary(1: DeleteSummaryRequest req)(api.delete="/api/v1/summary/delete")

    // 课程相关接口
    // 获取课表
    CourseListResponse GetCourseList(1: CourseListRequest req)(api.get="/api/v1/course/list")
    // 获取学期
    CourseTermListResponse GetTermList(1: CourseTermListRequest req)(api.get="/api/v1/course/term/list")
    // 校历信息：学期列表
    TermListResponse GetTermsList(1: TermListRequest req) (api.get="/api/v1/terms/list")
    // 校历信息：学期详情
    TermResponse GetTerm(1: TermRequest req) (api.get="/api/v1/terms/info")
}