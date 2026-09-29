package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorageConfig struct {
	BasePath string `mapstructure:"basePath"`
	BaseURL  string `mapstructure:"baseUrl"`
}

type LocalStorage struct {
	basePath string
	baseURL  string
}

var _ Storage = &LocalStorage{}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// Save 将字节数据保存为文件，返回存储的相对路径
func (s *LocalStorage) Save(ctx context.Context, originalName string, data []byte, folder string) (storedPath string, err error) {
	if err := s.validateFolder(folder); err != nil {
		return "", err
	}
	// 确保基础目录存在
	if err := os.MkdirAll(filepath.Join(s.basePath, folder), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	rawPath := generateKey(originalName, folder)
	fullPath := filepath.Join(s.basePath, rawPath)

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 返回相对路径（仅文件名，上层可结合 basePath 或 baseURL 使用）
	return rawPath, nil
}

// SaveStream 将 io.Reader 流保存为文件，返回存储的相对路径
func (s *LocalStorage) SaveStream(ctx context.Context, originalName string, reader io.Reader, folder string) (storedPath string, err error) {
	if err := s.validateFolder(folder); err != nil {
		return "", err
	}
	// 确保基础目录存在
	if err := os.MkdirAll(filepath.Join(s.basePath, folder), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	rawPath := generateKey(originalName, folder)
	fullPath := filepath.Join(s.basePath, rawPath)

	dest, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dest.Close()

	// 将 reader 内容复制到目标文件
	if _, err := io.Copy(dest, reader); err != nil {
		// 复制失败时清理已创建的不完整文件
		os.Remove(fullPath)
		return "", fmt.Errorf("复制文件流失败: %w", err)
	}

	return rawPath, nil
}

// validateFolder 校验上传目录，防止路径穿越：
// 目录必须为相对路径且不能包含 .. 组件、绝对路径或盘符
func (s *LocalStorage) validateFolder(folder string) error {
	if folder == "" {
		return nil
	}
	clean := filepath.Clean(folder)
	if clean == ".." ||
		strings.HasPrefix(clean, ".."+string(os.PathSeparator)) ||
		filepath.IsAbs(clean) ||
		strings.Contains(clean, ":") ||
		strings.ContainsRune(clean, 0) {
		return fmt.Errorf("非法目录: %s", folder)
	}
	return nil
}

// Delete 删除指定路径的文件（路径为 Save 方法返回的相对路径）
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	if path == "" {
		return nil // 空路径无需删除
	}

	fullPath := filepath.Join(s.basePath, path)

	// 安全防护：防止路径遍历攻击，确保最终路径仍在 basePath 内
	cleanBase := filepath.Clean(s.basePath)
	cleanFull := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanFull, cleanBase+string(os.PathSeparator)) && cleanFull != cleanBase {
		return fmt.Errorf("非法路径: %s", path)
	}

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// GetURL 根据存储的相对路径拼接完整的可访问 URL
func (s *LocalStorage) GetURL(ctx context.Context, path string) string {
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimLeft(path, "/")
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	if s.baseURL == "" {
		return "/" + path
	}

	// baseURL 可能是完整域名（CDN）或本地路径前缀（如 "uploads" / "/uploads"）
	if strings.HasPrefix(s.baseURL, "http://") || strings.HasPrefix(s.baseURL, "https://") {
		return strings.TrimRight(s.baseURL, "/") + "/" + path
	}
	// 本地路径前缀必须补前导斜杠，否则在子路由页面（如 /admin/files）会被解析成相对路径
	base := strings.Trim(s.baseURL, "/")
	if base == "" {
		return "/" + path
	}
	return "/" + base + "/" + path
}

func (s *LocalStorage) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)

	// 安全防护：防止路径遍历攻击，确保最终路径仍在 basePath 内
	cleanBase := filepath.Clean(s.basePath)
	cleanFull := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanFull, cleanBase+string(os.PathSeparator)) && cleanFull != cleanBase {
		return nil, fmt.Errorf("非法路径: %s", path)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	return file, nil
}
