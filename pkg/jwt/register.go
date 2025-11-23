package jwt

import (
	"errors"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type Claims struct {
	StudentID string `json:"student_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成 accessToken
func GenerateAccessToken(stuID string) (string, error) {
	tokenID := uuid.New().String()
	claims := &Claims{
		StudentID: stuID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(constant.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "hachimi",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Server.Secret))
	if err != nil {
		logger.Error("generate token failed", zap.Error(err), zap.Any("claims", claims))
		return "", ErrGenerateToken
	}
	return tokenString, nil
}

func VerifyAccessToken(token string) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(config.Server.Secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			zap.L().Warn("parse token error", zap.Error(err), zap.String("token", token))
		}
		return nil, ErrInvalidToken
	}

	return claims, nil
}
