package cache

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var ErrCacheClosed = errors.New("memory: cache is closed")

// MemoryCache 内存缓存实现
type MemoryCache struct {
	items     map[string]*cacheItem
	mutex     sync.RWMutex
	stop      chan struct{}
	interval  time.Duration // 清理间隔
	closed    bool          // 缓存是否已关闭
}

// cacheItem 缓存项
type cacheItem struct {
	value      any
	expiration int64 // 过期时间戳，0表示永不过期
}

var _ Cache = (*MemoryCache)(nil)

// NewMemoryCache 创建新的内存缓存实例
// cleanupInterval: 清理过期缓存的时间间隔，0表示不自动清理
func NewMemoryCache(config MemoryConfig) *MemoryCache {
	// 输出警告
	fmt.Println("Memory cache 在生产环境中不建议使用，重启会丢失缓存数据")

	cleanupInterval := time.Duration(0)
	c := config.CleanupInterval
	if c > 0 {
		cleanupInterval = time.Duration(c) * time.Second
	}

	cache := &MemoryCache{
		items:     make(map[string]*cacheItem),
		stop:      make(chan struct{}),
		interval:  cleanupInterval,
		closed:    false,
	}

	// 启动后台清理任务
	if cleanupInterval > 0 {
		go cache.cleanupExpired()
	}

	return cache
}

// Set 设置缓存值
func (m *MemoryCache) Set(key string, value any, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.IsClosed() {
		return ErrCacheClosed
	}

	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	m.items[key] = &cacheItem{
		value:      value,
		expiration: exp,
	}

	return nil
}

// get 获取缓存值（内部方法）
func (m *MemoryCache) get(key string) (any, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.IsClosed() {
		return nil, ErrCacheClosed
	}

	item, exists := m.items[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	// 检查是否过期
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		return nil, ErrKeyExpired
	}

	return item.value, nil
}

// GetString 获取字符串缓存值
func (m *MemoryCache) GetString(key string) (string, error) {
	value, err := m.get(key)
	if err != nil {
		return "", err
	}
	v, ok := value.(string)
	if !ok {
		return "", ErrInvalidType
	}
	return v, nil
}

// GetInt 获取整数缓存值
func (m *MemoryCache) GetInt(key string) (int, error) {
	value, err := m.get(key)
	if err != nil {
		return 0, err
	}
	v, ok := value.(int)
	if !ok {
		return 0, ErrInvalidType
	}
	return v, nil
}

// GetFloat 获取浮点数缓存值
func (m *MemoryCache) GetFloat(key string) (float64, error) {
	value, err := m.get(key)
	if err != nil {
		return 0, err
	}
	v, ok := value.(float64)
	if !ok {
		return 0, ErrInvalidType
	}
	return v, nil
}

// GetBool 获取布尔值缓存值
func (m *MemoryCache) GetBool(key string) (bool, error) {
	value, err := m.GetString(key)
	if err != nil {
		return false, err
	}

	// 转换为布尔值
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, ErrInvalidType
	}
}

// Delete 删除缓存值
func (m *MemoryCache) Delete(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.IsClosed() {
		return ErrCacheClosed
	}

	if _, exists := m.items[key]; !exists {
		return ErrKeyNotFound
	}

	delete(m.items, key)
	return nil
}

// Exists 检查键是否存在且未过期
func (m *MemoryCache) Exists(key string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.IsClosed() {
		return false
	}

	item, exists := m.items[key]
	if !exists {
		return false
	}

	// 检查是否过期
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		return false
	}

	return true
}

// GetWithExpiration 获取缓存值及其过期时间（扩展方法）
func (m *MemoryCache) GetWithExpiration(key string) (any, time.Time, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.IsClosed() {
		return nil, time.Time{}, ErrCacheClosed
	}

	item, exists := m.items[key]
	if !exists {
		return nil, time.Time{}, ErrKeyNotFound
	}

	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		return nil, time.Time{}, ErrKeyExpired
	}

	var expTime time.Time
	if item.expiration > 0 {
		expTime = time.Unix(0, item.expiration)
	}

	return item.value, expTime, nil
}

// cleanupExpired 清理过期缓存的后台任务
func (m *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.deleteExpired()
		case <-m.stop:
			return
		}
	}
}

// deleteExpired 删除所有过期缓存
func (m *MemoryCache) deleteExpired() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now().UnixNano()
	for key, item := range m.items {
		if item.expiration > 0 && now > item.expiration {
			delete(m.items, key)
		}
	}
}

// StopCleanup 停止后台清理任务
func (m *MemoryCache) StopCleanup() {
	close(m.stop)
}

// Close 关闭缓存
func (m *MemoryCache) Close() error {
	if m.IsClosed() {
		return ErrCacheClosed
	}

	m.StopCleanup()
	m.closed = true
	m.items = nil
	return nil
}
func (m *MemoryCache) IsClosed() bool {
	return m.closed
}

// Ping 检查缓存是否可用
func (m *MemoryCache) Ping() error {
	if m.IsClosed() {
		return ErrCacheClosed
	}
	return nil
}

func (m *MemoryCache) Self() any {
	return m
}
