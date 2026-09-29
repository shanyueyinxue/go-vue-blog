package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"blog/internal/dto"
	"blog/internal/model"
	"blog/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 文章状态常量
const (
	PostStatusDraft     = "draft"     // 草稿
	PostStatusPublished = "published" // 发布
)

// PostService 文章相关业务
type PostService struct {
	*Service
}

// NewPostService 创建文章服务
func NewPostService(base *Service) *PostService {
	return &PostService{Service: base}
}

// ListPosts 公开文章列表（仅已发布，按置顶与排序规则分页）
func (s *PostService) ListPosts(req *dto.PostListRequest) ([]model.Post, int64, *BizError) {
	query := s.DB().Model(&model.Post{}).
		Where("status = ?", PostStatusPublished)

	// 条件筛选
	if req.CategoryID > 0 {
		query = query.Where("category_id = ?", req.CategoryID)
	}
	if req.TagID > 0 {
		// 按标签筛选：关联 post_tags 表
		query = query.Where("id IN (SELECT post_id FROM post_tags WHERE tag_id = ?)", req.TagID)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("(title LIKE ? OR excerpt LIKE ?)", kw, kw)
	}

	// 总数统计
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询文章失败")
	}

	// 排序：置顶优先 + 指定排序方式
	topFirst := strings.ToLower(req.TopFirst) != "false" // 默认置顶优先
	switch req.Sort {
	case "views":
		if topFirst {
			query = query.Order("is_top DESC").Order("view_count DESC")
		} else {
			query = query.Order("view_count DESC")
		}
	case "oldest":
		if topFirst {
			query = query.Order("is_top DESC").Order("published_at ASC")
		} else {
			query = query.Order("published_at ASC")
		}
	default: // latest（默认）
		if topFirst {
			query = query.Order("is_top DESC").Order("published_at DESC")
		} else {
			query = query.Order("published_at DESC")
		}
	}

	// 列表不含正文，仅含摘要；预加载分类与标签
	var posts []model.Post
	err := query.
		Omit("content").
		Preload("Category").
		Preload("Tags").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&posts).Error
	if err != nil {
		return nil, 0, NewServerError("查询文章失败")
	}
	return posts, total, nil
}

// Archive 文章归档：按年月分组，年月降序输出。
// 月度文章数由 SQL 聚合（GROUP BY），再一次性取出文章列表挂载到对应月份，
// 避免把全部文章读进内存后再手动统计。
func (s *PostService) Archive() ([]dto.ArchiveItem, *BizError) {
	// 1. SQL 按月统计（按数据库方言选择日期格式化函数）
	ymSQL := "strftime('%Y-%m', published_at) AS ym, COUNT(*) AS cnt"
	switch s.DB().Dialector.Name() {
	case "mysql":
		ymSQL = "DATE_FORMAT(published_at, '%Y-%m') AS ym, COUNT(*) AS cnt"
	case "postgres":
		ymSQL = "TO_CHAR(published_at, 'YYYY-MM') AS ym, COUNT(*) AS cnt"
	}
	type monthStat struct {
		YM  string
		Cnt int64
	}
	var stats []monthStat
	if err := s.DB().Model(&model.Post{}).
		Select(ymSQL).
		Where("status = ? AND published_at IS NOT NULL", PostStatusPublished).
		Group("ym").
		Order("ym DESC").
		Scan(&stats).Error; err != nil {
		return nil, NewServerError("查询归档失败")
	}
	if len(stats) == 0 {
		return []dto.ArchiveItem{}, nil
	}

	// 2. 一次取出全部已发布文章（接口需返回每月的文章列表）
	var posts []model.Post
	if err := s.DB().Model(&model.Post{}).
		Select("id, title, slug, published_at").
		Where("status = ? AND published_at IS NOT NULL", PostStatusPublished).
		Order("published_at ASC").
		Find(&posts).Error; err != nil {
		return nil, NewServerError("查询归档失败")
	}

	// 3. 按月份挂载文章
	byMonth := make(map[string][]dto.ArchivePost, len(posts))
	for _, p := range posts {
		ym := p.PublishedAt.In(time.Local).Format("2006-01")
		byMonth[ym] = append(byMonth[ym], dto.ArchivePost{
			ID:          p.ID,
			Title:       p.Title,
			Slug:        p.Slug,
			PublishedAt: p.PublishedAt,
		})
	}

	// 4. 组装响应（SQL 已按 ym 降序）
	items := make([]dto.ArchiveItem, 0, len(stats))
	for _, st := range stats {
		year, month := parseYM(st.YM)
		items = append(items, dto.ArchiveItem{
			Year:  year,
			Month: month,
			Count: st.Cnt,
			Posts: byMonth[st.YM],
		})
	}
	return items, nil
}

