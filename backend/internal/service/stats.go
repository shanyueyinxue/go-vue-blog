package service

import (
	"blog/internal/dto"
	"blog/internal/model"

	"go.uber.org/zap"
)

// StatsService 仪表盘统计相关业务
type StatsService struct {
	*Service
}

// NewStatsService 创建统计服务
func NewStatsService(base *Service) *StatsService {
	return &StatsService{Service: base}
}

// Summary 仪表盘统计：各类数量与最近 5 篇已发布文章
func (s *StatsService) Summary() *dto.StatsSummary {
	type statsCounts struct {
		PostCount           int64
		PublishedCount      int64
		DraftCount          int64
		CategoryCount       int64
		TagCount            int64
		CommentCount        int64
		PendingCommentCount int64
		TotalViews          int64
	}

	var counts statsCounts
	// 单条 SQL 完成所有统计，减少数据库往返
	if err := s.DB().Raw(`
		SELECT
			(SELECT COUNT(*) FROM posts WHERE posts.deleted_at IS NULL) AS post_count,
			(SELECT COUNT(*) FROM posts WHERE posts.deleted_at IS NULL AND posts.status = ?) AS published_count,
			(SELECT COUNT(*) FROM posts WHERE posts.deleted_at IS NULL AND posts.status = ?) AS draft_count,
			(SELECT COUNT(*) FROM categories WHERE categories.deleted_at IS NULL) AS category_count,
			(SELECT COUNT(*) FROM tags) AS tag_count,
			(SELECT COUNT(*) FROM comments WHERE comments.deleted_at IS NULL) AS comment_count,
			(SELECT COUNT(*) FROM comments WHERE comments.deleted_at IS NULL AND comments.status = ?) AS pending_comment_count,
			(SELECT COALESCE(SUM(posts.view_count), 0) FROM posts WHERE posts.deleted_at IS NULL) AS total_views
	`, PostStatusPublished, PostStatusDraft, model.CommentStatusPending).Scan(&counts).Error; err != nil {
		s.Log().Warn("查询统计信息失败", zap.Error(err))
		return &dto.StatsSummary{RecentPosts: []model.Post{}}
	}

	// 最近 5 篇已发布文章
	var recentPosts []model.Post
	if err := s.DB().Select("id", "title", "slug", "status", "published_at").
		Where("status = ?", PostStatusPublished).
		Order("published_at DESC").
		Limit(5).
		Find(&recentPosts).Error; err != nil {
		s.Log().Warn("查询最近文章失败", zap.Error(err))
		recentPosts = []model.Post{}
	}

	return &dto.StatsSummary{
		PostCount:           counts.PostCount,
		PublishedCount:      counts.PublishedCount,
		DraftCount:          counts.DraftCount,
		CategoryCount:       counts.CategoryCount,
		TagCount:            counts.TagCount,
		CommentCount:        counts.CommentCount,
		PendingCommentCount: counts.PendingCommentCount,
		TotalViews:          counts.TotalViews,
		RecentPosts:         recentPosts,
	}
}
