package app

import (
	"blog/pkg/cache"
	"blog/pkg/captcha"
	"blog/pkg/email"
	"blog/pkg/jwt"
	"blog/pkg/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App 应用依赖容器，持有全局共享的组件实例
type App struct {
	Config  *Config                 // 应用配置
	Logger  *zap.Logger             // 日志记录器
	DB      *gorm.DB                // 数据库连接
	Cache   cache.Cache             // 缓存（redis / memory）
	JWT     *jwt.JWT                // JWT 工具
	Email   email.EmailService      // 邮件服务
	Captcha *captcha.CaptchaManager // 验证码管理器
	Storage storage.Storage         // 文件存储（local / oss）
}
