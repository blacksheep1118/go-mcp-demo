package router

import (
	"github.com/FantasyRL/go-mcp-demo/api/handler"
	"github.com/FantasyRL/go-mcp-demo/api/handler/api"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func customizedRegister(r *server.Hertz) {
	r.GET("/ping", handler.Ping)
}

// RegisterCustom 手写路由
func RegisterCustom(h *server.Hertz) { // 注意参数叫 h
	h.POST("/api/user/login-jwch", api.LoginByJWC)
}
