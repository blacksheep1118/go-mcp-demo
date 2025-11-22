package mw

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type loginDataKey struct{}

// LoginData 中间件往 ctx 里塞的结构
type LoginData struct {
	ID     string
	Cookie string
}

// GetHeader 把前端带来的 X-Student-Id 与 X-Jwch-Cookie 注入上下文
func GetHeader() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		id := string(ctx.GetHeader("X-Student-Id"))
		cookie := string(ctx.GetHeader("X-Jwch-Cookie"))
		if id == "" || cookie == "" {
			hlog.Warn("missing jwch header")
			// 继续走，后面 handler 自己负责报错
		}
		ld := &LoginData{ID: id, Cookie: cookie}
		c = context.WithValue(c, loginDataKey{}, ld)
		ctx.Next(c)
	}
}

// ExtractLoginData 供下游 handler 取出 loginData
func ExtractLoginData(c context.Context) (*LoginData, bool) {
	v, ok := c.Value(loginDataKey{}).(*LoginData)
	return v, ok
}
