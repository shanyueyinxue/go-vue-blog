package validate

import (
	validator "github.com/go-playground/validator/v10"
	"time"
)

// 自定义日期验证函数
func validateDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()

	if dateStr == "" {
		return true
	}

	// 尝试解析日期，支持多种常见格式
	layouts := []string{
		"2006-01-02",
		"2006/01/02",
		"01-02-2006",
		"01/02/2006",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if _, err := time.Parse(layout, dateStr); err == nil {
			return true
		}
	}

	return false
}
