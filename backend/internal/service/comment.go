package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"blog/internal/dto"
	"blog/internal/middleware"
	"blog/internal/model"
	"blog/pkg/email"
	"blog/pkg/utils"
	"blog/pkg/xss"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// commentLimiter 评论 IP 防刷限流器：同一 IP 每分钟最多 10 次
var commentLimiter = middleware.NewRateLimiter(10, time.Minute)

// CommentService 评论相关业务
type CommentService struct {
	*Service
	Config  *ConfigService
	Captcha *CaptchaService
}

// NewCommentService 创建评论服务
func NewCommentService(base *Service, config *ConfigService, captcha *CaptchaService) *CommentService {
	return &CommentService{Service: base, Config: config, Captcha: captcha}
}

// ListComments 公开获取某文章已通过评论；返回顶层评论（时间升序），replies 为楼中楼子评论。
// 返回公开 DTO（不含 email/phone/ip/user_agent 等隐私字段，见 dto.PublicComment），
// 单次查询取出该文章全部已通过评论，再在内存中组装树，避免按层级反复查询。
func (s *CommentService) ListComments(slug string, req *dto.GetCommentsRequest) ([]*dto.PublicComment, int64, *BizError) {
	// 1. 通过 slug 找到已发布文章
	var post model.Post
	if err := s.DB().Where("slug = ? AND status = ?", slug, PostStatusPublished).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, NewNotFound("文章不存在或未发布")
		}
		return nil, 0, NewServerError("查询文章失败")
	}

	// 2. 统计顶层评论总数（分页依据）
	var total int64
	if err := s.DB().Model(&model.Comment{}).
		Where("post_id = ? AND status = ? AND parent_id = ?", post.ID, model.CommentStatusApproved, 0).
		Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询评论失败")
	}
	start := req.Offset()
	if total == 0 || int64(start) >= total {
		return []*dto.PublicComment{}, total, nil
	}

	// 3. 单次查询取出该文章全部已通过评论（时间升序），内存中组装树
	var all []model.Comment
	if err := s.DB().
		Where("post_id = ? AND status = ?", post.ID, model.CommentStatusApproved).
		Order("created_at ASC").Order("id ASC").
		Find(&all).Error; err != nil {
		return nil, 0, NewServerError("查询评论失败")
	}

	// 4. 当前页顶层评论在 all 中的下标（all 已按时间升序，顶层评论相对顺序一致）
	var rootIndices []int
	for i := range all {
		if all[i].ParentID == 0 {
			rootIndices = append(rootIndices, i)
		}
	}
	end := start + req.PageSize
	if end > len(rootIndices) {
		end = len(rootIndices)
	}
	pageRootIndices := rootIndices[start:end]

	// 5. 按 parent_id 分组，仅组装当前页顶层评论可达的子树（输出为公开 DTO）
	children := make(map[uint][]*model.Comment, len(all))
	for i := range all {
		children[all[i].ParentID] = append(children[all[i].ParentID], &all[i])
	}
	var attach func(parent *model.Comment) *dto.PublicComment
	attach = func(parent *model.Comment) *dto.PublicComment {
		pub := toPublicComment(parent)
		for _, child := range children[parent.ID] {
			pub.Replies = append(pub.Replies, attach(child))
		}
		return pub
	}

	result := make([]*dto.PublicComment, 0, len(pageRootIndices))
	for _, idx := range pageRootIndices {
		result = append(result, attach(&all[idx]))
	}
	return result, total, nil
}

// toPublicComment 将评论模型转换为公开 DTO，剔除 email/phone/ip/user_agent 等隐私字段，
// 防止公开接口经模型 JSON 序列化泄露评论者隐私（docs/api.md 6.1 约定）。
func toPublicComment(c *model.Comment) *dto.PublicComment {
	return &dto.PublicComment{
		ID:         c.ID,
		PostID:     c.PostID,
		ParentID:   c.ParentID,
		Nickname:   c.Nickname,
		Content:    c.Content,
		Status:     c.Status,
		IsAuthor:   c.IsAuthor,
		ReplyCount: c.ReplyCount,
		Likes:      c.Likes,
		CreatedAt:  c.CreatedAt,
	}
}

