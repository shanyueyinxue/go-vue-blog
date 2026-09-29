package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"blog/internal/app"
	"blog/internal/dto"
	"blog/internal/model"
	"blog/pkg/cache"
	"blog/pkg/captcha"
	"blog/pkg/email"
	"blog/pkg/response"
	"blog/pkg/storage"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// emailStub 邮件服务桩：记录但不实际发送
type emailStub struct{}

func (emailStub) Init(*email.EmailConfig) error  { return nil }
func (emailStub) SetHeader(key, value string)    {}
func (emailStub) Dial() error                    { return nil }
func (emailStub) Close() error                   { return nil }
func (emailStub) IsClosed() bool                 { return false }
func (emailStub) Send(*email.Email) error        { return nil }
func (emailStub) DialAndSend(*email.Email) error { return nil }

// newTestApp 构建最小可用的 App（sqlite 内存库 + 内存缓存 + dev 验证码 + 本地临时存储）
func newTestApp(t *testing.T) *app.App {
	t.Helper()
	// 每个测试使用独立的具名内存库（cache=shared 模式下若 DSN 相同会共享同一数据库，
	// 导致测试间数据互相污染，如配置 ID、文章等绝对引用错位）
	dsn := fmt.Sprintf("file:mem_%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	require.Nil(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.Post{},
		&model.Category{},
		&model.Tag{},
		&model.PostTag{},
		&model.Comment{},
		&model.SiteConfig{},
		&model.Friend{},
		&model.File{},
		&model.Music{},
	))

	logger := zap.NewNop()
	cacheIns := cache.NewMemoryCache(cache.MemoryConfig{})
	captchaMgr, err := captcha.NewCaptchaManager(captcha.NewDevProvider(), 300, captcha.ManagerOptions{
		Cache:  cacheIns,
		Logger: logger,
	})
	require.Nil(t, err)

	return &app.App{
		Config: &app.Config{
			App:     app.AppConfig{Env: "test", Debug: false, Timezone: "Asia/Shanghai"},
			Captcha: captcha.CaptchaConfig{Length: 4, Expiry: 300, Option: "dev"},
		},
		Logger:  logger,
		DB:      db,
		Cache:   cacheIns,
		JWT:     nil,
		Email:   emailStub{},
		Captcha: captchaMgr,
		Storage: storage.NewLocalStorage(t.TempDir(), "uploads"),
	}
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

// TestCommentCaptchaFlow 评论验证码流程：错误验证码拒绝、正确验证码通过、关闭验证码可免验证
func TestCommentCaptchaFlow(t *testing.T) {
	a := newTestApp(t)
	base := New(a)
	configSvc := NewConfigService(base)
	captchaSvc := NewCaptchaService(base)
	commentSvc := NewCommentService(base, configSvc, captchaSvc)

	now := time.Now()
	post := model.Post{Title: "测试文章", Slug: "captcha-post", Status: PostStatusPublished, PublishedAt: &now}
	require.NoError(t, a.DB.Create(&post).Error)

	// 站点配置：评论需要验证码、无需审核
	_, err := configSvc.CreateConfig(&dto.ConfigUpsertRequest{
		Version:            "1.0.0",
		Title:              strPtr("测试站点"),
		CommentNeedReview:  intPtr(0),
		CommentNeedCaptcha: intPtr(1),
	})
	require.Nil(t, err)

	emailAddr := "visitor@example.com"
	token, _, err := captchaSvc.SendEmailCaptcha(&dto.EmailCaptchaRequest{Email: emailAddr, Scene: "comment"})
	require.Nil(t, err)
	require.NotEmpty(t, token)

	// 错误验证码被拒绝
	_, _, bizErr := commentSvc.CreateComment(&dto.CreateCommentRequest{
		Slug: post.Slug, Nickname: "访客", Email: emailAddr, Content: "你好",
		CaptchaToken: token, CaptchaCode: "0000",
	}, "10.0.0.1", "test-ua")
	require.NotNil(t, bizErr)

	// 正确验证码（dev provider 固定 1234）通过，且直接审核通过
	comment, status, bizErr := commentSvc.CreateComment(&dto.CreateCommentRequest{
		Slug: post.Slug, Nickname: "访客", Email: emailAddr, Content: "你好",
		CaptchaToken: token, CaptchaCode: "1234",
	}, "10.0.0.2", "test-ua")
	require.Nil(t, bizErr)
	require.Equal(t, model.CommentStatusApproved, status)
	require.NotNil(t, comment)
	require.NotZero(t, comment.ID)

	// 关闭验证码后无需验证码即可提交
	_, err = configSvc.UpdateConfig(1, &dto.ConfigUpsertRequest{CommentNeedCaptcha: intPtr(0)})
	require.Nil(t, err)
	_, status2, bizErr := commentSvc.CreateComment(&dto.CreateCommentRequest{
		Slug: post.Slug, Nickname: "访客2", Email: "visitor2@example.com", Content: "无需验证码",
	}, "10.0.0.3", "test-ua")
	require.Nil(t, bizErr)
	require.Equal(t, model.CommentStatusApproved, status2)
}

