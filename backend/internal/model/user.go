package model

import (
	"time"

	"gorm.io/gorm"
)

// User 管理员用户模型（对应 docs/database.md 3.1 users 表）
// 本项目为单用户系统：仅管理员一个有效用户（is_master=1）
type User struct {
	ID          uint           `gorm:"primarykey" json:"id"`                // 主键
	Username    string         `gorm:"size:50;uniqueIndex" json:"username"` // 登录名（唯一）
	Password    string         `gorm:"size:255" json:"-"`                   // bcrypt 加密后的密码（不对外输出）
	Email       string         `gorm:"size:100" json:"email"`               // 邮箱（可空）
	Phone       string         `gorm:"size:20" json:"phone"`                // 手机号（可空）
	Avatar      string         `gorm:"size:255" json:"avatar"`              // 头像地址
	PwdVersion  uint           `gorm:"not null;default:0" json:"-"`         // 密码版本号，修改密码后递增，用于使旧 Token 失效
	IsMaster    int            `gorm:"default:0" json:"is_master"`          // 是否超级管理员：1 是 / 0 否
	Role        string         `gorm:"size:20;default:admin" json:"role"`   // 角色，如 admin
	Status      int            `gorm:"default:1" json:"status"`             // 状态：1 启用 / 0 禁用
	LastIP      string         `gorm:"size:50" json:"last_ip"`              // 最近登录 IP
	LastLoginAt *time.Time     `json:"last_login_at"`                       // 最近登录时间
	CreatedAt   time.Time      `json:"created_at"`                          // 创建时间
	UpdatedAt   time.Time      `json:"updated_at"`                          // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                      // 软删除标记
}
