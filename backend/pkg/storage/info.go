package storage

import (
	"io"
	"mime/multipart"
)

type FileInfo struct {
	OriginalName string
	Size         int64
	MimeType     string
	Data         []byte
	Header       *multipart.FileHeader
}

func GetFileInfo(file *multipart.FileHeader) (*FileInfo, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// io.ReadAll 读满全部内容，避免单次 Read 可能短读
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return &FileInfo{
		OriginalName: file.Filename,
		Size:         file.Size,
		MimeType:     mimeType,
		Data:         data,
		Header:       file,
	}, nil
}
