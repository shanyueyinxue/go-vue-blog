package service

import (
	"bytes"
	"context"
	"mime/multipart"
	"path/filepath"
	"strings"

	"blog/internal/dto"
	"blog/internal/model"
	"blog/pkg/storage"
	"blog/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 文件上传大小上限（50MB）
const maxUploadSize = 50 << 20

// 允许上传的 MIME 类型白名单（图片/文档/视频）
var allowedMimeTypes = map[string]bool{
	// 图片
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	// "image/svg+xml": true,
	// 文档
	"application/pdf":    true,
	"text/plain":         true,
	"text/markdown":      true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	// 视频
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,
	"video/x-msvideo": true,
	// 音频（含 http.DetectContentType 的变体，如 wav/wave）
	"audio/mpeg":   true,
	"audio/mp3":    true,
	"audio/x-mpeg": true,
	"audio/ogg":    true,
	"audio/opus":   true,
	"audio/wav":    true,
	"audio/wave":   true,
	"audio/x-wav":  true,
	"audio/flac":   true,
	"audio/aac":    true,
	"audio/mp4":    true,
	"audio/x-m4a":  true,
	"audio/webm":   true,
}

// officeZipExtMimes 基于 zip 容器的 Office 文档扩展名 → MIME 映射。
// http.DetectContentType 对 docx/xlsx 只能识别出 application/zip，
// 需要按扩展名映射回真实 MIME 才能通过白名单校验。
var officeZipExtMimes = map[string]string{
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
}

// mimeByExtension 根据文件扩展名返回 Office zip 容器的真实 MIME；未知扩展名返回空串
func mimeByExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return officeZipExtMimes[ext]
}

// textSafeExts 文本类文件允许保留的安全扩展名（小写、含点）。
// 文本内容（含 Markdown）经 http.DetectContentType 检测后统一为 text/plain，
// 与 HTML/JS/SVG 等可执行内容无法通过魔数区分。若允许携带可执行扩展名
// （.html/.js/.svg/.xml 等），文件经 /uploads 静态服务会以 text/html 等可执行类型
// 输出，形成同源存储型 XSS。因此文本类 MIME 必须校验扩展名，不在白名单内拒绝保存。
var textSafeExts = map[string]bool{
	".txt": true, ".text": true,
	".md": true, ".markdown": true, ".mdown": true,
	".log": true, ".csv": true, ".conf": true, ".ini": true,
	".json": true, ".yaml": true, ".yml": true, ".toml": true,
}

// textMimeExts 文本类 MIME → 允许扩展名白名单映射。
// 命中该映射的 MIME 在通过 allowedMimeTypes 白名单后，还会校验扩展名
// （见 Upload）：扩展名不在允许列表内直接拒绝保存，防止文本文件伪装成可执行文件。
var textMimeExts = map[string]map[string]bool{
	"text/plain":    textSafeExts,
	"text/markdown": textSafeExts,
}

// extAllowed 校验文件扩展名是否属于该 MIME 允许的白名单。
// 返回 true 表示允许（该 MIME 不在 textMimeExts 中，或扩展名在白名单内）。
func extAllowed(mime, filename string) bool {
	exts, ok := textMimeExts[mime]
	if !ok {
		return true
	}
	return exts[strings.ToLower(filepath.Ext(filename))]
}

// folderByMime 按 MIME 类型返回默认存储子目录
func folderByMime(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "images"
	case strings.HasPrefix(mime, "video/"):
		return "videos"
	case strings.HasPrefix(mime, "audio/"):
		return "audios"
	default:
		return "files"
	}
}

// FileService 文件上传/列表/删除相关业务
type FileService struct {
	*Service
}

// NewFileService 创建文件服务
func NewFileService(base *Service) *FileService {
	return &FileService{Service: base}
}

