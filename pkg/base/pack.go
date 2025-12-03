package base

import (
	"errors"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	jwchErrno "github.com/west2-online/jwch/errno"
)

var EvaluationNotFoundError = errno.NewErrNo(errno.BizJwchEvaluationNotFoundCode, "请先对任课教师进行评价")

// HandleJwchError 对于jwch库返回的错误类型，需要使用 HandleJwchError 来保留 cookie 异常
func HandleJwchError(err error) error {
	var jwchErr jwchErrno.ErrNo
	if errors.As(err, &jwchErr) {
		if errors.Is(jwchErr, jwchErrno.EvaluationNotFoundError) {
			return EvaluationNotFoundError
		}
	}
	if errors.As(err, &jwchErr) {
		if errors.Is(jwchErr, jwchErrno.CookieError) {
			return errno.NewErrNo(errno.BizJwchCookieExceptionCode, jwchErr.ErrorMsg)
		}
	}
	return err
}
