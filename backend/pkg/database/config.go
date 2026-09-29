package database

import (
	"time"
)

// Config 数据库配置结构体
type DatabaseConfig struct {
	Driver   string       `mapstructure:"driver"` // mysql, postgres 或 sqlite
	MySQL    MysqlConfig  `mapstructure:"mysql"`
	Postgres PgConfig     `mapstructure:"postgres"`
	SQLite   SqliteConfig `mapstructure:"sqlite"`
}

// MysqlConfig MySQL数据库配置
type MysqlConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parseTime"`
	Loc             string        `mapstructure:"loc"` // 时区，如Asia/Shanghai
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}

// PgConfig PostgreSQL数据库配置
type PgConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"sslmode"` // disable, allow, prefer, require, verify-ca, verify-full
	Timezone        string        `mapstructure:"timezone"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}

// SqliteConfig SQLite数据库配置
type SqliteConfig struct {
	Path            string        `mapstructure:"path"`
	ForeignKeys     bool          `mapstructure:"foreign_keys"` // 是否启用外键约束
	MaxOpenConns    int           `mapstructure:"maxOpenConns"` // SQLite通常设置为1
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}
