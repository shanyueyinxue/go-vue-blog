package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"
)

// generateKey 生成存储相对路径：<纳秒时间戳>_<24位随机十六进制><原扩展名>。
// 随机部分使用 crypto/rand（不可预测），保证文件名（含备份文件）无法被遍历/撞库猜中；
// 时间戳前缀用于保持同一目录内的可排序性。
func generateKey(originalName, folder string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano()
	fileName := fmt.Sprintf("%d_%s%s", timestamp, randomHex(12), ext)
	if folder != "" {
		return folder + "/" + fileName
	}
	return fileName
}

// randomHex 生成 n 字节的随机十六进制字符串（crypto/rand）。
// crypto/rand 极端失败时回退为纳秒时间戳，保证不 panic（此时文件名仍含时间戳，可区分）。
func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
