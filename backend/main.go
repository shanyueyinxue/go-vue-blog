package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"blog/internal/app"
	"blog/internal/router"
	"blog/internal/service"
	"blog/pkg/config"
	"blog/pkg/jwt"
	"blog/pkg/scheduler"

	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置文件（默认 ./config.yaml；可通过环境变量覆盖）
	cfgLoader := config.NewConfig[app.Config](config.ConfigOptions{
		ConfigPath: ".",
		ConfigType: "yaml",
		ConfigName: "config",
		EnvPrefix:  "BLOG", // 环境变量前缀，如 BLOG_SERVER_PORT
	})
	// 生产环境未提供配置文件时，提供合理的默认值兜底
	cfgLoader.SetDefaultMap(map[string]any{
		"server.port":            8090,
		"server.host":            "127.0.0.1",
		"app.env":                "development",
		"app.debug":              true,
		"jwt.secret":             "change-me",
		"storage.type":           "local",
		"storage.local.basePath": "uploads",
		"storage.local.baseUrl":  "uploads",
		"backup.enabled":         true,
		"backup.cron":            "0 0 4 * * 1",
		"backup.dir":             "backups",
	})

	cfg, err := cfgLoader.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	// 空值兜底：配置文件里显式写空值（如空字符串）会绕过 SetDefault，
	// 这里统一回退默认值，避免端口 0、空驱动等导致行为异常或启动失败
	normalizeConfig(cfg)
	// 生产环境强制要求强 JWT 密钥，防止使用默认弱密钥被伪造 Token
	if cfg.App.Env == "production" {
		if err := jwt.ValidateSecret(cfg.JWT.Secret); err != nil {
			log.Fatalf("生产环境启动检查失败: %v", err)
		}
	}

	// 2. 初始化应用（数据库、缓存、日志、邮件、验证码、存储等）
	appInstance, err := app.New(cfg)
	if err != nil {
		log.Fatalf("应用初始化失败: %v", err)
	}
	defer appInstance.Cache.Close()

	// 3. 注册路由
	r := router.Setup(appInstance)

	// 4. 定时任务：每周数据库备份（可配置关闭）
	if cfg.Backup.Enabled {
		if cfg.Backup.Cron == "" {
			cfg.Backup.Cron = "0 0 4 * * 1"
		}
		sched, err := scheduler.NewScheduler()
		if err != nil {
			log.Fatalf("初始化调度器失败: %v", err)
		}
		// cfg.Backup.Cron = "*/10 * * * * *"
		backupSvc := service.NewBackupService(service.New(appInstance))
		if err := sched.AddTask(scheduler.NewTask("db-backup", cfg.Backup.Cron, func() {
			if _, err := backupSvc.RunBackup(); err != nil {
				appInstance.Logger.Warn("数据库备份失败", zap.Error(err))
			}
		})); err != nil {
			log.Fatalf("注册备份任务失败: %v", err)
		}
		if err := sched.Start(); err != nil {
			log.Fatalf("启动调度器失败: %v", err)
		}
		defer sched.Stop()
		appInstance.Logger.Info("数据库定时备份已启用", zap.String("cron", cfg.Backup.Cron))
	}

	// 5. 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	appInstance.Logger.Info("博客后端服务启动", zap.String("addr", addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appInstance.Logger.Fatal("服务器启动失败", zap.Error(err))
	}
}

// normalizeConfig 空值兜底：配置文件中显式写空值（空字符串 / 空数字）会绕过
// viper.SetDefault，导致拿到空端口、空驱动等不准确配置。这里把「有合理默认值
// 且空值无意义」的配置项统一回退到默认值。
func normalizeConfig(cfg *app.Config) {
	// 服务器
	if cfg.Server.Host == "" {
		cfg.Server.Host = "127.0.0.1"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8090
	}
	// 应用环境：空值按开发环境处理
	if cfg.App.Env == "" {
		cfg.App.Env = "development"
	}
	// JWT 密钥：空值回退默认（生产环境仍会被 ValidateSecret 拒绝）
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "change-me"
	}
	// 存储
	if cfg.Storage.Type == "" {
		cfg.Storage.Type = "local"
	}
	if cfg.Storage.Local.BasePath == "" {
		cfg.Storage.Local.BasePath = "uploads"
	}
	if cfg.Storage.Local.BaseURL == "" {
		cfg.Storage.Local.BaseURL = "uploads"
	}
	// 数据库驱动：空值回退 sqlite，否则启动即失败
	if cfg.DB.Driver == "" {
		cfg.DB.Driver = "sqlite"
	}
	// 验证码提供者：空值回退 email（与 config-example 一致）
	if cfg.Captcha.Option == "" {
		cfg.Captcha.Option = "email"
	}
	// 备份任务 cron：空值回退每周一 04:00，否则调度器注册失败
	if cfg.Backup.Cron == "" {
		cfg.Backup.Cron = "0 0 4 * * 1"
	}
}