// CreateComment 提交评论；匿名提交，默认待审核；IP 限制每分钟 10 次防刷；
// 提交前需通过邮箱验证码校验（站点配置 comment_need_captcha 可关闭）
func (s *CommentService) CreateComment(req *dto.CreateCommentRequest, ip, userAgent string) (*model.Comment, string, *BizError) {
	// 昵称：trim 后校验，避免超长昵称先触发数据库错误再返回 500
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		return nil, "", NewParamError("昵称不能为空")
	}
	if utf8.RuneCountInString(nickname) > 50 {
		return nil, "", NewParamError("昵称长度不能超过 50 个字符")
	}
	// 邮箱必填（用于接收验证码与评论通知）
	if !utils.IsEmail(req.Email) {
		return nil, "", NewParamError("请填写有效邮箱")
	}
	// 内容长度限制（按字符数）
	content := strings.TrimSpace(req.Content)
	if utf8.RuneCountInString(content) < 1 || utf8.RuneCountInString(content) > 2000 {
		return nil, "", NewParamError("评论内容长度需在 1-2000 字符之间")
	}

	// IP 防刷：同一 IP 每分钟最多 10 条评论
	if !commentLimiter.Allow(ip) {
		return nil, "", NewTooManyRequests("评论过于频繁，请稍后再试")
	}

	// 邮箱验证码校验（读取当前生效配置，可关闭）
	needCaptcha := 1
	if cfg := s.Config.ActiveConfig(); cfg != nil {
		needCaptcha = cfg.CommentNeedCaptcha
	}
	if needCaptcha == 1 {
		if err := s.Captcha.VerifyForEmail(req.CaptchaToken, req.CaptchaCode, req.Email); err != nil {
			return nil, "", err
		}
	}

	// 1. 定位已发布文章（post_id 或 slug 二选一）
	var post model.Post
	if req.PostID > 0 {
		err := s.DB().Where("id = ? AND status = ?", req.PostID, PostStatusPublished).First(&post).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, "", NewNotFound("文章不存在或未发布")
			}
			return nil, "", NewServerError("查询文章失败")
		}
	} else if req.Slug != "" {
		err := s.DB().Where("slug = ? AND status = ?", req.Slug, PostStatusPublished).First(&post).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, "", NewNotFound("文章不存在或未发布")
			}
			return nil, "", NewServerError("查询文章失败")
		}
	} else {
		return nil, "", NewNotFound("文章不存在")
	}

	// 2. 校验父评论（若为回复）
	if req.ParentID > 0 {
		var parent model.Comment
		if err := s.DB().First(&parent, req.ParentID).Error; err != nil || parent.PostID != post.ID {
			return nil, "", NewParamError("父评论不存在或不属于该文章")
		}
	}

	// 3. 读取当前生效配置，判断是否需要审核
	needReview := 1
	if cfg := s.Config.ActiveConfig(); cfg != nil {
		needReview = cfg.CommentNeedReview
	}
	status := model.CommentStatusApproved
	if needReview == 1 {
		status = model.CommentStatusPending
	}

	// 4. 构造评论（昵称与内容均做 XSS 过滤，使用 trim 后的值）
	comment := model.Comment{
		PostID:    post.ID,
		ParentID:  req.ParentID,
		Nickname:  xss.Sanitize(nickname),
		Email:     req.Email,
		Phone:     req.Phone,
		Content:   xss.Sanitize(content),
		Status:    status,
		IsAuthor:  0,
		IP:        ip,
		UserAgent: userAgent,
	}
	if status == model.CommentStatusApproved {
		now := time.Now()
		comment.ReviewedAt = &now
	}

	// 5. 事务写入评论并同步计数
	err := s.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		// 审核通过时更新文章评论数
		if status == model.CommentStatusApproved {
			if err := tx.Model(&post).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
				return err
			}
			// 更新父评论的回复数
			if req.ParentID > 0 {
				if err := tx.Model(&model.Comment{}).Where("id = ?", req.ParentID).
					UpdateColumn("reply_count", gorm.Expr("reply_count + 1")).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, "", NewServerError("提交评论失败")
	}

	// 6. 异步发送邮件通知（失败不影响主流程）
	//   - 直接通过时通知评论者
	if status == model.CommentStatusApproved && utils.IsEmail(req.Email) {
		s.notifyApproved(req.Email, req.Nickname)
	}
	//   - 有回复时通知父评论者
	if req.ParentID > 0 {
		var parent model.Comment
		if err := s.DB().First(&parent, req.ParentID).Error; err == nil && utils.IsEmail(parent.Email) {
			s.notifyReplied(parent.Email, parent.Nickname)
		}
	}

	s.Log().Info("收到评论",
		zap.Uint("id", comment.ID),
		zap.Uint("post_id", post.ID),
		zap.String("status", status),
		zap.String("ip", ip),
	)
	return &comment, status, nil
}

