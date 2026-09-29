package service

import (
	"context"
	"strings"

	"blog/internal/dto"
	"blog/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MusicService 音乐相关业务
// 音频文件本体存在 files 表：删除/替换音乐时同步删除关联文件记录与存储对象（避免孤儿文件）
type MusicService struct {
	*Service
}

// NewMusicService 创建音乐服务
func NewMusicService(base *Service) *MusicService {
	return &MusicService{Service: base}
}

// isAudioMime 判断 MIME 是否为音频。
// 注意：m4a（MP4 容器）会被 http.DetectContentType 识别为 video/mp4，故一并放行。
func isAudioMime(mime string) bool {
	return strings.HasPrefix(mime, "audio/") || mime == "video/mp4"
}

// ListMusics 公开音乐列表（仅启用，排序权重升序），拼装完整访问 URL 供浏览器直连
func (s *MusicService) ListMusics() ([]dto.PublicMusic, *BizError) {
	var musics []model.Music
	if err := s.DB().Model(&model.Music{}).
		Where("status = ?", 1).
		Order("sort_order ASC").Order("id ASC").
		Find(&musics).Error; err != nil {
		return nil, NewServerError("查询音乐失败")
	}

	urls, err := s.fileURLMap(musicFileIDs(musics))
	if err != nil {
		return nil, err
	}
	out := make([]dto.PublicMusic, 0, len(musics))
	for _, m := range musics {
		out = append(out, dto.PublicMusic{
			ID:     m.ID,
			Name:   m.Name,
			Artist: m.Artist,
			Cover:  m.Cover,
			URL:    urls[m.FileID],
			Lrc:    m.Lrc,
		})
	}
	return out, nil
}

// ListAdminMusics 管理端音乐列表（分页，含禁用，可按歌名/歌手搜索）
func (s *MusicService) ListAdminMusics(req *dto.MusicQuery) ([]model.Music, int64, *BizError) {
	query := s.DB().Model(&model.Music{})
	if strings.TrimSpace(req.Keyword) != "" {
		kw := "%" + strings.TrimSpace(req.Keyword) + "%"
		query = query.Where("name LIKE ? OR artist LIKE ?", kw, kw)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询音乐失败")
	}

	var musics []model.Music
	if err := query.Order("sort_order ASC").Order("id ASC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&musics).Error; err != nil {
		return nil, 0, NewServerError("查询音乐失败")
	}

	// 拼装音频完整访问 URL
	urls, err := s.fileURLMap(musicFileIDs(musics))
	if err != nil {
		return nil, 0, err
	}
	for i := range musics {
		musics[i].URL = urls[musics[i].FileID]
	}
	return musics, total, nil
}

// GetMusic 管理端音乐详情
func (s *MusicService) GetMusic(id uint) (*model.Music, *BizError) {
	var music model.Music
	if err := s.DB().First(&music, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("音乐不存在")
		}
		return nil, NewServerError("查询音乐失败")
	}
	urls, err := s.fileURLMap([]uint{music.FileID})
	if err != nil {
		return nil, err
	}
	music.URL = urls[music.FileID]
	return &music, nil
}

// CreateMusic 创建音乐（name、file_id 必填，file_id 必须指向已上传的音频文件）
func (s *MusicService) CreateMusic(req *dto.MusicRequest) (*model.Music, *BizError) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, NewParamError("歌名不能为空")
	}
	if req.FileID == nil || *req.FileID == 0 {
		return nil, NewParamError("请选择音频文件")
	}
	if err := s.checkAudioFile(*req.FileID); err != nil {
		return nil, err
	}

	music := model.Music{
		Name:   name,
		FileID: *req.FileID,
		Status: 1,
	}
	if req.Artist != nil {
		music.Artist = *req.Artist
	}
	if req.Cover != nil {
		music.Cover = *req.Cover
	}
	if req.SortOrder != nil {
		music.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		music.Status = *req.Status
	}
	if req.Lrc != nil {
		music.Lrc = *req.Lrc
	}

	if err := s.DB().Create(&music).Error; err != nil {
		return nil, NewServerError("创建音乐失败")
	}
	s.Log().Info("创建音乐", zap.Uint("id", music.ID), zap.String("name", music.Name), zap.Uint("file_id", music.FileID))
	return &music, nil
}