// parseYM 解析 "YYYY-MM" 为 (年, 月)
func parseYM(ym string) (int, int) {
	parts := strings.Split(ym, "-")
	year, _ := strconv.Atoi(parts[0])
	month := 0
	if len(parts) > 1 {
		month, _ = strconv.Atoi(parts[1])
	}
	return year, month
}

// GetPost 公开文章详情（仅已发布）
func (s *PostService) GetPost(slug string) (*model.Post, *BizError) {
	var post model.Post
	err := s.DB().Preload("Category").Preload("Tags").
		Where("slug = ? AND status = ?", slug, PostStatusPublished).
		First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("文章不存在或未发布")
		}
		return nil, NewServerError("查询文章失败")
	}
	return &post, nil
}

// ViewPost 文章浏览量 +1；按 IP 每小时去重，返回最新浏览量
func (s *PostService) ViewPost(slug, ip string) (int, *BizError) {
	var post model.Post
	if err := s.DB().Where("slug = ? AND status = ?", slug, PostStatusPublished).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, NewNotFound("文章不存在或未发布")
		}
		return 0, NewServerError("查询文章失败")
	}

	// IP 去重：命中缓存说明 1 小时内已计过
	viewKey := fmt.Sprintf("post_view:%s:%s", ip, slug)
	if !s.Cache().Exists(viewKey) {
		// 原子更新浏览量
		if err := s.DB().Model(&post).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err == nil {
			post.ViewCount++
		}
		// 记录去重标记，1 小时后过期
		_ = s.Cache().Set(viewKey, "1", time.Hour)
	}
	return post.ViewCount, nil
}

// ListAdminPosts 管理端文章列表（支持状态、关键字、分类、标签、置顶筛选，含草稿）
func (s *PostService) ListAdminPosts(req *dto.PostListRequest) ([]model.Post, int64, *BizError) {
	query := s.DB().Model(&model.Post{})
	// 支持按状态筛选（不传返回全部，含草稿）
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("(title LIKE ? OR excerpt LIKE ?)", kw, kw)
	}
	// 按分类筛选
	if req.CategoryID > 0 {
		query = query.Where("category_id = ?", req.CategoryID)
	}
	// 按标签筛选（关联 post_tags 表）
	if req.TagID > 0 {
		query = query.Where("id IN (SELECT post_id FROM post_tags WHERE tag_id = ?)", req.TagID)
	}
	// 按置顶筛选：1 仅置顶 / 0 仅非置顶 / 空为全部
	switch req.IsTop {
	case "1":
		query = query.Where("is_top = ?", 1)
	case "0":
		query = query.Where("is_top = ?", 0)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询文章失败")
	}

	var posts []model.Post
	err := query.
		Omit("content").
		Preload("Category").
		Preload("Tags").
		Order("id DESC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&posts).Error
	if err != nil {
		return nil, 0, NewServerError("查询文章失败")
	}
	return posts, total, nil
}

