package model

import "time"

// File 文件上传记录模型（对应 docs/database.md 3.9 files 表）
// 删除时底层文件已同步物理清理，软删除无恢复价值，故采用物理删除
type File struct {
	ID        uint      `gorm:"primarykey" json:"id"`      // 主键
	UserID    *uint     `gorm:"index" json:"user_id"`      // 上传者 ID（单用户系统可为空，为多用户扩展预留）
	Filename  string    `gorm:"size:255" json:"filename"`  // 原始文件名
	Path      string    `gorm:"size:500" json:"path"`      // 存储路径（相对路径）
	Thumbnail string    `gorm:"size:500" json:"thumbnail"` // 缩略图路径（仅图片文件）
	MimeType  string    `gorm:"size:100" json:"mime_type"` // MIME 类型
	Size      int64     `json:"size"`                      // 文件大小（字节）
	CreatedAt time.Time `json:"created_at"`                // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                // 更新时间

	// 以下为查询时由存储层拼装的完整可访问地址（非数据库字段）
	URL          string `gorm:"-" json:"url,omitempty"`           // 完整访问 URL
	ThumbnailURL string `gorm:"-" json:"thumbnail_url,omitempty"` // 缩略图完整访问 URL
}
