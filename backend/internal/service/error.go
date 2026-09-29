package service

import (
	"errors"

	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// BizError 业务错误，携带 HTTP 状态码与业务错误码，便于 handler 统一响应
type BizError struct {
	HTTPStatus int    // HTTP 状态码
	Code       int    // 业务错误码
	Message    string // 错误提示信息
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	return e.Message
}

// NewParamError 参数错误（HTTP 400）
func NewParamError(msg string) *BizError {
	return &BizError{HTTPStatus: 400, Code: response.CodeParamError, Message: msg}
}

// NewUnauthorized 未认证（HTTP 401）
func NewUnauthorized(msg string) *BizError {
	return &BizError{HTTPStatus: 401, Code: response.CodeUnauthorized, Message: msg}
}

// NewForbidden 无权限（HTTP 403）
func NewForbidden(msg string) *BizError {
	return &BizError{HTTPStatus: 403, Code: response.CodeForbidden, Message: msg}
}

// NewNotFound 资源不存在（HTTP 404）
func NewNotFound(msg string) *BizError {
	return &BizError{HTTPStatus: 404, Code: response.CodeNotFound, Message: msg}
}

// NewConflict 资源冲突（HTTP 409）
func NewConflict(msg string) *BizError {
	return &BizError{HTTPStatus: 409, Code: response.CodeConflict, Message: msg}
}

// NewTooManyRequests 请求过于频繁（HTTP 429）
func NewTooManyRequests(msg string) *BizError {
	return &BizError{HTTPStatus: 429, Code: response.CodeTooManyRequest, Message: msg}
}

// NewServerError 服务器内部错误（HTTP 500）
func NewServerError(msg string) *BizError {
	return &BizError{HTTPStatus: 500, Code: response.CodeServerError, Message: msg}
}

// HandleError 将 error 转换为统一 HTTP 响应；未知错误按服务器内部错误处理
func HandleError(c *gin.Context, err error) {
	var be *BizError
	if errors.As(err, &be) {
		response.Error(c, be.HTTPStatus, be.Code, be.Message)
		return
	}
	response.ServerError(c, "服务器内部错误")
}
