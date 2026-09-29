package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 业务错误码（与 docs/api.md 1.4 保持一致）
const (
	CodeSuccess        = 0     // 成功
	CodeParamError     = 40000 // 请求参数错误
	CodeUnauthorized   = 40100 // 未认证或 Token 已过期/无效
	CodeForbidden      = 40300 // 无权限（非管理员）
	CodeNotFound       = 40400 // 资源不存在
	CodeConflict       = 40900 // 资源冲突（如 slug 已存在）
	CodeTooManyRequest = 42900 // 请求过于频繁（限流）
	CodeServerError    = 50000 // 服务器内部错误
)

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code"`    // 错误码，0 表示成功
	Message string `json:"message"` // 提示信息
	Data    any    `json:"data"`    // 业务数据，可为 null
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: "ok", Data: data})
}

// Error 通用错误响应
func Error(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message, Data: nil})
}

// ParamError 参数错误（HTTP 400）
func ParamError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, CodeParamError, message)
}

// Unauthorized 未认证（HTTP 401）
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden 无权限（HTTP 403）
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, CodeForbidden, message)
}

// NotFound 资源不存在（HTTP 404）
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message)
}

// Conflict 资源冲突（HTTP 409）
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, CodeConflict, message)
}

// TooManyRequests 请求过于频繁（HTTP 429）
func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, CodeTooManyRequest, message)
}

// ServerError 服务器内部错误（HTTP 500）
func ServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, CodeServerError, message)
}
