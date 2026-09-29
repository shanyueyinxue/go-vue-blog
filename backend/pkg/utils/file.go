package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func ReadFile(file string) (name string, b []byte, err error) {
	name = filepath.Base(file)
	b, err = os.ReadFile(file)
	return
}

// SplitPathAndName 分离文件路径和文件名；将路径中 \\ 改为 /，以便统一处理
func SplitPathAndName(fullPath string) (dir string, name string) {
	fullPath = filepath.Clean(fullPath)
	name = filepath.Base(fullPath)
	dir = filepath.Dir(fullPath)

	hasSeparator := strings.ContainsAny(fullPath, "/\\")

	// 如果 dir == "." 且原始输入不含分隔符，说明输入只有文件名
	// 例如 "file.txt" → Dir 返回 "."，但没有分隔符 → 返回路径为空
	if dir == "." && !hasSeparator {
		return "", name
	}
	dir = strings.ReplaceAll(dir, "\\", "/")

	// 其他情况：返回真实的目录路径（可能是 "."、"/"、"C:\" 等）
	return dir, name
}
