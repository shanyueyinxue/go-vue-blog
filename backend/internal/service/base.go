package service

import (
	"blog/internal/app"
	"blog/pkg/cache"
	"blog/pkg/jwt"
	"blog/pkg/storage"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service 所有业务服务的基类，统一持有应用依赖
type Service struct {
	App *app.App // 应用依赖容器
}

// New 创建业务服务基类
func New(a *app.App) *Service {
	return &Service{App: a}
}

// DB 便捷访问数据库
func (s *Service) DB() *gorm.DB {
	return s.App.DB
}

// Cache 便捷访问缓存
func (s *Service) Cache() cache.Cache {
	return s.App.Cache
}

// JWT 便捷访问 JWT 工具
func (s *Service) JWT() *jwt.JWT {
	return s.App.JWT
}

// Storage 便捷访问文件存储
func (s *Service) Storage() storage.Storage {
	return s.App.Storage
}

// Log 便捷访问日志
func (s *Service) Log() *zap.Logger {
	return s.App.Logger
}
