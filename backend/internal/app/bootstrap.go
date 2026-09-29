package app

import (
	"fmt"
	"time"

	"blog/internal/model"
	"blog/pkg/cache"
	"blog/pkg/captcha"
	"blog/pkg/database"
	"blog/pkg/email"
	"blog/pkg/jwt"
	"blog/pkg/logger"
	"blog/pkg/storage"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// New 初始化应用全部依赖并返回 App 实例
func New(cfg *Config) (*App, error) {
	// 1. 设置时区（统一所有时间使用配置时区）
	if loc, err := time.LoadLocation(cfg.App.Timezone); err == nil {
		time.Local = loc
	}

	// 2. 初始化日志
	log, err := logger.NewLogger(cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	// 3. 初始化数据库并自动迁移表结构
	// TranslateError: 将驱动唯一键/外键等错误翻译为 gorm.ErrDuplicatedKey 等，
	// 供服务层「slug 唯一键冲突重试」等逻辑统一识别（见 internal/service/slug_retry.go）
	db, err := database.SetupDatabase(cfg.DB, &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}
	// 启动时确保存在管理员账户（单用户后台）
	if err := ensureAdmin(db, cfg.User); err != nil {
		return nil, fmt.Errorf("创建管理员失败: %w", err)
	}

	// 4. 初始化缓存（redis / memory）
	cacheIns, err := cache.InitCache(cfg.Cache)
	if err != nil {
		return nil, fmt.Errorf("初始化缓存失败: %w", err)
	}

	// 5. 初始化邮件服务（用于验证码、评论通知等）
	emailSvc := email.NewGomailService()
	if err := emailSvc.Init(&cfg.Email); err != nil {
		log.Warn("邮件服务初始化失败（邮件相关功能将不可用）", zap.Error(err))
	}

	// 6. 初始化验证码管理器
	captchaMgr, err := captcha.InitCaptcha(cfg.Captcha, captcha.CaptchaOption{
		EmailService: emailSvc,
		Cache:        cacheIns,
		Logger:       log,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化验证码失败: %w", err)
	}

	// 7. 初始化文件存储（local / oss）
	storageIns, err := storage.NewStorage(cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("初始化存储失败: %w", err)
	}

	return &App{
		Config:  cfg,
		Logger:  log,
		DB:      db,
		Cache:   cacheIns,
		JWT:     jwt.NewJWT(cfg.JWT),
		Email:   emailSvc,
		Captcha: captchaMgr,
		Storage: storageIns,
	}, nil
}

// migrate 执行数据库自动迁移（docs/database.md 第 7 节）
func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Post{},
		&model.Category{},
		&model.Tag{},
		&model.PostTag{},
		&model.Comment{},
		&model.SiteConfig{},
		&model.Friend{},
		&model.Work{},
		&model.File{},
		&model.Music{},
	)
}

// ensureAdmin 保证启动时存在管理员账户：
// 若 users 表为空则按配置创建超级管理员（密码 bcrypt 加密）
func ensureAdmin(db *gorm.DB, cfg UserConfig) error {
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // 已存在用户，跳过创建
	}

	// 生成 bcrypt 哈希密码（cost 取配置值，默认 12）
	cost := cfg.BcryptCost
	if cost < 10 {
		cost = 12
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.MasterPassword), cost)
	if err != nil {
		return err
	}

	admin := model.User{
		Username: cfg.MasterUsername,
		Password: string(hash),
		Email:    cfg.MasterEmail,
		Phone:    cfg.MasterPhone,
		IsMaster: 1,
		Role:     "admin",
		Status:   1,
	}
	return db.Create(&admin).Error
}
