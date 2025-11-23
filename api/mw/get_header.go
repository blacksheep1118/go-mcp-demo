package mw

import (
	"context"
	"github.com/FantasyRL/go-mcp-demo/api/pack"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/jwt"
	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
	"github.com/cloudwego/hertz/pkg/app"
)

// GetHeaderParams 把前端带来的 Id 与 Cookie 注入上下文
func GetHeaderParams() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		id := string(c.GetHeader("Id"))
		cookies := string(c.GetHeader("Cookies"))
		if id == "" || cookies == "" {
			pack.RespError(c, errno.AuthInvalid)
			c.Abort()
			return
		}
		ld := &utils.LoginData{ID: id, Cookie: cookies}
		ctx = utils.WithLoginData(ctx, ld)
		c.Next(ctx)
	}
}

func Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := string(c.GetHeader("Authorization"))
		cliams, err := jwt.VerifyAccessToken(token)
		if err != nil {
			pack.RespError(c, err)
			c.Abort()
			return
		}
		// 将 stu_id 传入 context
		ctx = utils.WithStuID(ctx, cliams.StudentID)
		c.Next(ctx)
	}
}
