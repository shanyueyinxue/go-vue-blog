package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
)

type Storage interface {
	Save(ctx context.Context, originalName string, data []byte, folder string) (storedPath string, err error)
	SaveStream(ctx context.Context, originalName string, reader io.Reader, folder string) (storedPath string, err error)
	Delete(ctx context.Context, path string) error
	GetURL(ctx context.Context, path string) string
	GetFile(ctx context.Context, path string) (io.ReadCloser, error)
}

var ErrUnsupportedStorageType = errors.New("unsupported storage type")

type StorageConfig struct {
	Type  string             `mapstructure:"type"` // local / oss
	Local LocalStorageConfig `mapstructure:"local"`
	OSS   OSSStorageConfig   `mapstructure:"oss"`
}

type OSSStorageConfig struct {
	AccessKey string `mapstructure:"accessKey"`
	SecretKey string `mapstructure:"secretKey"`
	Endpoint  string `mapstructure:"endpoint"` // S3 endpoint
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`  // S3 region
	BaseURL   string `mapstructure:"baseUrl"` // 自定义域名 / CDN 域名
	UseHTTPS  bool   `mapstructure:"useHttps"`
}

func NewStorage(config StorageConfig) (Storage, error) {
	switch config.Type {
	case "local":
		return NewLocalStorage(config.Local.BasePath, config.Local.BaseURL), nil
	case "oss":
		return NewOSSStorage(config.OSS)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedStorageType, config.Type)
	}
}