// CreatePost 创建文章（含标签关联）
func (s *PostService) CreatePost(req *dto.PostUpsertRequest) (*model.Post, *BizError) {
	// 必填校验
	if strings.TrimSpace(req.Title) == "" {
		return nil, NewParamError("文章标题不能为空")
	}

	post := model.Post{
		Title:   req.Title,
		Content: req.Content,
		Status:  PostStatusDraft,
		IsTop:   0,
	}
	// 处理可选字段
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = utils.GenerateSlug(req.Title)
	}
	post.Slug = s.uniqueSlug(slug, 0)
	post.Excerpt = derefStr(req.Excerpt)
	post.CoverImage = derefStr(req.CoverImage)
	if req.CategoryID != nil && *req.CategoryID > 0 {
		post.CategoryID = req.CategoryID
	}
	if req.Status != nil && (*req.Status == PostStatusDraft || *req.Status == PostStatusPublished) {
		post.Status = *req.Status
	}
	if req.IsTop != nil {
		post.IsTop = *req.IsTop
	}
	// 发布状态下补全发布时间
	if post.Status == PostStatusPublished {
		if req.PublishedAt != nil {
			post.PublishedAt = req.PublishedAt
		} else {
			now := time.Now()
			post.PublishedAt = &now
		}
	}

	// 事务创建文章与标签关联。
	// 注意：标签必须在事务外加载 —— 事务已占用数据库连接（SQLite 单连接池），
	// 若在回调内再走 s.DB() 取连接会互相等待形成死锁，导致请求超时并阻塞其他接口。
	tags := s.loadTags(req.TagIDs)
	err := withUniqueSlugRetry(maxSlugRetryAttempts,
		func() error {
			return s.DB().Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(&post).Error; err != nil {
					return err
				}
				// 关联标签
				return tx.Model(&post).Association("Tags").Replace(tags)
			})
		},
		// 唯一键冲突：多为「软删除墓碑仍占用唯一索引」或并发占位，
		// 改用含软删除记录的查重生成新 slug 后重试
		func() { post.Slug = s.uniqueSlugUnscoped(post.Slug, 0) },
	)
	if err != nil {
		return nil, NewServerError("创建文章失败")
	}

	// 重新加载完整数据返回
	s.DB().Preload("Category").Preload("Tags").First(&post, post.ID)
	s.Log().Info("创建文章", zap.Uint("id", post.ID), zap.String("title", post.Title), zap.String("slug", post.Slug))
	return &post, nil
}

// GetAdminPost 管理端文章详情
func (s *PostService) GetAdminPost(id uint) (*model.Post, *BizError) {
	var post model.Post
	if err := s.DB().Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("文章不存在")
		}
		return nil, NewServerError("查询文章失败")
	}
	return &post, nil
}

// UpdatePost 更新文章（含标签关联）
func (s *PostService) UpdatePost(id uint, req *dto.PostUpsertRequest) (*model.Post, *BizError) {
	var post model.Post
	if err := s.DB().First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("文章不存在")
		}
		return nil, NewServerError("查询文章失败")
	}

	updates := map[string]any{}
	// 仅处理请求体中实际提供的字段
	if req.Title != "" {
		post.Title = req.Title
		updates["title"] = req.Title
	}
	if req.Slug != "" {
		post.Slug = s.uniqueSlug(req.Slug, post.ID)
		updates["slug"] = post.Slug
	}
	if req.Content != "" {
		post.Content = req.Content
		updates["content"] = req.Content
	}
	if req.Excerpt != nil {
		post.Excerpt = *req.Excerpt
		updates["excerpt"] = *req.Excerpt
	}
	if req.CoverImage != nil {
		post.CoverImage = *req.CoverImage
		updates["cover_image"] = *req.CoverImage
	}
	if req.CategoryID != nil {
		// 0 表示清空分类
		if *req.CategoryID == 0 {
			post.CategoryID = nil
			updates["category_id"] = nil
		} else {
			post.CategoryID = req.CategoryID
			updates["category_id"] = *req.CategoryID
		}
	}
	if req.Status != nil && (*req.Status == PostStatusDraft || *req.Status == PostStatusPublished) {
		post.Status = *req.Status
		updates["status"] = *req.Status
	}
	if req.IsTop != nil {
		post.IsTop = *req.IsTop
		updates["is_top"] = *req.IsTop
	}
	// 发布时间：仅当状态变为发布且未设置时自动补全
	if req.PublishedAt != nil {
		post.PublishedAt = req.PublishedAt
		updates["published_at"] = *req.PublishedAt
	} else if post.Status == PostStatusPublished && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
		updates["published_at"] = now
	}

	// 更新标签关联（标签在事务外加载，避免单连接池下事务内取连接死锁）
	tags := s.loadTags(req.TagIDs)
	err := withUniqueSlugRetry(maxSlugRetryAttempts,
		func() error {
			return s.DB().Transaction(func(tx *gorm.DB) error {
				if len(updates) > 0 {
					if err := tx.Model(&post).Updates(updates).Error; err != nil {
						return err
					}
				}
				// 更新标签关联
				if req.TagIDs != nil {
					if err := tx.Model(&post).Association("Tags").Replace(tags); err != nil {
						return err
					}
				}
				return nil
			})
		},
		// 唯一键冲突：仅在本次更新涉及 slug 时有重试意义
		func() {
			if req.Slug != "" {
				post.Slug = s.uniqueSlugUnscoped(req.Slug, post.ID)
				updates["slug"] = post.Slug
			}
		},
	)
	if err != nil {
		return nil, NewServerError("更新文章失败")
	}

	// 重新加载完整数据返回
	s.DB().Preload("Category").Preload("Tags").First(&post, post.ID)
	s.Log().Info("更新文章", zap.Uint("id", post.ID), zap.String("title", post.Title), zap.String("slug", post.Slug))
	return &post, nil
}