// Upload 上传文件：检测大小与真实 MIME 类型、保存到存储并记录数据库
func (s *FileService) Upload(fileHeader *multipart.FileHeader, dir string, userID uint) (*model.File, error) {
	// 大小限制
	if fileHeader.Size > maxUploadSize {
		return nil, NewParamError("文件大小不能超过 50MB")
	}

	// 读取文件内容并检测真实 MIME 类型（防伪装）
	fileInfo, err := storage.GetFileInfo(fileHeader)
	if err != nil {
		return nil, NewParamError("读取文件失败")
	}
	realMime, err := utils.DetectRealMime(fileHeader)
	if err != nil {
		realMime = fileInfo.MimeType
	}
	// docx/xlsx 等是 zip 容器，检测结果会是 application/zip；
	// 按扩展名映射回真实 MIME，保证白名单匹配且入库类型正确
	if realMime == "application/zip" {
		if m := mimeByExtension(fileHeader.Filename); m != "" {
			realMime = m
		}
	}
	// 类型白名单校验
	if !allowedMimeTypes[realMime] {
		return nil, NewParamError("不支持的文件类型: " + realMime)
	}
	// 文本类 MIME 扩展名白名单校验：text/plain、text/markdown 的内容与 HTML/JS 无法通过
	// 魔数区分，必须匹配允许的安全扩展名（如 .txt/.md），否则拒绝保存，防止文本文件
	// 携带 .html/.js/.svg 等可执行扩展名被 /uploads 以可执行类型输出（同源存储型 XSS）。
	if !extAllowed(realMime, fileHeader.Filename) {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		return nil, NewParamError("文本文件扩展名不允许: " + ext)
	}

	// 未显式指定目录时，按 MIME 类型自动分目录（图片→images、视频→videos、音频→audios、其它→files）
	if dir == "" {
		dir = folderByMime(realMime)
	}

	// 保存到存储（local / oss）
	ctx := context.Background()
	storedPath, err := s.Storage().Save(ctx, fileHeader.Filename, fileInfo.Data, dir)
	if err != nil {
		return nil, NewServerError("文件保存失败")
	}

	// 记录文件信息到数据库
	file := model.File{
		UserID:   &userID,
		Filename: fileHeader.Filename,
		Path:     storedPath,
		MimeType: realMime,
		Size:     fileHeader.Size,
	}
	if err := s.DB().Create(&file).Error; err != nil {
		// 数据库记录失败时清理已保存的文件
		_ = s.Storage().Delete(ctx, storedPath)
		return nil, NewServerError("文件记录保存失败")
	}

	// 仅图片文件异步生成缩略图
	if isImageMime(realMime) {
		go s.generateThumbnail(file.ID, realMime, fileInfo.Data)
	}

	s.Log().Info("上传文件", zap.Uint("id", file.ID), zap.String("filename", file.Filename), zap.String("path", file.Path), zap.Int64("size", file.Size))
	return &file, nil
}

func isImageMime(mimeType string) bool {
	return mimeType == "image/jpeg" || mimeType == "image/png" || mimeType == "image/gif" || mimeType == "image/webp"
}

func (s *FileService) generateThumbnail(fileID uint, mimeType string, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			s.Log().Warn("生成缩略图发生异常", zap.Any("recover", r), zap.Uint("file_id", fileID))
		}
	}()

	thumbnailData, err := utils.GenerateThumbnail(bytes.NewReader(data), 320, mimeType)
	if err != nil {
		s.Log().Warn("生成缩略图失败", zap.Error(err), zap.Uint("file_id", fileID))
		return
	}

	var file model.File
	if err := s.DB().First(&file, fileID).Error; err != nil {
		return
	}

	ctx := context.Background()
	thumbnailPath, err := s.Storage().Save(ctx, "thumb_"+file.Filename, thumbnailData, "thumbs")
	if err != nil {
		s.Log().Warn("保存缩略图失败", zap.Error(err), zap.Uint("file_id", fileID))
		return
	}

	if err := s.DB().Model(&file).Update("thumbnail", thumbnailPath).Error; err != nil {
		s.Log().Warn("更新缩略图字段失败", zap.Error(err), zap.Uint("file_id", fileID))
		_ = s.Storage().Delete(ctx, thumbnailPath)
	}
}

// GetURL 获取文件的完整访问地址
func (s *FileService) GetURL(ctx context.Context, path string) string {
	return s.Storage().GetURL(ctx, path)
}

// ListFiles 文件列表（支持按文件名搜索、MIME 类型筛选）
func (s *FileService) ListFiles(req *dto.FileQuery) ([]model.File, int64, *BizError) {
	query := s.DB().Model(&model.File{})
	if req.MimeType != "" {
		query = query.Where("mime_type LIKE ?", req.MimeType+"%")
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("filename LIKE ?", kw)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询文件失败")
	}

	var files []model.File
	if err := query.Order("id DESC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&files).Error; err != nil {
		return nil, 0, NewServerError("查询文件失败")
	}

	// 拼装完整可访问 URL（local / oss 通用，前端无需再按存储类型拼接路径）
	ctx := context.Background()
	for i := range files {
		files[i].URL = s.Storage().GetURL(ctx, files[i].Path)
		if files[i].Thumbnail != "" {
			files[i].ThumbnailURL = s.Storage().GetURL(ctx, files[i].Thumbnail)
		}
	}
	return files, total, nil
}

// DeleteFile 删除文件（物理删除）：清理本地/OSS 中的实际文件并删除数据库记录
func (s *FileService) DeleteFile(id uint) *BizError {
	var file model.File
	if err := s.DB().First(&file, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("文件不存在")
		}
		return NewServerError("查询文件失败")
	}

	// 删除存储中的文件（失败仅记录日志，不阻塞删除记录）
	ctx := context.Background()
	if err := s.Storage().Delete(ctx, file.Path); err != nil {
		s.Log().Warn("删除存储文件失败", zap.Error(err), zap.String("path", file.Path))
	}
	// 一并清理缩略图，避免遗留孤儿文件
	if file.Thumbnail != "" {
		if err := s.Storage().Delete(ctx, file.Thumbnail); err != nil {
			s.Log().Warn("删除缩略图失败", zap.Error(err), zap.String("path", file.Thumbnail))
		}
	}
	// 删除数据库记录
	if err := s.DB().Delete(&file).Error; err != nil {
		return NewServerError("删除文件记录失败")
	}
	s.Log().Info("删除文件", zap.Uint("id", file.ID), zap.String("filename", file.Filename))
	return nil
}
