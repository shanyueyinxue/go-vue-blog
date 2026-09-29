package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache 基于 go-redis 的缓存实现
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache(config RedisConfig) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:            config.Host + ":" + config.Port,
		Password:        config.Password,
		DB:              config.DB,
		PoolSize:        config.MaxOpenConnections,
		ConnMaxLifetime: time.Duration(config.ConnMaxLifetime) * time.Second,
		ConnMaxIdleTime: time.Duration(config.IdleTimeout) * time.Second,
		MaxIdleConns:    config.MaxIdleConnections,
		DialTimeout:     5 * time.Second,
		PoolTimeout:     time.Duration(config.WaitTimeout) * time.Millisecond,
		ReadTimeout:     time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(config.WriteTimeout) * time.Second,
	})

	return &RedisCache{
		client: client,
		ctx:    context.Background(),
	}
}

// Set 设置缓存值
func (r *RedisCache) Set(key string, value any, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// GetString 获取字符串值
func (r *RedisCache) GetString(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// GetInt 获取整数值
func (r *RedisCache) GetInt(key string) (int, error) {
	result, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// 转换为整数
	val, err := strconv.Atoi(result)
	if err != nil {
		return 0, fmt.Errorf("无法将值转换为整数: %v", err)
	}
	return val, nil
}

// GetFloat 获取浮点数值
func (r *RedisCache) GetFloat(key string) (float64, error) {
	result, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// 转换为浮点数
	val, err := strconv.ParseFloat(result, 64)
	if err != nil {
		return 0, fmt.Errorf("无法将值转换为浮点数: %v", err)
	}
	return val, nil
}

// GetBool 获取布尔值
func (r *RedisCache) GetBool(key string) (bool, error) {
	result, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	// 转换为布尔值
	switch strings.ToLower(result) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("无法将值转换为布尔值: %s", result)
	}
}

// Delete 删除缓存值
func (r *RedisCache) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(key string) bool {
	result, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false
	}
	return result > 0
}

// Close 关闭 Redis 连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Ping 测试连接
func (r *RedisCache) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

// Self 返回 Redis 客户端实例
func (r *RedisCache) Self() any {
	return r.client
}
