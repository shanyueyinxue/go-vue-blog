package handler

import (
	"blog/internal/app"
	"blog/internal/service"
)

// Handler 所有 HTTP handler 的基类，持有应用依赖与业务服务
type Handler struct {
	App             *app.App
	AuthService     *service.AuthService
	CaptchaService  *service.CaptchaService
	UserService     *service.UserService
	CategoryService *service.CategoryService
	TagService      *service.TagService
	PostService     *service.PostService
	CommentService  *service.CommentService
	ConfigService   *service.ConfigService
	FriendService   *service.FriendService
	WorkService     *service.WorkService
	FileService     *service.FileService
	MusicService    *service.MusicService
	StatsService    *service.StatsService
}

// New 创建 Handler 并组装全部业务服务
func New(a *app.App) *Handler {
	base := service.New(a)
	configSvc := service.NewConfigService(base)
	captchaSvc := service.NewCaptchaService(base)
	return &Handler{
		App:             a,
		AuthService:     service.NewAuthService(base),
		CaptchaService:  captchaSvc,
		UserService:     service.NewUserService(base, captchaSvc),
		CategoryService: service.NewCategoryService(base),
		TagService:      service.NewTagService(base),
		PostService:     service.NewPostService(base),
		CommentService:  service.NewCommentService(base, configSvc, captchaSvc),
		ConfigService:   configSvc,
		FriendService:   service.NewFriendService(base),
		WorkService:     service.NewWorkService(base),
		FileService:     service.NewFileService(base),
		MusicService:    service.NewMusicService(base),
		StatsService:    service.NewStatsService(base),
	}
}
