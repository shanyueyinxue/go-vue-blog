package cache

import (
	"fmt"
	"time"
)

// Cache 通用缓存接口
type Cache interface {
	// Set 设置缓存值，expiration为过期时间，0表示永不过久
	Set(key string, value any, expiration time.Duration) error

	// GetString 获取字符串缓存值
	GetString(key string) (string, error)
	// GetInt 获取整数缓存值
	GetInt(key string) (int, error)
	// GetFloat 获取浮点数缓存值
	GetFloat(key string) (float64, error)
	// GetBool 获取布尔值缓存值
	GetBool(key string) (bool, error)

	// Delete 删除缓存值
	Delete(key string) error

	// Exists 检查键是否存在
	Exists(key string) bool

	// Ping 检查缓存服务是否可用
	Ping() error

	// Close 关闭缓存
	Close() error
	// Self 返回 客户端实例
	Self() any
}

func InitCache(cfg CacheConfig) (Cache, error) {
	var cache Cache
	switch cfg.Driver {
	case "redis":
		fmt.Println("use redis cache")
		cache = NewRedisCache(cfg.Redis)
	case "memory":
		fmt.Println("use memory cache")
		cache = NewMemoryCache(cfg.Memory)
	default:
		return nil, fmt.Errorf("unsupported cache driver: %s", cfg.Driver)
	}
	return cache, nil
}

// CacheError 缓存错误类型
type CacheError struct {
	message string
}

func (e *CacheError) Error() string {
	return e.message
}

// 错误定义
var (
	ErrKeyNotFound     = &CacheError{"key not found"}
	ErrKeyExpired      = &CacheError{"key expired"}
	ErrInvalidType     = &CacheError{"invalid type"}
	ErrListEmpty       = &CacheError{"list is empty"}
	ErrIndexOutOfRange = &CacheError{"list index out of range"}
)

// 空 Cache 实现
type EmptyCache struct {
}

var _ Cache = (*EmptyCache)(nil)

func (c *EmptyCache) Set(key string, value any, expiration time.Duration) error {
	return nil
}

func (c *EmptyCache) GetString(key string) (string, error) {
	return "", ErrKeyNotFound
}

func (c *EmptyCache) GetInt(key string) (int, error) {
	return 0, ErrKeyNotFound
}

func (c *EmptyCache) GetFloat(key string) (float64, error) {
	return 0, ErrKeyNotFound
}

func (c *EmptyCache) GetBool(key string) (bool, error) {
	return false, ErrKeyNotFound
}

func (c *EmptyCache) Delete(key string) error {
	return nil
}

func (c *EmptyCache) Exists(key string) bool {
	return false
}

func (c *EmptyCache) Ping() error {
	return nil
}

func (c *EmptyCache) Close() error {
	return nil
}

func (c *EmptyCache) Self() any {
	return c
}
