package model

import (
	"time"

	"gorm.io/gorm"
)

// Music 音乐模型（对应 docs/database.md 3.11 musics 表）
// 音频文件本体保存在 files 表（与 File.FileID 关联），删除音乐时同步删除关联文件记录与存储对象
type Music struct {
	ID        uint           `gorm:"primarykey" json:"id"`                   // 主键
	Name      string         `gorm:"size:100;not null" json:"name"`          // 歌名（APlayer name）
	Artist    string         `gorm:"size:100" json:"artist"`                 // 歌手（APlayer artist）
	Cover     string         `gorm:"size:500" json:"cover"`                  // 封面图 URL（可空）
	FileID    uint           `gorm:"index:idx_musics_status_sort,priority:4" json:"file_id"` // 关联 files 表 ID（音频文件记录）
	SortOrder int            `gorm:"default:0;index:idx_musics_status_sort,priority:3" json:"sort_order"` // 排序权重
	Status    int            `gorm:"index:idx_musics_status_sort,priority:1" json:"status"` // 状态：1 启用 / 0 禁用（服务层默认 1）
	Lrc       string         `gorm:"type:text" json:"lrc"`                   // 歌词文本（预留，后续 LRC 功能）
	CreatedAt time.Time      `json:"created_at"`                             // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`                             // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                         // 软删除标记

	// 以下为查询时由存储层拼装的完整可访问地址（非数据库字段）
	URL string `gorm:"-" json:"url,omitempty"` // 音频完整访问 URL
}
