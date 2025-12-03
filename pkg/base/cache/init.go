package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// InitRedisClient 初始化 Redis 客户端
func InitRedisClient() (*redis.Client, error) {
	if config.Redis == nil {
		return nil, errors.New("redis config is nil")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       1,
	})
	l := logger.GetRedisLogger()
	redis.SetLogger(l)
	client.AddHook(l)
	_, err := client.Ping(context.TODO()).Result()
	if err != nil {
		return nil, fmt.Errorf("client.NewRedisClient: ping redis failed: %w", err)
	}
	return client, nil
}
