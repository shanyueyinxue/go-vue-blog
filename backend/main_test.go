package main

import (
	"testing"

	"blog/internal/app"
	"blog/pkg/database"
	"blog/pkg/jwt"
)

// TestNormalizeConfig 空值兜底：配置里显式写空值（空字符串/0）时，
// normalizeConfig 应回退默认值，避免绕过 SetDefault 导致配置不准确
func TestNormalizeConfig(t *testing.T) {
	cfg := &app.Config{} // 全部零值 = 配置文件里显式写了空值
	normalizeConfig(cfg)

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host 应为 127.0.0.1，实际 %q", cfg.Server.Host)
	}
	if cfg.Server.Port != 8090 {
		t.Errorf("Server.Port 应为 8090，实际 %d", cfg.Server.Port)
	}
	if cfg.App.Env != "development" {
		t.Errorf("App.Env 应为 development，实际 %q", cfg.App.Env)
	}
	if cfg.JWT.Secret != "change-me" {
		t.Errorf("JWT.Secret 应为 change-me，实际 %q", cfg.JWT.Secret)
	}
	if cfg.Storage.Type != "local" {
		t.Errorf("Storage.Type 应为 local，实际 %q", cfg.Storage.Type)
	}
	if cfg.Storage.Local.BasePath != "uploads" {
		t.Errorf("Storage.Local.BasePath 应为 uploads，实际 %q", cfg.Storage.Local.BasePath)
	}
	if cfg.Storage.Local.BaseURL != "uploads" {
		t.Errorf("Storage.Local.BaseURL 应为 uploads，实际 %q", cfg.Storage.Local.BaseURL)
	}
	if cfg.DB.Driver != "sqlite" {
		t.Errorf("DB.Driver 应为 sqlite，实际 %q", cfg.DB.Driver)
	}
	if cfg.Captcha.Option != "email" {
		t.Errorf("Captcha.Option 应为 email，实际 %q", cfg.Captcha.Option)
	}
	if cfg.Backup.Cron != "0 0 4 * * 1" {
		t.Errorf("Backup.Cron 应为 0 0 4 * * 1，实际 %q", cfg.Backup.Cron)
	}
}

// TestNormalizeConfigKeepsExplicitValues 已显式配置的值不应被覆盖
func TestNormalizeConfigKeepsExplicitValues(t *testing.T) {
	cfg := &app.Config{
		Server: app.ServerConfig{Host: "0.0.0.0", Port: 9000},
		App:    app.AppConfig{Env: "production"},
		JWT:    jwt.JwtConfig{Secret: "a-very-long-random-secret-12345"},
		DB:     database.DatabaseConfig{Driver: "mysql"},
	}
	normalizeConfig(cfg)

	if cfg.Server.Host != "0.0.0.0" || cfg.Server.Port != 9000 {
		t.Errorf("显式配置被覆盖: %s:%d", cfg.Server.Host, cfg.Server.Port)
	}
	if cfg.App.Env != "production" {
		t.Errorf("App.Env 被覆盖为 %q", cfg.App.Env)
	}
	if cfg.JWT.Secret != "a-very-long-random-secret-12345" {
		t.Errorf("JWT.Secret 被覆盖")
	}
	if cfg.DB.Driver != "mysql" {
		t.Errorf("DB.Driver 被覆盖为 %q", cfg.DB.Driver)
	}
}