// DeletePost 删除文章（软删除），同时清理 post_tags 关联与关联评论
func (s *PostService) DeletePost(id uint) *BizError {
	var post model.Post
	if err := s.DB().First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("文章不存在")
		}
		return NewServerError("查询文章失败")
	}

	err := s.DB().Transaction(func(tx *gorm.DB) error {
		// 1. 清理标签关联
		if err := tx.Model(&post).Association("Tags").Clear(); err != nil {
			return err
		}
		// 2. 软删除关联评论
		if err := tx.Where("post_id = ?", post.ID).Delete(&model.Comment{}).Error; err != nil {
			return err
		}
		// 3. 软删除文章
		return tx.Delete(&post).Error
	})
	if err != nil {
		return NewServerError("删除文章失败")
	}
	s.Log().Info("删除文章", zap.Uint("id", post.ID), zap.String("title", post.Title))
	return nil
}

// uniqueSlug 保证文章 slug 唯一，冲突时追加数字后缀。
// 按默认作用域查重（忽略软删除行）：业务友好，正常创建/更新路径使用。
func (s *PostService) uniqueSlug(slug string, excludeID uint) string {
	return s.uniqueSlugWithScope(slug, excludeID, false)
}

// uniqueSlugUnscoped 含软删除记录查重：唯一键冲突重试时使用，
// 绕开「软删除墓碑仍占用唯一索引」导致的重复冲突（见 withUniqueSlugRetry）。
func (s *PostService) uniqueSlugUnscoped(slug string, excludeID uint) string {
	return s.uniqueSlugWithScope(slug, excludeID, true)
}

// uniqueSlugWithScope 通用 slug 查重；includeDeleted=true 时用 Unscoped 含软删除行
func (s *PostService) uniqueSlugWithScope(slug string, excludeID uint, includeDeleted bool) string {
	candidate := slug
	for i := 1; ; i++ {
		var count int64
		query := s.DB().Model(&model.Post{})
		if includeDeleted {
			query = query.Unscoped()
		}
		query = query.Where("slug = ?", candidate)
		if excludeID > 0 {
			query = query.Where("id <> ?", excludeID)
		}
		query.Count(&count)
		if count == 0 {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", slug, i)
	}
}

// loadTags 根据 ID 数组加载标签实体（用于关联）
func (s *PostService) loadTags(ids []uint) []model.Tag {
	if len(ids) == 0 {
		return nil
	}
	var tags []model.Tag
	s.DB().Where("id IN ?", ids).Find(&tags)
	return tags
}

// derefStr 解引用字符串指针，nil 返回空串
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