// ListAdminComments 管理端评论列表（支持状态、文章、关键字筛选）
func (s *CommentService) ListAdminComments(req *dto.AdminCommentQuery) ([]model.Comment, int64, *BizError) {
	query := s.DB().Model(&model.Comment{})
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.PostID > 0 {
		query = query.Where("post_id = ?", req.PostID)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("(nickname LIKE ? OR content LIKE ?)", kw, kw)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询评论失败")
	}

	var comments []model.Comment
	if err := query.Order("id DESC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&comments).Error; err != nil {
		return nil, 0, NewServerError("查询评论失败")
	}
	s.populateAdminCommentMeta(comments)
	return comments, total, nil
}

// populateAdminCommentMeta 填充评论列表的关联文章标题和父评论昵称。
func (s *CommentService) populateAdminCommentMeta(comments []model.Comment) {
	if len(comments) == 0 {
		return
	}

	postIDs := make([]uint, 0, len(comments))
	parentIDs := make([]uint, 0, len(comments))
	for _, comment := range comments {
		if comment.PostID > 0 {
			postIDs = append(postIDs, comment.PostID)
		}
		if comment.ParentID > 0 {
			parentIDs = append(parentIDs, comment.ParentID)
		}
	}

	if len(postIDs) > 0 {
		var posts []model.Post
		if err := s.DB().Select("id", "title").Where("id IN ?", postIDs).Find(&posts).Error; err == nil {
			titleMap := make(map[uint]string, len(posts))
			for _, post := range posts {
				titleMap[post.ID] = post.Title
			}
			for i := range comments {
				comments[i].PostTitle = titleMap[comments[i].PostID]
			}
		}
	}

	if len(parentIDs) > 0 {
		var parents []model.Comment
		if err := s.DB().Select("id", "nickname").Where("id IN ?", parentIDs).Find(&parents).Error; err == nil {
			nicknameMap := make(map[uint]string, len(parents))
			for _, parent := range parents {
				nicknameMap[parent.ID] = parent.Nickname
			}
			for i := range comments {
				comments[i].ParentNickname = nicknameMap[comments[i].ParentID]
			}
		}
	}
}

// UpdateComment 管理端修改评论：**仅允许修改 status 与 is_author 两个字段**，其余字段一律忽略。
// - status 变更时同步维护 posts.comment_count 与父评论 reply_count（approved ↔ 非 approved），
//   审核通过时写入 reviewed_at 并通知评论者；
// - is_author 变更直接更新；
// - 两个字段都未变化时直接返回，不产生更新。
func (s *CommentService) UpdateComment(id uint, req *dto.CommentAdminUpdateRequest) (*model.Comment, *BizError) {
	var comment model.Comment
	if err := s.DB().First(&comment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("评论不存在")
		}
		return nil, NewServerError("查询评论失败")
	}

	statusChanged := req.Status != nil && *req.Status != comment.Status
	authorChanged := req.IsAuthor != nil && *req.IsAuthor != comment.IsAuthor
	if !statusChanged && !authorChanged {
		return &comment, nil
	}

	// 计数同步（仅状态变化时；delta 需在 Updates 之前计算，
	// GORM 的 Updates(map) 会就地改写 comment.Status 导致后续判断失效）：
	//   从非 approved 变为 approved → 文章评论数 +1（若为回复，父评论回复数 +1）
	//   从 approved 变为非 approved → 文章评论数 -1（若为回复，父评论回复数 -1）
	delta := 0
	if statusChanged {
		if *req.Status == model.CommentStatusApproved && comment.Status != model.CommentStatusApproved {
			delta = 1
		} else if comment.Status == model.CommentStatusApproved && *req.Status != model.CommentStatusApproved {
			delta = -1
		}
	}

	err := s.DB().Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{}
		if statusChanged {
			updates["status"] = *req.Status
			if *req.Status == model.CommentStatusApproved {
				now := time.Now()
				updates["reviewed_at"] = now
			}
		}
		if authorChanged {
			updates["is_author"] = *req.IsAuthor
		}
		if err := tx.Model(&comment).Updates(updates).Error; err != nil {
			return err
		}

		if delta != 0 {
			if err := tx.Model(&model.Post{}).Where("id = ?", comment.PostID).
				UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error; err != nil {
				return err
			}
			if comment.ParentID > 0 {
				if err := tx.Model(&model.Comment{}).Where("id = ?", comment.ParentID).
					UpdateColumn("reply_count", gorm.Expr("reply_count + ?", delta)).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, NewServerError("更新评论失败")
	}

	// 审核通过时通知评论者
	if statusChanged && *req.Status == model.CommentStatusApproved && utils.IsEmail(comment.Email) {
		s.notifyApproved(comment.Email, comment.Nickname)
	}

	s.Log().Info("更新评论",
		zap.Uint("id", comment.ID),
		zap.String("old_status", comment.Status),
		zap.Int("old_is_author", comment.IsAuthor),
	)
	return &comment, nil
}

