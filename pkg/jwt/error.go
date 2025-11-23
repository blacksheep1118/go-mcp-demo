package jwt

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidToken  = status.Error(codes.Unauthenticated, "invalid token")
	ErrParseToken    = status.Error(codes.Unauthenticated, "parse token error")
	ErrGenerateToken = status.Error(codes.Unauthenticated, "generate token error")
	ErrEmptyToken    = status.Error(codes.Unauthenticated, "empty token")
	ErrTokenExpired  = status.Error(codes.Unauthenticated, "token expired")
)
