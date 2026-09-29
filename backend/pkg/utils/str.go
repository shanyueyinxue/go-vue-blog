package utils

import (
	"fmt"
	"math"
	"math/rand"
)

func RandomStringMixed(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return RandomStringWithCharset(n, letters)
}

func RandomNumber(n int) string {
	const numbers = "0123456789"
	return RandomStringWithCharset(n, numbers)
}

func RandomLetter(n int) string {
	const numbers = "abcdefghijklmnopqrstuvwxyz"
	return RandomStringWithCharset(n, numbers)
}

func RandomStringWithCharset(n int, charset string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// SliceToString 切片转换为字符串； sep 为分隔符
func SliceToString(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}

	if len(slice) == 1 {
		return slice[0]
	}
	result := ""
	for i, s := range slice {
		result += s
		if i < len(slice)-1 {
			result += sep
		}
	}
	return result
}

// SecondsToMinutes 秒数转换为分钟数; 向上取整
func SecondsToMinutes(seconds int) string {
	minutes := float64(seconds) / 60.0
	roundedMinutes := math.Ceil(minutes)
	return fmt.Sprintf("%d", int(roundedMinutes))
}