// DeleteComment 删除评论（软删除）；存在子评论时级联逻辑删除；同步计数
func (s *CommentService) DeleteComment(id uint) *BizError {
	var comment model.Comment
	if err := s.DB().First(&comment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("评论不存在")
		}
		return NewServerError("查询评论失败")
	}

	// 收集将被删除（含级联子评论）的已通过评论数量
	deletedApproved := 0
	if comment.Status == model.CommentStatusApproved {
		deletedApproved = 1
	}
	err := s.DB().Transaction(func(tx *gorm.DB) error {
		// 递归收集并软删除所有子评论
		var idsToDelete []uint
		var collect func(parentID uint)
		collect = func(parentID uint) {
			var children []model.Comment
			tx.Where("parent_id = ?", parentID).Find(&children)
			for _, child := range children {
				collect(child.ID)
				idsToDelete = append(idsToDelete, child.ID)
			}
		}
		collect(comment.ID)

		// 统计子评论中已通过的数量（用于回退计数）
		if len(idsToDelete) > 0 {
			var approvedCount int64
			tx.Model(&model.Comment{}).Where("id IN ? AND status = ?", idsToDelete, model.CommentStatusApproved).
				Count(&approvedCount)
			deletedApproved += int(approvedCount)
			if err := tx.Where("id IN ?", idsToDelete).Delete(&model.Comment{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Delete(&comment).Error; err != nil {
			return err
		}

		// 回退文章评论数与父评论回复数
		if deletedApproved > 0 {
			if err := tx.Model(&model.Post{}).Where("id = ?", comment.PostID).
				UpdateColumn("comment_count", gorm.Expr("comment_count - ?", deletedApproved)).Error; err != nil {
				return err
			}
			if comment.ParentID > 0 {
				if err := tx.Model(&model.Comment{}).Where("id = ?", comment.ParentID).
					UpdateColumn("reply_count", gorm.Expr("reply_count - ?", deletedApproved)).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return NewServerError("删除评论失败")
	}
	s.Log().Info("删除评论", zap.Uint("id", comment.ID), zap.String("nickname", comment.Nickname))
	return nil
}

// notifyApproved 评论审核通过时通知评论者（异步，失败不影响主流程）
func (s *CommentService) notifyApproved(to, nickname string) {
	cfg := s.Config.ActiveConfig()
	siteName := "我的博客"
	if cfg != nil && cfg.Title != "" {
		siteName = cfg.Title
	}
	subject := fmt.Sprintf("您在 %s 的评论已通过审核", siteName)
	body := fmt.Sprintf("<p>您好 <strong>%s</strong>：</p><p>您在「%s」发表的评论已通过审核并成功发布。</p>", nickname, siteName)
	s.sendEmailAsync(to, subject, body)
}

// notifyReplied 评论收到回复时通知父评论者（异步）
func (s *CommentService) notifyReplied(to, parentNickname string) {
	subject := "您的评论收到了新回复"
	body := fmt.Sprintf("<p>您好 <strong>%s</strong>：</p><p>有人回复了您的评论，欢迎回来查看。</p>", parentNickname)
	s.sendEmailAsync(to, subject, body)
}

// sendEmailAsync 异步发送邮件，使用 defer/recover 保证不 panic
func (s *CommentService) sendEmailAsync(to, subject, body string) {
	if to == "" {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.Log().Warn("异步发送邮件发生异常", zap.Any("recover", r))
			}
		}()
		msg := &email.Email{
			To:      []string{to},
			Subject: subject,
			Body:    body,
		}
		if err := s.App.Email.DialAndSend(msg); err != nil {
			s.Log().Warn("发送邮件失败", zap.Error(err), zap.String("to", to))
		}
	}()
}
