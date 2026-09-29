package utils

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

// detectRealMime 读取文件前512字节检测真实 MIME 类型
func DetectRealMime(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	// 使用 http.DetectContentType 检测真实类型
	realMime := http.DetectContentType(buf[:n])
	// 清理可能包含的参数（如 charset=utf-8）
	if parsedMime, _, err := mime.ParseMediaType(realMime); err == nil {
		realMime = parsedMime
	}
	return realMime, nil
}

// SafeGetImageDimensions 安全获取图片尺寸（防止压缩炸弹）
// 注意：仅当需要解析尺寸时调用，且必须限制最大宽高
func SafeGetImageDimensions(reader io.ReadSeeker, mimeType string, option ...map[string]int) (int, int, error) {
	// 重置读取位置
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return 0, 0, err
	}
	// 解析选项
	options := map[string]int{
		"max_head_size": 5 * 1024 * 1024, // 限制头部大小（5MB）
		"max_width":     15000,
		"max_height":    15000,
	}
	if len(option) > 0 && option[0] != nil {
		for k, v := range option[0] {
			options[k] = v
		}
	}
	// 校验并修正选项（避免零值覆盖）
	maxHead := options["max_head_size"]
	if maxHead <= 0 {
		maxHead = 5 * 1024 * 1024
	}
	maxW := options["max_width"]
	if maxW <= 0 {
		maxW = 15000
	}
	maxH := options["max_height"]
	if maxH <= 0 {
		maxH = 15000
	}

	limitedReader := io.LimitReader(reader, int64(maxHead))

	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		cfg, _, err := image.DecodeConfig(limitedReader)
		if err != nil {
			return 0, 0, fmt.Errorf("decode config: %w", err)
		}
		if cfg.Width > maxW || cfg.Height > maxH {
			return 0, 0, fmt.Errorf("image dimensions %dx%d exceed limit %dx%d", cfg.Width, cfg.Height, maxW, maxH)
		}
		return cfg.Width, cfg.Height, nil
	default:
		return 0, 0, fmt.Errorf("unsupported image type: %s", mimeType)
	}
}

// GenerateThumbnail 生成缩略图，限制最大尺寸（宽、高均不超过 maxSize），保持原始宽高比
// r: 原始图片数据的读取器（调用方负责在读完后关闭）
// maxSize: 缩略图的最大宽度和最大高度（必须为正整数）
// mimeType: 原始图片的 MIME 类型，用于解码和编码
// 返回: 缩略图数据（格式尽量与原始格式相同）和错误信息

func GenerateThumbnail(r io.Reader, maxSize int, mimeType string) ([]byte, error) {
	if maxSize <= 0 {
		return nil, errors.New("maxSize must be positive integer")
	}

	// 1. 解码图片
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	// 2. 生成缩略图
	//    imaging.Fit 会保持宽高比，并将图片缩放至 maxSize x maxSize 的范围内。
	//    使用 imaging.Lanczos 滤镜，这是该库提供的高质量滤镜之一。
	thumbnail := imaging.Fit(img, maxSize, maxSize, imaging.Lanczos)

	// 3. 编码输出
	buf := new(bytes.Buffer)
	// 根据原MIME类型选择编码格式，提升兼容性；未知格式默认输出JPEG
	switch mimeType {
	case "image/png":
		err = imaging.Encode(buf, thumbnail, imaging.PNG)
	case "image/gif":
		err = imaging.Encode(buf, thumbnail, imaging.GIF)
	case "image/jpeg", "image/jpg", "":
		fallthrough
	default:
		// JPEG编码时可以指定质量，默认95是高质量
		err = imaging.Encode(buf, thumbnail, imaging.JPEG, imaging.JPEGQuality(95))
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
