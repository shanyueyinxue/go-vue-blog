package router

import (
	"strings"
	"time"

	"blog/internal/app"
	"blog/internal/handler"
	"blog/internal/middleware"
	"blog/pkg/response"
	"blog/pkg/validate"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Setup 注册全部路由并返回 gin.Engine
func Setup(a *app.App) *gin.Engine {
	// 设置 Gin 运行模式
	gin.SetMode(ginMode(a.Config.App.Debug))
	// 注册自定义校验器（date / phone 等）
	validate.RegisterValidation()

	r := gin.New()
	// 配置可信代理：只有来自这些代理的请求才信任 X-Forwarded-For 等头
	if err := r.SetTrustedProxies(a.Config.Server.TrustedProxies); err != nil {
		a.Logger.Fatal("配置 trustedProxies 失败", zap.Error(err))
	}
	// 全局中间件：日志、5xx 错误详情、Recovery、CORS、安全响应头。
	// ErrorLog 必须注册在 Recovery 之前（更外层）：panic 时 Recovery 的 defer 先写 500，
	// ErrorLog 的 c.Next() 之后代码上浮时才能读到 status=500 并记录错误详情。
	// 5xx 错误通知器：生产环境异步邮件通知站长（带 5 分钟冷却防风暴），开发环境仅记日志。
	// SecurityHeaders 输出 X-Content-Type-Options: nosniff 等安全响应头（对所有响应含静态文件生效）。
	errorNotifier := middleware.NewErrorNotifier(a.Logger, a.Email, a.Config.User.MasterEmail, a.Config.App.Env)
	r.Use(gin.Logger(), middleware.ErrorLog(a.Logger, errorNotifier), gin.Recovery(), middleware.CORS(a.Config.CORS.Origins), middleware.SecurityHeaders())

	h := handler.New(a)

	// 仅 local 存储开启静态文件访问；OSS 文件通过对象存储地址访问
	if a.Config.Storage.Type == "local" && a.Config.Storage.Local.BasePath != "" {
		// 路由前缀必须绝对路径（如 /uploads），相对值会导致 gin 路由异常
		baseURL := strings.Trim(a.Config.Storage.Local.BaseURL, "/")
		if baseURL == "" {
			baseURL = "uploads"
		}
		r.Static("/"+baseURL, a.Config.Storage.Local.BasePath)
	}

	// ---------- 公开接口（无需认证） ----------
	public := r.Group("/api")
	{
		if a.Config.App.Debug && a.Config.App.Env == "development" {
			public.GET("/ping", func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{
					"msg": "pong",
				})
			})
			public.GET("/test-panic", func(ctx *gin.Context) {
				panic("test-panic")
			})
		}

		// 站点配置
		public.GET("/configs/public", h.PublicConfig)

		// 文章
		public.GET("/posts", h.ListPosts)
		public.GET("/posts/archive", h.Archive)
		public.GET("/posts/:slug", h.GetPost)
		public.POST("/posts/:slug/view", h.ViewPost)
		public.GET("/posts/:slug/comments", h.ListComments)

		// 评论提交（匿名）
		public.POST("/comments", h.CreateComment)

		// 分类 / 标签
		public.GET("/categories", h.ListCategories)
		public.GET("/tags", h.ListTags)

		// 友链
		public.GET("/friends", h.ListFriends)

		// 作品集
		public.GET("/works", h.ListWorks)

		// 音乐（APlayer 播放列表）
		public.GET("/musics", h.ListMusics)

		// 验证码
		// 发送邮箱验证码：按 IP 限流（每分钟 5 次），防止恶意遍历邮箱地址造成邮件轰炸 /
		// SMTP 配额耗尽；验证码本身还有"同一邮箱 60 秒冷却"（service/captcha.go）
		public.POST("/captcha/email", middleware.RateLimit(5, time.Minute), h.SendEmailCaptcha)
		public.POST("/captcha/verify", h.VerifyCaptcha)

		// 忘记密码（公开，无需 JWT）
		// 发送验证码按 IP 限流（每分钟 5 次）防邮件轰炸
		public.POST("/forgot-password/send-code", middleware.RateLimit(5, time.Minute), h.SendForgotPasswordCode)
		public.POST("/forgot-password/reset", h.ResetForgotPassword)
	}

	// ---------- 管理接口（需 JWT） ----------
	admin := r.Group("/api/admin")
	{
		// 登录与刷新 Token（无需 JWT；登录加防爆破限流）
		admin.POST("/login", middleware.LoginRateLimit(), h.Login)
		admin.POST("/refresh", h.RefreshToken)

		// 以下接口均需 JWT 认证
		authed := admin.Group("", middleware.JWTAuth(a.JWT, a.DB))
		{
			// 管理员信息
			authed.GET("/profile", h.Profile)
			authed.PUT("/profile", h.UpdateProfile)
			authed.PUT("/profile/password", h.UpdatePassword)

			// 用户管理（仅超级管理员 is_master=1）
			users := authed.Group("/users", middleware.MasterOnly())
			{
				users.GET("", h.ListUsers)
				users.POST("", h.CreateUser)
				users.GET("/:id", h.GetUser)
				users.PUT("/:id", h.UpdateUser)
				users.PUT("/:id/password", h.ResetUserPassword)
				users.DELETE("/:id", h.DeleteUser)
			}

			// 文章管理
			authed.GET("/posts", h.ListAdminPosts)
			authed.POST("/posts", h.CreatePost)
			authed.GET("/posts/:id", h.GetAdminPost)
			authed.PUT("/posts/:id", h.UpdatePost)
			authed.DELETE("/posts/:id", h.DeletePost)

			// 分类管理
			authed.GET("/categories", h.ListAdminCategories)
			authed.POST("/categories", h.CreateCategory)
			authed.PUT("/categories/:id", h.UpdateCategory)
			authed.DELETE("/categories/:id", h.DeleteCategory)

			// 标签管理
			authed.GET("/tags", h.ListAdminTags)
			authed.POST("/tags", h.CreateTag)
			authed.PUT("/tags/:id", h.UpdateTag)
			authed.DELETE("/tags/:id", h.DeleteTag)

			// 评论管理
			authed.GET("/comments", h.ListAdminComments)
			authed.PUT("/comments/:id", h.UpdateComment)
			authed.DELETE("/comments/:id", h.DeleteComment)

			// 站点配置管理
			authed.GET("/configs", h.ListAdminConfigs)
			authed.GET("/configs/:id", h.GetAdminConfig)
			authed.POST("/configs", h.CreateConfig)
			authed.PUT("/configs/:id", h.UpdateConfig)
			authed.PUT("/configs/:id/activate", h.ActivateConfig)
			authed.DELETE("/configs/:id", h.DeleteConfig)

			// 友链管理
			authed.GET("/friends", h.ListAdminFriends)
			authed.POST("/friends", h.CreateFriend)
			authed.PUT("/friends/:id", h.UpdateFriend)
			authed.DELETE("/friends/:id", h.DeleteFriend)

			// 作品集管理
			authed.GET("/works", h.ListAdminWorks)
			authed.POST("/works", h.CreateWork)
			authed.GET("/works/:id", h.GetAdminWork)
			authed.PUT("/works/:id", h.UpdateWork)
			authed.DELETE("/works/:id", h.DeleteWork)

			// 文件上传
			authed.POST("/files", h.UploadFile)
			authed.GET("/files", h.ListFiles)
			authed.DELETE("/files/:id", h.DeleteFile)

			// 音乐管理（音频文件经 /files 上传，删除音乐同步清理关联文件）
			authed.GET("/musics", h.ListAdminMusics)
			authed.POST("/musics", h.CreateMusic)
			authed.GET("/musics/:id", h.GetAdminMusic)
			authed.PUT("/musics/:id", h.UpdateMusic)
			authed.DELETE("/musics/:id", h.DeleteMusic)

			// 仪表盘统计
			authed.GET("/stats/summary", h.StatsSummary)
		}
	}

	// 未匹配路由统一返回 40400
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "接口不存在")
	})

	return r
}

// ginMode 根据配置返回 Gin 运行模式
func ginMode(debug bool) string {
	if debug {
		return gin.DebugMode
	}
	return gin.ReleaseMode
}
