package validate

import (
	"blog/pkg/utils"

	validator "github.com/go-playground/validator/v10"
)

// ValidatePhone 验证手机号码是否合法
func validatePhone(fl validator.FieldLevel) bool {
	phoneStr := fl.Field().String()

	if phoneStr == "" {
		return true
	}

	return utils.IsPhoneInChina(phoneStr)
}