// TestConfigActivateAndClear 配置激活流转 + *string 清空语义
func TestConfigActivateAndClear(t *testing.T) {
	a := newTestApp(t)
	configSvc := NewConfigService(New(a))

	v1, err := configSvc.CreateConfig(&dto.ConfigUpsertRequest{
		Version: "1.0.0", Title: strPtr("标题A"), Description: strPtr("描述A"),
	})
	if err != nil {
		t.Fatalf("CreateConfig err: %#v (%T)", err, err)
	}
	require.Nil(t, err)
	require.Equal(t, 1, v1.IsActive)

	// 创建新版本自动激活并取消旧版本
	v2, err := configSvc.CreateConfig(&dto.ConfigUpsertRequest{Version: "2.0.0", Title: strPtr("标题B")})
	require.Nil(t, err)
	require.Equal(t, 1, v2.IsActive)

	var check model.SiteConfig
	require.NoError(t, a.DB.First(&check, v1.ID).Error)
	require.Equal(t, 0, check.IsActive)

	// 激活旧版本
	require.Nil(t, configSvc.ActivateConfig(v1.ID))
	act := configSvc.ActiveConfig()
	require.NotNil(t, act)
	require.Equal(t, "标题A", act.Title)

	// 空字符串清除字段（*string 语义）
	_, err = configSvc.UpdateConfig(v1.ID, &dto.ConfigUpsertRequest{Description: strPtr("")})
	require.Nil(t, err)
	require.NoError(t, a.DB.First(&check, v1.ID).Error)
	require.Equal(t, "", check.Description)

	// nil 表示不修改该字段
	_, err = configSvc.UpdateConfig(v1.ID, &dto.ConfigUpsertRequest{Title: strPtr("标题A2")})
	require.Nil(t, err)
	require.NoError(t, a.DB.First(&check, v1.ID).Error)
	require.Equal(t, "标题A2", check.Title)
	require.Equal(t, "", check.Description) // 未传 description，保持为空
}

// 最小合法 PDF 内容（http.DetectContentType 识别为 application/pdf，且不会触发缩略图异步任务）
var tinyPDF = []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

