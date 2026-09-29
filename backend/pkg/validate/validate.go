package validate

import (
	"github.com/gin-gonic/gin/binding"
	validator "github.com/go-playground/validator/v10"
)

func RegisterValidation() {
	// 获取Gin使用的验证器实例
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证器
		v.RegisterValidation("date", validateDate)
		v.RegisterValidation("phone", validatePhone)
	}
}
