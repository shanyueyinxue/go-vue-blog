package utils

import (
	"fmt"
	"runtime"
	"strings"
)

// IsValidFilePath 判断路径字符串是否合法（可作为文件或目录路径使用）
// 支持 Windows 和 Unix-like 系统（Linux、macOS 等）
func IsValidFilePath(path string) bool {
	if path == "" {
		return false
	}

	// 可选：限制最大长度（避免过长路径）
	if len(path) > 4096 {
		return false
	}

	if runtime.GOOS == "windows" {
		return isValidWindowsPath(path)
	}
	return isValidUnixPath(path)
}

// Unix-like 系统（Linux、macOS 等）的路径合法性检查
func isValidUnixPath(path string) bool {
	// 禁止空字符（NULL）
	if strings.ContainsRune(path, 0) {
		return false
	}

	// 按 '/' 分割路径组件
	components := strings.Split(path, "/")
	for i, comp := range components {
		// 根路径的第一个组件为空（如 "/home" -> ["", "home"]），跳过
		if i == 0 && comp == "" && strings.HasPrefix(path, "/") {
			continue
		}
		// 不允许空组件（如 "//home" 或 "a//b"）
		if comp == "" {
			return false
		}
		// 组件中不能含有 '/'（已被分割，但防御性检查）
		if strings.ContainsRune(comp, '/') {
			return false
		}
		// 文件名长度限制（255 是常见上限）
		if len(comp) > 255 {
			return false
		}
	}
	return true
}

// Windows 路径合法性检查
func isValidWindowsPath(path string) bool {
	// 禁止空字符和控制字符（0x00-0x1F）
	for _, r := range path {
		if r == 0 || (r >= 0x00 && r <= 0x1F) {
			return false
		}
	}

	// 反斜杠和正斜杠都允许作为分隔符，但需要检查是否单独出现（不额外禁止）
	// 注意：冒号 ':' 仅允许出现在驱动器号位置（如 "C:"），其他地方禁止
	if err := checkColonUsage(path); err != nil {
		return false
	}

	// 禁止以下非法字符（Windows 文件名/路径中不允许）
	illegalChars := `<>"|?*`
	if strings.ContainsAny(path, illegalChars) {
		return false
	}

	// 不能以空格或点结尾（某些 Windows API 会拒绝）
	if strings.HasSuffix(path, " ") || strings.HasSuffix(path, ".") {
		return false
	}

	// 保留设备名检查（大小写不敏感，忽略扩展名）
	if isReservedDeviceName(path) {
		return false
	}

	// 分割路径组件（支持 \ 和 / 作为分隔符）
	normalized := strings.ReplaceAll(path, "/", "\\")
	components := strings.Split(normalized, "\\")
	for i, comp := range components {
		// 跳过卷标部分（如 "C:"）
		if i == 0 && strings.HasSuffix(comp, ":") {
			continue
		}
		// 空组件：仅允许根路径后的第一个空（如 "C:\" 会产生空组件），其他情况不允许
		if comp == "" {
			// 允许根目录后的空（例如 "C:\" -> ["C:", ""]）
			if i == 1 && len(components) > 1 && components[0] != "" && strings.HasSuffix(components[0], ":") {
				continue
			}
			return false
		}
		// 组件长度限制（Windows 通常 255）
		if len(comp) > 255 {
			return false
		}
		// 组件不能以空格或点结尾（与整体路径规则一致）
		if strings.HasSuffix(comp, " ") || strings.HasSuffix(comp, ".") {
			return false
		}
	}
	return true
}

// 检查冒号的使用（Windows 只允许在驱动器号后出现一次）
func checkColonUsage(path string) error {
	colonIndex := strings.Index(path, ":")
	if colonIndex == -1 {
		return nil
	}
	// 必须位于第 2 个字符（索引 1），且前面是字母，后面是分隔符或结束
	if colonIndex == 1 {
		if (path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z') {
			// 冒号后必须是分隔符或结尾
			if len(path) == 2 || path[2] == '\\' || path[2] == '/' {
				// 检查是否还有第二个冒号
				if strings.Contains(path[colonIndex+1:], ":") {
					return errInvalidColon
				}
				return nil
			}
		}
	}
	return errInvalidColon
}

var errInvalidColon = fmt.Errorf("invalid colon usage")

// 检查是否为 Windows 保留设备名（如 CON, PRN, AUX, NUL, COM1, LPT1 等）
func isReservedDeviceName(path string) bool {
	// 提取文件名（最后一个分隔符之后的部分）
	normalized := strings.ReplaceAll(path, "/", "\\")
	lastSlash := strings.LastIndex(normalized, "\\")
	fileName := normalized
	if lastSlash != -1 {
		fileName = normalized[lastSlash+1:]
	}
	// 去掉扩展名（如果有）
	dotIndex := strings.Index(fileName, ".")
	if dotIndex != -1 {
		fileName = fileName[:dotIndex]
	}
	fileName = strings.ToUpper(fileName)

	// 保留设备名列表
	reserved := map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true,
		"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
		"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}
	return reserved[fileName]
}
