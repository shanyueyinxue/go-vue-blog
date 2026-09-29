package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type OSSStorage struct {
	client  *minio.Client
	bucket  string
	baseURL string
}

var _ Storage = &OSSStorage{}

func NewOSSStorage(cfg OSSStorageConfig) (Storage, error) {
	useSSL := cfg.UseHTTPS
	if !useSSL && strings.HasPrefix(cfg.Endpoint, "https://") {
		useSSL = true
	}
	endpoint := strings.TrimPrefix(cfg.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 OSS 客户端失败: %w", err)
	}

	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("检查 Bucket 失败: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("Bucket 不存在: %s", cfg.Bucket)
	}

	return &OSSStorage{
		client:  client,
		bucket:  cfg.Bucket,
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
	}, nil
}

func (s *OSSStorage) Save(ctx context.Context, originalName string, data []byte, folder string) (storedPath string, err error) {
	key := generateKey(originalName, folder)
	reader := bytes.NewReader(data)

	_, err = s.client.PutObject(ctx, s.bucket, key, reader, int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("OSS 上传失败: %w", err)
	}
	return key, nil
}

func (s *OSSStorage) SaveStream(ctx context.Context, originalName string, reader io.Reader, folder string) (storedPath string, err error) {
	key := generateKey(originalName, folder)

	_, err = s.client.PutObject(ctx, s.bucket, key, reader, -1, minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("OSS 上传失败: %w", err)
	}
	return key, nil
}

func (s *OSSStorage) Delete(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}
	err := s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("OSS 删除失败: %w", err)
	}
	return nil
}

func (s *OSSStorage) GetURL(ctx context.Context, path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if s.baseURL != "" {
		return s.baseURL + "/" + path
	}
	endpoint := strings.TrimRight(s.client.EndpointURL().String(), "/")
	return endpoint + "/" + s.bucket + "/" + path
}

func (s *OSSStorage) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("OSS 获取文件失败: %w", err)
	}
	return obj, nil
}