// TestFileUploadAndDelete 文件上传/删除：删除时同时清理缩略图
func TestFileUploadAndDelete(t *testing.T) {
	a := newTestApp(t)
	fileSvc := NewFileService(New(a))

	// 构造 multipart 上传
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("file", "doc.pdf")
	require.Nil(t, err)
	_, err = fw.Write(tinyPDF)
	require.Nil(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, header, err := req.FormFile("file")
	require.Nil(t, err)

	file, err := fileSvc.Upload(header, "", 1)
	require.Nil(t, err)
	require.NotEmpty(t, file.Path)

	// 模拟已生成的缩略图：写入存储并关联到记录
	thumbStored, err := a.Storage.Save(context.Background(), "thumb_doc.png", []byte("thumb"), "thumbs")
	require.Nil(t, err)
	require.NoError(t, a.DB.Model(file).Update("thumbnail", thumbStored).Error)

	// 删除文件：主文件与缩略图都应被清理
	require.Nil(t, fileSvc.DeleteFile(file.ID))

	_, err = a.Storage.GetFile(context.Background(), file.Path)
	require.Error(t, err, "主文件应已被删除")
	_, err = a.Storage.GetFile(context.Background(), thumbStored)
	require.Error(t, err, "缩略图应一并被删除")

	// 数据库记录已物理删除（查询不到）
	var count int64
	require.NoError(t, a.DB.Model(&model.File{}).Where("id = ?", file.ID).Count(&count).Error)
	require.Zero(t, count)
}

// TestCreatePostWithTagsNoDeadlock 回归测试：SQLite 单连接池（maxOpenConns=1）下
// 带标签创建/更新文章不得死锁（修复前 loadTags 在事务回调内用 s.DB() 取连接会互相等待）。
func TestCreatePostWithTagsNoDeadlock(t *testing.T) {
	a := newTestApp(t)
	sqlDB, err := a.DB.DB()
	require.NoError(t, err)
	// 模拟生产 SQLite 配置的 maxOpenConns: 1
	sqlDB.SetMaxOpenConns(1)

	base := New(a)
	postSvc := NewPostService(base)

	tagA := model.Tag{Name: "Go", Slug: "go"}
	tagB := model.Tag{Name: "Vue", Slug: "vue"}
	require.NoError(t, a.DB.Create(&tagA).Error)
	require.NoError(t, a.DB.Create(&tagB).Error)

	// 带标签创建文章，超时保护：回归死锁时测试失败而不是挂起
	created := make(chan *BizError, 1)
	go func() {
		_, bizErr := postSvc.CreatePost(&dto.PostUpsertRequest{
			Title:   "带标签的文章",
			Content: "正文",
			TagIDs:  []uint{tagA.ID, tagB.ID},
		})
		created <- bizErr
	}()
	select {
	case bizErr := <-created:
		require.Nil(t, bizErr)
	case <-time.After(10 * time.Second):
		t.Fatal("CreatePost 带标签时死锁/超时（事务回调内取连接）")
	}

	// 标签关联成功
	var post model.Post
	require.NoError(t, a.DB.Where("title = ?", "带标签的文章").First(&post).Error)
	var tags []model.Tag
	require.NoError(t, a.DB.Model(&post).Association("Tags").Find(&tags))
	require.Len(t, tags, 2)

	// 更新标签同样不得死锁
	updated := make(chan *BizError, 1)
	go func() {
		_, bizErr := postSvc.UpdatePost(post.ID, &dto.PostUpsertRequest{TagIDs: []uint{tagA.ID}})
		updated <- bizErr
	}()
	select {
	case bizErr := <-updated:
		require.Nil(t, bizErr)
	case <-time.After(10 * time.Second):
		t.Fatal("UpdatePost 更新标签时死锁/超时（事务回调内取连接）")
	}
	var tags2 []model.Tag
	require.NoError(t, a.DB.Model(&post).Association("Tags").Find(&tags2))
	require.Len(t, tags2, 1)
	require.Equal(t, tagA.ID, tags2[0].ID)
}

// TestCommentNicknameValidation 评论昵称 trim 与长度校验
func TestCommentNicknameValidation(t *testing.T) {
	a := newTestApp(t)
	base := New(a)
	configSvc := NewConfigService(base)
	captchaSvc := NewCaptchaService(base)
	commentSvc := NewCommentService(base, configSvc, captchaSvc)

	// 关闭验证码，专注昵称校验
	_, err := configSvc.CreateConfig(&dto.ConfigUpsertRequest{
		Version: "1.0.0", CommentNeedCaptcha: intPtr(0),
	})
	require.Nil(t, err)

	now := time.Now()
	post := model.Post{Title: "t", Slug: "nickname-post", Status: PostStatusPublished, PublishedAt: &now}
	require.NoError(t, a.DB.Create(&post).Error)

	// 全空格昵称被拒绝
	_, _, bizErr := commentSvc.CreateComment(&dto.CreateCommentRequest{
		Slug: post.Slug, Nickname: "   ", Email: "a@example.com", Content: "内容",
	}, "10.0.0.10", "ua")
	require.NotNil(t, bizErr)

	// 超长昵称（>50 字符）被拒绝，而不是触发数据库错误
	longNick := strings.Repeat("名", 51)
	_, _, bizErr = commentSvc.CreateComment(&dto.CreateCommentRequest{
		Slug: post.Slug, Nickname: longNick, Email: "a@example.com", Content: "内容",
	}, "10.0.0.11", "ua")
	require.NotNil(t, bizErr)

	// 正常昵称首尾空格被 trim 后入库
	comment, _, bizErr := commentSvc.CreateComment(&dto.CreateCommentRequest{
		Slug: post.Slug, Nickname: "  访客  ", Email: "a@example.com", Content: "内容",
	}, "10.0.0.12", "ua")
	require.Nil(t, bizErr)
	require.Equal(t, "访客", comment.Nickname)
}

// TestUploadDocxMimeMapping docx（zip 容器）按扩展名映射 MIME，通过白名单校验
func TestUploadDocxMimeMapping(t *testing.T) {
	a := newTestApp(t)
	fileSvc := NewFileService(New(a))

	// 真实 docx 即 zip 容器（PK 魔数开头），http.DetectContentType 只能识别为 application/zip
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("file", "report.docx")
	require.NoError(t, err)
	_, err = fw.Write([]byte("PK\x03\x04\x14\x00\x00\x00\x08\x00\x00\x00\x00\x00"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, header, err := req.FormFile("file")
	require.NoError(t, err)

	file, err := fileSvc.Upload(header, "", 1)
	require.Nil(t, err)
	require.Equal(t, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", file.MimeType)
}

// TestPostSlugReuseAfterSoftDelete 回归测试（方案 C）：软删除记录仍占用唯一索引时，
// 同名 slug 重建应通过「唯一键冲突重试」自动追加后缀，而不是返回 500。
// 修复前：uniqueSlug 按默认作用域查重（忽略软删除行）→ 认为 slug 可用 →
// INSERT/UPDATE 触发唯一索引冲突 → 500；修复后：捕获 gorm.ErrDuplicatedKey 重试，
// 改用含软删除记录的 Unscoped 查重生成 slug-N。
func TestPostSlugReuseAfterSoftDelete(t *testing.T) {
	a := newTestApp(t)
	postSvc := NewPostService(New(a))

	// 1. 创建 slug=hello 的文章并软删除（墓碑占用唯一索引的 hello）
	p1, err := postSvc.CreatePost(&dto.PostUpsertRequest{Title: "Hello", Slug: "hello", Status: strPtr("published")})
	require.Nil(t, err)
	require.Equal(t, "hello", p1.Slug)
	require.Nil(t, postSvc.DeletePost(p1.ID))

	// 2. 创建路径：再以相同 slug 创建，应自动生成为 hello-1
	p2, err := postSvc.CreatePost(&dto.PostUpsertRequest{Title: "Hello Again", Slug: "hello", Status: strPtr("published")})
	require.Nil(t, err)
	require.Equal(t, "hello-1", p2.Slug)

	// 3. 更新路径：把 p2 改名到被墓碑占用的 slug world，应自动落为 world-1
	p3, err := postSvc.CreatePost(&dto.PostUpsertRequest{Title: "World", Slug: "world", Status: strPtr("published")})
	require.Nil(t, err)
	require.Nil(t, postSvc.DeletePost(p3.ID))
	p4, err := postSvc.UpdatePost(p2.ID, &dto.PostUpsertRequest{Slug: "world"})
	require.Nil(t, err)
	require.Equal(t, "world-1", p4.Slug)
}

// 1x1 透明 PNG（http.DetectContentType 识别为 image/png）
var tinyPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

// TestUploadTextMimeExtensionWhitelist 文本类 MIME 扩展名白名单：
// 文本内容（检测为 text/plain）不允许携带可执行扩展名（.html/.js/.svg 等），
// 只允许 .txt/.md 等安全扩展名，防止同源存储型 XSS。
func TestUploadTextMimeExtensionWhitelist(t *testing.T) {
	a := newTestApp(t)
	fileSvc := NewFileService(New(a))

	upload := func(name string, content []byte) (*model.File, error) {
		t.Helper()
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		fw, err := w.CreateFormFile("file", name)
		require.NoError(t, err)
		_, err = fw.Write(content)
		require.NoError(t, err)
		require.NoError(t, w.Close())

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", w.FormDataContentType())
		_, header, err := req.FormFile("file")
		require.NoError(t, err)
		return fileSvc.Upload(header, "", 1)
	}

	textContent := []byte("hello, this is plain text")

	// 1. 纯文本内容 + .html 扩展名 → 拒绝（可执行扩展名）
	_, err := upload("evil.html", textContent)
	require.Error(t, err)
	require.Contains(t, err.Error(), "扩展名不允许")

	// 2. 纯文本内容 + .js / .svg → 拒绝
	_, err = upload("evil.js", textContent)
	require.Error(t, err)
	_, err = upload("evil.svg", textContent)
	require.Error(t, err)

	// 3. 纯文本内容 + 安全扩展名 → 通过
	f1, err := upload("note.txt", textContent)
	require.Nil(t, err)
	require.Equal(t, "text/plain", f1.MimeType)

	// 4. Markdown 内容（魔数检测同为 text/plain）+ .md → 通过
	f2, err := upload("post.md", []byte("# 标题\n正文"))
	require.Nil(t, err)
	require.Equal(t, "text/plain", f2.MimeType)

	// 5. 真实 HTML 内容（魔数检测为 text/html）→ 类型白名单直接拒绝
	_, err = upload("page.html", []byte("<!DOCTYPE html><html><body>hi</body></html>"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "不支持的文件类型")
}

// TestUploadFolderByMime 未指定目录时按 MIME 类型自动分目录；显式目录优先
func TestUploadFolderByMime(t *testing.T) {
	a := newTestApp(t)
	fileSvc := NewFileService(New(a))

	upload := func(name string, content []byte) string {
		t.Helper()
		body := &bytes.Buffer{}
		w := multipart.NewWriter(body)
		fw, err := w.CreateFormFile("file", name)
		require.NoError(t, err)
		_, err = fw.Write(content)
		require.NoError(t, err)
		require.NoError(t, w.Close())

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", w.FormDataContentType())
		_, header, err := req.FormFile("file")
		require.NoError(t, err)

		file, err := fileSvc.Upload(header, "", 1)
		require.Nil(t, err)
		return file.Path
	}

	// 图片 → images/
	pngPath := upload("pic.png", tinyPNG)
	require.True(t, strings.HasPrefix(pngPath, "images/"), "图片应存到 images/，实际: %s", pngPath)

	// 文档 → files/
	pdfPath := upload("doc.pdf", tinyPDF)
	require.True(t, strings.HasPrefix(pdfPath, "files/"), "文档应存到 files/，实际: %s", pdfPath)

	// 显式指定目录时不自动分目录
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", "explicit.pdf")
	_, _ = fw.Write(tinyPDF)
	require.NoError(t, w.Close())
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, header, err := req.FormFile("file")
	require.NoError(t, err)
	file, err := fileSvc.Upload(header, "custom", 1)
	require.Nil(t, err)
	require.True(t, strings.HasPrefix(file.Path, "custom/"), "显式目录应保留，实际: %s", file.Path)
}

// TestBackupRun 数据库备份：导出 CSV → zip → storage 保存，内容可读且包含各表数据
func TestBackupRun(t *testing.T) {
	a := newTestApp(t)
	backupSvc := NewBackupService(New(a))

	// 造一点数据
	now := time.Now()
	post := model.Post{Title: "备份测试文章", Slug: "backup-post", Status: PostStatusPublished, PublishedAt: &now}
	require.NoError(t, a.DB.Create(&post).Error)

	path, err := backupSvc.RunBackup()
	require.Nil(t, err)
	require.True(t, strings.HasPrefix(path, "backups/"), "默认应保存到 backups/，实际: %s", path)

	// 配置自定义备份目录时按配置保存
	a.Config.Backup.Dir = "data/backups"
	path2, err := backupSvc.RunBackup()
	require.Nil(t, err)
	require.True(t, strings.HasPrefix(path2, "data/backups/"), "应保存到配置的目录 data/backups/，实际: %s", path2)

	// 从 storage 读回压缩包并校验内容
	rc, err := a.Storage.GetFile(context.Background(), path)
	require.NoError(t, err)
	defer rc.Close()

	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	// 压缩包包含全部表的 CSV
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	for _, want := range []string{
		"users.csv", "categories.csv", "tags.csv", "post_tags.csv",
		"posts.csv", "comments.csv", "site_configs.csv", "friends.csv", "files.csv",
	} {
		require.True(t, names[want], "压缩包应包含 %s", want)
	}

	// posts.csv 包含我们创建的文章（表头 + 数据行）
	var postsCSV []byte
	for _, f := range zr.File {
		if f.Name == "posts.csv" {
			r, err := f.Open()
			require.NoError(t, err)
			postsCSV, err = io.ReadAll(r)
			require.NoError(t, err)
			r.Close()
			break
		}
	}
	require.Contains(t, string(postsCSV), "title")
	require.Contains(t, string(postsCSV), "备份测试文章")
}

// TestBackupRunMasking 备份脱敏与文件名复杂度：
// - 密码哈希、邮箱、手机号、IP、UA 在 CSV 中不可见原文；
// - 文件名含 32 位随机十六进制后缀，防止被遍历/撞库猜中。
func TestBackupRunMasking(t *testing.T) {
	a := newTestApp(t)
	backupSvc := NewBackupService(New(a))

	// 造带敏感字段的数据
	now := time.Now()
	user := model.User{
		Username: "master",
		Password: "$2a$10$abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMN",
		Email:    "master@example.com",
		Phone:    "13800138000",
		LastIP:   "203.0.113.7",
	}
	require.NoError(t, a.DB.Create(&user).Error)

	post := model.Post{Title: "脱敏测试", Slug: "mask-post", Status: PostStatusPublished, PublishedAt: &now}
	require.NoError(t, a.DB.Create(&post).Error)
	comment := model.Comment{
		PostID:    post.ID,
		Nickname:  "访客",
		Email:     "visitor@example.com",
		Phone:     "13900139000",
		Content:   "不错",
		Status:    model.CommentStatusApproved,
		IP:        "198.51.100.9",
		UserAgent: "Mozilla/5.0 SecretUA",
	}
	require.NoError(t, a.DB.Create(&comment).Error)

	path, err := backupSvc.RunBackup()
	require.Nil(t, err)
	// 实际落盘文件名由 storage.generateKey 生成：时间戳 + 24 位随机十六进制（crypto/rand，防撞库）
	require.Regexp(t, `^backups/\d+_[0-9a-f]{24}\.zip$`, path, "备份文件名应带高熵随机后缀: %s", path)

	rc, err := a.Storage.GetFile(context.Background(), path)
	require.NoError(t, err)
	defer rc.Close()

	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	readCSV := func(name string) string {
		t.Helper()
		for _, f := range zr.File {
			if f.Name == name {
				r, err := f.Open()
				require.NoError(t, err)
				b, err := io.ReadAll(r)
				require.NoError(t, err)
				r.Close()
				return string(b)
			}
		}
		t.Fatalf("压缩包缺少 %s", name)
		return ""
	}

	usersCSV := readCSV("users.csv")
	require.NotContains(t, usersCSV, "$2a$10$abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMN", "密码哈希必须脱敏")
	require.Contains(t, usersCSV, "***", "密码字段应整体脱敏")
	require.NotContains(t, usersCSV, "master@example.com", "邮箱必须脱敏")
	require.Contains(t, usersCSV, "m***@example.com", "邮箱应部分脱敏")
	require.NotContains(t, usersCSV, "13800138000", "手机号必须脱敏")
	require.Contains(t, usersCSV, "138****8000", "手机号应部分脱敏")
	require.NotContains(t, usersCSV, "203.0.113.7", "IP 必须脱敏")
	require.Contains(t, usersCSV, "203.0.*.*", "IP 应部分脱敏")

	commentsCSV := readCSV("comments.csv")
	require.NotContains(t, commentsCSV, "visitor@example.com", "评论邮箱必须脱敏")
	require.NotContains(t, commentsCSV, "13900139000", "评论手机号必须脱敏")
	require.NotContains(t, commentsCSV, "198.51.100.9", "评论 IP 必须脱敏")
	require.NotContains(t, commentsCSV, "SecretUA", "评论 UA 必须脱敏")
}

// TestPublicCommentsHidePrivacyFields 公开评论接口经 PublicComment DTO 输出，
// JSON 序列化后不含 email/phone/ip/user_agent 等隐私字段。
func TestPublicCommentsHidePrivacyFields(t *testing.T) {
	a := newTestApp(t)
	now := time.Now()
	post := model.Post{Title: "评论隐私", Slug: "privacy-post", Status: PostStatusPublished, PublishedAt: &now}
	require.NoError(t, a.DB.Create(&post).Error)

	parent := model.Comment{
		PostID: post.ID, Nickname: "父评论", Email: "parent@example.com",
		Phone: "13700137000", Content: "顶层", Status: model.CommentStatusApproved,
		IP: "192.0.2.1", UserAgent: "UA-Parent",
	}
	require.NoError(t, a.DB.Create(&parent).Error)
	child := model.Comment{
		PostID: post.ID, ParentID: parent.ID, Nickname: "子评论", Email: "child@example.com",
		Content: "回复", Status: model.CommentStatusApproved, IP: "192.0.2.2", UserAgent: "UA-Child",
	}
	require.NoError(t, a.DB.Create(&child).Error)

	svc := NewCommentService(New(a), NewConfigService(New(a)), nil)
	list, total, bizErr := svc.ListComments("privacy-post", &dto.GetCommentsRequest{
		PageRequest: response.PageRequest{Page: 1, PageSize: 10},
	})
	require.Nil(t, bizErr)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.Len(t, list[0].Replies, 1)
	require.Equal(t, "父评论", list[0].Nickname)
	require.Equal(t, "子评论", list[0].Replies[0].Nickname)

	// JSON 序列化后不得包含隐私字段（字段值或字段名）
	raw, err := json.Marshal(list)
	require.NoError(t, err)
	body := string(raw)
	require.NotContains(t, body, "parent@example.com")
	require.NotContains(t, body, "child@example.com")
	require.NotContains(t, body, "13700137000")
	require.NotContains(t, body, "192.0.2.1")
	require.NotContains(t, body, "UA-Parent")
	require.NotContains(t, body, `"email"`)
	require.NotContains(t, body, `"ip"`)
	require.NotContains(t, body, `"user_agent"`)
}
