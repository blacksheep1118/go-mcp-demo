namespace go model
include "openapi.thrift"
struct BaseResp {
    1: i64 code (api.body="code", openapi.property='{
        title: "状态码",
        description: "响应状态码",
        type: "integer"
    }')
    2: string msg (api.body="msg", openapi.property='{
        title: "消息",
        description: "响应消息",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "基础响应",
        description: "所有响应的基础结构",
        required: ["code", "msg"]
    }'
)

struct User {
    1: string id (api.body="id", openapi.property='{
        title: "用户ID",
        description: "唯一标识用户的ID",
        type: "string"
    }')
    2: string name (api.body="name", openapi.property='{
        title: "用户名",
        description: "用户的显示名称",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "用户信息",
        description: "包含用户基本信息的结构",
        required: ["id", "name"]
    }'
)

/*
    以下代码来自fzuhelper-server
 */
struct Course {
    1: required string name(api.body="name", openapi.property='{
        title: "课程名称",
        description: "课程的名称",
        type: "string"
    }')                            // 课程名称
    2: required string teacher(api.body="teacher",openapi.property='{
        title: "教师姓名",
        description: "授课教师的姓名",
        type: "string"
    }')                          // 教师
    3: required list<CourseScheduleRule> scheduleRules(api.body="scheduleRules", openapi.property='{
        title: "排课规则",
        description: "课程的排课规则列表",
        type: "array"

    }')  // 排课规则
    4: required string remark(api.body="remark", openapi.property='{
        title: "备注",
        description: "课程的备注信息",
        type: "string"
    }')                           // 备注
    5: required string lessonplan(api.body="lessonplan", openapi.property='{
        title: "授课计划",
        description: "课程的授课计划",
        type: "string"
    }')                       // 授课计划
    6: required string syllabus(api.body="syllabus", openapi.property='{
        title: "教学大纲",
        description: "课程的教学大纲",
        type: "string"
    }')                         // 教学大纲
    7: required string rawScheduleRules(api.body="rawScheduleRules", openapi.property='{
        title: "原始排课规则",
        description: "课程的原始排课规则数据",
        type: "string"
    }')                 // (原始数据) 排课规则
    8: required string rawAdjust(api.body="rawAdjust", openapi.property='{
        title: "原始调课规则",
        description: "课程的原始调课规则数据",
        type: "string"
    }')                        // (原始数据) 调课规则
    9: required string examType(api.body="examType", openapi.property='{
        title: "考试类型",
        description: "课程的考试类型",
        type: "string"
    }')                        // 考试类型(用于查看是否免听
}

// 课程安排
struct CourseScheduleRule {
    1: required string location(api.body="location", openapi.property='{
        title: "上课地点",
        description: "课程的上课地点",
        type: "string"
    }')         // 定制
    2: required i64 startClass(api.body="startClass", openapi.property='{
        title: "开始节数",
        description: "课程的开始节数",
        type: "integer"
    }')          // 开始节数
    3: required i64 endClass(api.body="endClass", openapi.property='{
        title: "结束节数",
        description: "课程的结束节数",
        type: "integer"
    }')            // 结束节数
    4: required i64 startWeek(api.body="startWeek", openapi.property='{
        title: "起始周",
        description: "课程的起始周数",
        type: "integer"
    }')           // 起始周
    5: required i64 endWeek(api.body="endWeek", openapi.property='{
        title: "结束周",
        description: "课程的结束周数",
        type: "integer"
    }')             // 结束周
    6: required i64 weekday(api.body="weekday", openapi.property='{
        title: "星期几",
        description: "课程的上课星期几",
        type: "integer"
    }')             // 星期几
    7: required bool single(api.body="single", openapi.property='{
        title: "单周",
        description: "课程是否在单周上课",
        type: "boolean"
    }')             // 单周
    8: required bool double(api.body="double", openapi.property='{
        title: "双周",
        description: "课程是否在双周上课",
        type: "boolean"
    }')             // 双周
    9: required bool adjust(api.body="adjust", openapi.property='{
        title: "调课标志",
        description: "课程是否为调课",
        type: "boolean"
    }')             // 是否是调课
}

struct Term {
    1: optional string term_id(api.body="term_id", openapi.property='{
        title: "学期ID",
        description: "唯一标识学期的ID",
        type: "string"
    }')
    2: optional string school_year(api.body="school_year", openapi.property='{
        title: "学年",
        description: "学期所属的学年",
        type: "string"
    }')
    3: optional string term(api.body="term", openapi.property='{
        title: "学期名称",
        description: "学期的名称",
        type: "string"
    }')
    4: optional string start_date(api.body="start_date", openapi.property='{
        title: "开始日期",
        description: "学期的开始日期",
        type: "string"
    }')
    5: optional string end_date(api.body="end_date", openapi.property='{
        title: "结束日期",
        description: "学期的结束日期",
        type: "string"
    }')
}

struct TermEvent {
    1: optional string name(api.body="name", openapi.property='{
        title: "事件名称",
        description: "学期事件的名称",
        type: "string"
    }')
    2: optional string start_date(api.body="start_date", openapi.property='{
        title: "开始日期",
        description: "事件的开始日期",
        type: "string"
    }')
    3: optional string end_date(api.body="end_date", openapi.property='{
        title: "结束日期",
        description: "事件的结束日期",
        type: "string"
    }')
}

struct TermList {
    1: optional string current_term(api.body="current_term", openapi.property='{
        title: "当前学期",
        description: "当前学期的标识",
        type: "string"
    }')
    2: optional list<Term> terms(api.body="terms", openapi.property='{
        title: "学期列表",
        description: "包含所有学期的列表",
        type: "array"
    }')
}

struct TermInfo {
    1: optional string term_id(api.body="term_id", openapi.property='{
        title: "学期ID",
        description: "唯一标识学期的ID",
        type: "string"
    }')
    2: optional string term(api.body="term", openapi.property='{
        title: "学期名称",
        description: "学期的名称",
        type: "string"
    }')
    3: optional string school_year(api.body="school_year", openapi.property='{
        title: "学年",
        description: "学期所属的学年",
        type: "string"
    }')
    4: optional list<TermEvent> events(api.body="events", openapi.property='{
        title: "学期事件列表",
        description: "包含学期相关事件的列表",
        type: "array"
    }')
}