// UpdateMusic 更新音乐（仅更新提供的字段；更换 file_id 时同步删除旧音频文件记录与存储对象）
func (s *MusicService) UpdateMusic(id uint, req *dto.MusicRequest) (*model.Music, *BizError) {
	var music model.Music
	if err := s.DB().First(&music, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("音乐不存在")
		}
		return nil, NewServerError("查询音乐失败")
	}

	updates := map[string]any{}
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if req.Artist != nil {
		updates["artist"] = *req.Artist
	}
	if req.Cover != nil {
		updates["cover"] = *req.Cover
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Lrc != nil {
		updates["lrc"] = *req.Lrc
	}

	// 更换音频文件：校验新文件，事务内删除旧文件记录，提交后清理旧文件存储对象
	var oldFile *model.File
	if req.FileID != nil && *req.FileID > 0 && *req.FileID != music.FileID {
		if err := s.checkAudioFile(*req.FileID); err != nil {
			return nil, err
		}
		updates["file_id"] = *req.FileID
		if music.FileID > 0 {
			var old model.File
			if err := s.DB().First(&old, music.FileID).Error; err == nil {
				oldFile = &old
			}
		}
	}

	if len(updates) > 0 {
		err := s.DB().Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&music).Updates(updates).Error; err != nil {
				return err
			}
			if oldFile != nil {
				if err := tx.Delete(oldFile).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, NewServerError("更新音乐失败")
		}
		// 事务提交后清理旧音频文件存储对象（失败仅记录日志，不阻塞）
		if oldFile != nil {
			s.cleanupStoredFiles(oldFile)
		}
	}

	if err := s.DB().First(&music, id).Error; err != nil {
		return nil, NewServerError("查询音乐失败")
	}
	s.Log().Info("更新音乐", zap.Uint("id", music.ID), zap.String("name", music.Name))
	return &music, nil
}

// DeleteMusic 删除音乐并同步删除关联的 files 记录与存储对象（软删除音乐行）
func (s *MusicService) DeleteMusic(id uint) *BizError {
	var music model.Music
	if err := s.DB().First(&music, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("音乐不存在")
		}
		return NewServerError("查询音乐失败")
	}

	// 事务：音乐行 + 关联文件记录一起删除，保证一致性；存储对象清理放在提交后（失败仅日志）
	var fileToClean *model.File
	err := s.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&music).Error; err != nil {
			return err
		}
		if music.FileID == 0 {
			return nil
		}
		var file model.File
		if err := tx.First(&file, music.FileID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil // 文件记录已不存在，无需清理
			}
			return err
		}
		if err := tx.Delete(&file).Error; err != nil {
			return err
		}
		fileToClean = &file
		return nil
	})
	if err != nil {
		return NewServerError("删除音乐失败")
	}
	if fileToClean != nil {
		s.cleanupStoredFiles(fileToClean)
	}
	s.Log().Info("删除音乐", zap.Uint("id", music.ID), zap.String("name", music.Name), zap.Uint("file_id", music.FileID))
	return nil
}

// checkAudioFile 校验文件记录存在且为音频类型
func (s *MusicService) checkAudioFile(fileID uint) *BizError {
	var file model.File
	if err := s.DB().First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewParamError("音频文件不存在")
		}
		return NewServerError("查询文件失败")
	}
	if !isAudioMime(file.MimeType) {
		return NewParamError("关联文件不是音频文件")
	}
	return nil
}

// cleanupStoredFiles 清理文件对应的存储对象（主文件与缩略图；失败仅记录日志）
func (s *MusicService) cleanupStoredFiles(file *model.File) {
	ctx := context.Background()
	for _, p := range []string{file.Path, file.Thumbnail} {
		if p == "" {
			continue
		}
		if err := s.Storage().Delete(ctx, p); err != nil {
			s.Log().Warn("清理音乐关联文件失败", zap.Error(err), zap.String("path", p))
		}
	}
}

// musicFileIDs 提取音乐列表中的关联文件 ID（去重）
func musicFileIDs(musics []model.Music) []uint {
	seen := map[uint]bool{}
	ids := make([]uint, 0, len(musics))
	for _, m := range musics {
		if m.FileID == 0 || seen[m.FileID] {
			continue
		}
		seen[m.FileID] = true
		ids = append(ids, m.FileID)
	}
	return ids
}

// fileURLMap 批量查询文件记录并拼装完整访问 URL（map[fileID]url）
func (s *MusicService) fileURLMap(fileIDs []uint) (map[uint]string, *BizError) {
	result := make(map[uint]string)
	if len(fileIDs) == 0 {
		return result, nil
	}
	var files []model.File
	if err := s.DB().Where("id IN ?", fileIDs).Find(&files).Error; err != nil {
		return nil, NewServerError("查询音频文件失败")
	}
	ctx := context.Background()
	for _, f := range files {
		result[f.ID] = s.Storage().GetURL(ctx, f.Path)
	}
	return result, nil
}
