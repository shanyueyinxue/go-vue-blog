package service

import (
	"testing"
	"time"

	"blog/internal/dto"
	"blog/internal/model"

	"github.com/stretchr/testify/require"
)

// TestCommentAdminUpdate 管理端修改评论：仅允许 status / is_author，
// status 变更时同步维护 posts.comment_count 计数
func TestCommentAdminUpdate(t *testing.T) {
	a := newTestApp(t)
	base := New(a)
	configSvc := NewConfigService(base)
	captchaSvc := NewCaptchaService(base)
	commentSvc := NewCommentService(base, configSvc, captchaSvc)

	// 配置：免验证码、直接通过（便于验证计数）
	_, err := configSvc.CreateConfig(&dto.ConfigUpsertRequest{
		Version:            "1.0.0",
		CommentNeedReview:  intPtr(0),
		CommentNeedCaptcha: intPtr(0),
	})
	require.Nil(t, err)

	now := time.Now()
	post := model.Post{Title: "评论修改测试", Slug: "comment-update-post", Status: PostStatusPublished, PublishedAt: &now}
	require.NoError(t, a.DB.Create(&post).Error)

	// 创建评论（直接通过 → approved，文章计数 +1）
	comment, status, bizErr := commentSvc.CreateComment(&dto.CreateCommentRequest{
		Slug: post.Slug, Nickname: "访客", Email: "visitor@example.com", Content: "内容",
	}, "10.0.0.20", "ua")
	require.Nil(t, bizErr)
	require.Equal(t, model.CommentStatusApproved, status)

	var postRow model.Post
	require.NoError(t, a.DB.First(&postRow, post.ID).Error)
	require.Equal(t, 1, postRow.CommentCount)

	// 1. 修改 is_author（不涉及计数）
	updated, bizErr := commentSvc.UpdateComment(comment.ID, &dto.CommentAdminUpdateRequest{IsAuthor: intPtr(1)})
	require.Nil(t, bizErr)
	require.Equal(t, 1, updated.IsAuthor)
	require.Equal(t, model.CommentStatusApproved, updated.Status)

	// 2. approved → rejected：计数回退为 0
	_, bizErr = commentSvc.UpdateComment(comment.ID, &dto.CommentAdminUpdateRequest{Status: strPtr("rejected")})
	require.Nil(t, bizErr)
	require.NoError(t, a.DB.First(&postRow, post.ID).Error)
	require.Zero(t, postRow.CommentCount)

	// 3. rejected → approved：计数恢复 1，且写入 reviewed_at
	updated, bizErr = commentSvc.UpdateComment(comment.ID, &dto.CommentAdminUpdateRequest{Status: strPtr("approved")})
	require.Nil(t, bizErr)
	require.Equal(t, model.CommentStatusApproved, updated.Status)
	require.NotNil(t, updated.ReviewedAt)
	require.NoError(t, a.DB.First(&postRow, post.ID).Error)
	require.Equal(t, 1, postRow.CommentCount)

	// 4. approved → pending：计数回退为 0
	_, bizErr = commentSvc.UpdateComment(comment.ID, &dto.CommentAdminUpdateRequest{Status: strPtr("pending")})
	require.Nil(t, bizErr)
	require.NoError(t, a.DB.First(&postRow, post.ID).Error)
	require.Zero(t, postRow.CommentCount)

	// 5. 状态相同：直接返回，计数不再变化
	_, bizErr = commentSvc.UpdateComment(comment.ID, &dto.CommentAdminUpdateRequest{Status: strPtr("pending")})
	require.Nil(t, bizErr)
	require.NoError(t, a.DB.First(&postRow, post.ID).Error)
	require.Zero(t, postRow.CommentCount)

	// 6. is_author 相同：直接返回，无副作用
	_, bizErr = commentSvc.UpdateComment(comment.ID, &dto.CommentAdminUpdateRequest{IsAuthor: intPtr(1)})
	require.Nil(t, bizErr)

	// 7. 评论不存在 → NotFound
	_, bizErr = commentSvc.UpdateComment(99999, &dto.CommentAdminUpdateRequest{Status: strPtr("approved")})
	require.NotNil(t, bizErr)
}
