package utils

import "os"

// 判断目录是否存在
func IsDirExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}
