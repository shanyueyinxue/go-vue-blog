package app

import (
	"blog/pkg/cache"
	"blog/pkg/captcha"
	"blog/pkg/database"
	"blog/pkg/email"
	"blog/pkg/jwt"
	"blog/pkg/logger"
	"blog/pkg/storage"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"` // 监听端口
	Host string `mapstructure:"host"` // 监听地址
	// 可信反向代理地址列表（CIDR 或 IP）。来自这些代理的请求才会信任
	// X-Forwarded-For / X-Real-IP 头，用于获取真实客户端 IP（限流、防刷等）。
	// 未配置反向代理时留空即可，此时始终使用 TCP 直连地址。
	TrustedProxies []string `mapstructure:"trustedProxies"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	// 允许跨域访问的来源白名单（完整 Origin，如 https://blog.example.com）。
	// 留空表示不开放跨域（同源访问不受影响）。
	Origins []string `mapstructure:"origins"`
}

// BackupConfig 定时备份配置
type BackupConfig struct {
	Enabled bool   `mapstructure:"enabled"` // 是否启用每周数据库备份（导出 CSV 压缩后保存到 storage）
	Cron    string `mapstructure:"cron"`    // 6 段 cron 表达式（含秒），默认每周一 04:00
	Dir     string `mapstructure:"dir"`     // 备份保存目录（storage 相对目录），默认 backups
}

// AppConfig 应用配置
type AppConfig struct {
	Env      string `mapstructure:"env"`      // 运行环境：development / production
	Debug    bool   `mapstructure:"debug"`    // 是否开启调试模式
	Timezone string `mapstructure:"timezone"` // 时区，如 Asia/Shanghai
}

// UserConfig 单用户（管理员）初始化配置
type UserConfig struct {
	MasterUsername string `mapstructure:"masterUsername"` // 管理员用户名
	MasterPassword string `mapstructure:"masterPassword"` // 管理员初始密码
	MasterEmail    string `mapstructure:"masterEmail"`    // 管理员邮箱
	MasterPhone    string `mapstructure:"masterPhone"`    // 管理员手机号
	BcryptCost     int    `mapstructure:"bcryptCost"`     // bcrypt 加密强度
	LockDays       int    `mapstructure:"lockDays"`       // 注销前锁定天数（预留）
}

// Config 总配置（与 config.yaml 一一对应）
type Config struct {
	Server  ServerConfig            `mapstructure:"server"`
	App     AppConfig               `mapstructure:"app"`
	JWT     jwt.JwtConfig           `mapstructure:"jwt"`
	Log     logger.LogConfig        `mapstructure:"log"`
	DB      database.DatabaseConfig `mapstructure:"db"`
	Cache   cache.CacheConfig       `mapstructure:"cache"`
	Captcha captcha.CaptchaConfig   `mapstructure:"captcha"`
	Email   email.EmailConfig       `mapstructure:"email"`
	User    UserConfig              `mapstructure:"user"`
	Storage storage.StorageConfig   `mapstructure:"storage"`
	CORS    CORSConfig              `mapstructure:"cors"`
	Backup  BackupConfig            `mapstructure:"backup"`
}
