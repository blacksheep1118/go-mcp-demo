package utils

import "context"

type loginDataKey struct{}

// LoginData 中间件往 ctx 里塞的结构
type LoginData struct {
	ID     string
	Cookie string
}

func WithLoginData(ctx context.Context, loginData *LoginData) context.Context {
	return context.WithValue(ctx, loginDataKey{}, loginData)
}

func WithStuID(ctx context.Context, stuID string) context.Context {
	return context.WithValue(ctx, "stu_id", stuID)
}

// ExtractLoginData 供下游 handler 取出 loginData
func ExtractLoginData(ctx context.Context) (*LoginData, bool) {
	v, ok := ctx.Value(loginDataKey{}).(*LoginData)
	return v, ok
}

// ExtractStuID 供下游 handler 取出 stu_id
func ExtractStuID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value("stu_id").(string)
	return v, ok
}
