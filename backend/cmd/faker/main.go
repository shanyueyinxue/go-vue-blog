package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"blog/internal/app"
	"blog/internal/model"
	"blog/pkg/config"
	"blog/pkg/utils"

	"gorm.io/gorm"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	cfgLoader := config.NewConfig[app.Config](config.ConfigOptions{
		ConfigPath: filepath.Dir(*configPath),
		ConfigType: "yaml",
		ConfigName: strings.TrimSuffix(filepath.Base(*configPath), filepath.Ext(*configPath)),
		EnvPrefix:  "BLOG",
	})
	cfgLoader.SetDefaultMap(map[string]any{
		"server.port": 8090,
		"server.host": "127.0.0.1",
		"app.env":     "development",
		"app.debug":   true,
		"jwt.secret":  "change-me",
	})
	cfg, err := cfgLoader.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	appInstance, err := app.New(cfg)
	if err != nil {
		log.Fatalf("初始化应用失败: %v", err)
	}
	defer appInstance.Cache.Close()

	if err := seed(appInstance.DB); err != nil {
		log.Fatalf("生成 Faker 数据失败: %v", err)
	}
	fmt.Println("Faker 数据填充完成")
}

func seed(db *gorm.DB) error {
	rng := rand.New(rand.NewSource(42))

	categories := []model.Category{
		{Name: "Go", Slug: "go", Description: "Go 语言学习与工程实践", Status: 1, SortOrder: 1},
		{Name: "前端", Slug: "frontend", Description: "Vue、TypeScript 与前端工程化", Status: 1, SortOrder: 2},
		{Name: "工具", Slug: "tools", Description: "效率工具与开发环境", Status: 1, SortOrder: 3},
		{Name: "Python", Slug: "python", Description: "Python 脚本与数据分析", Status: 1, SortOrder: 4},
	}

	tags := []model.Tag{
		{Name: "Go", Slug: "go"},
		{Name: "Vue", Slug: "vue"},
		{Name: "TypeScript", Slug: "typescript"},
		{Name: "Markdown", Slug: "markdown"},
		{Name: "Git", Slug: "git"},
		{Name: "SQL", Slug: "sql"},
		{Name: "Python", Slug: "python"},
		{Name: "效率工具", Slug: "productivity"},
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Post{},
		&model.Category{},
		&model.Tag{},
		&model.PostTag{},
		&model.Comment{},
		&model.SiteConfig{},
		&model.Friend{},
		&model.Work{},
		&model.File{},
	); err != nil {
		return fmt.Errorf("迁移数据表失败: %w", err)
	}

	categoryMap := make(map[string]model.Category, len(categories))
	for i := range categories {
		category := categories[i]
		if err := db.Where("slug = ?", category.Slug).FirstOrCreate(&category).Error; err != nil {
			return err
		}
		categoryMap[category.Slug] = category
	}

	tagMap := make(map[string]model.Tag, len(tags))
	for i := range tags {
		tag := tags[i]
		if err := db.Where("slug = ?", tag.Slug).FirstOrCreate(&tag).Error; err != nil {
			return err
		}
		tagMap[tag.Slug] = tag
	}

	homeImages := []string{
		"https://picsum.photos/seed/blog-1/1920/1080",
		"https://picsum.photos/seed/blog-2/1920/1080",
		"https://picsum.photos/seed/blog-3/1920/1080",
		"https://picsum.photos/seed/blog-4/1920/1080",
		"https://picsum.photos/seed/blog-5/1920/1080",
		"https://picsum.photos/seed/blog-6/1920/1080",
	}

	var activeConfigCount int64
	if err := db.Model(&model.SiteConfig{}).Where("is_active = ?", 1).Count(&activeConfigCount).Error; err != nil {
		return err
	}
	if activeConfigCount == 0 {
		homeImagesJSON, _ := json.Marshal(homeImages)
		cfg := model.SiteConfig{
			Version:           "1.0.0",
			IsActive:          1,
			Title:             "山月印雪",
			Subtitle:          "个人博客",
			Description:       "记录生活、学习、工作中的感悟、心得与经验。",
			ICP:               "示例ICP备00000000号",
			Theme:             "particlex",
			AboutContent:      "## 关于我\n\n这里是自动生成的关于页面内容。",
			HomeImages:        string(homeImagesJSON),
			CommentNeedReview: 1,
			CreatedBy:         "faker",
		}
		if err := db.Create(&cfg).Error; err != nil {
			return err
		}
	}

	friends := []model.Friend{
		{Name: "Vue.js", URL: "https://vuejs.org", Description: "渐进式 JavaScript 框架", Status: 1, SortOrder: 1},
		{Name: "Go 语言", URL: "https://go.dev", Description: "简单、高效、可靠的编程语言", Status: 1, SortOrder: 2},
		{Name: "GitHub", URL: "https://github.com", Description: "代码托管与协作平台", Status: 1, SortOrder: 3},
	}
	for _, friend := range friends {
		if err := db.Where("url = ?", friend.URL).FirstOrCreate(&friend).Error; err != nil {
			return err
		}
	}

	works := []model.Work{
		{
			Name:        "个人博客系统",
			Slug:        "my-blog",
			Description: "基于 Gin + Vue 3 的轻量级博客系统，支持评论审核、动态站点配置与 Markdown 写作。",
			Cover:       "https://picsum.photos/seed/work-1/800/450",
			DemoURL:     "https://blog.example.com",
			RepoURL:     "https://github.com/example/my-blog",
			TechStack:   `["Go","Gin","Vue 3","MySQL"]`,
			Year:        "2025",
			IsTop:       1,
			Status:      1,
			SortOrder:   1,
		},
		{
			Name:        "TODO 任务管理",
			Slug:        "todo-app",
			Description: "一个极简的 TODO 任务管理应用，支持拖拽排序与本地存储。",
			Cover:       "https://picsum.photos/seed/work-2/800/450",
			DemoURL:     "https://todo.example.com",
			ArticleURL:  "https://blog.example.com/post/todo-app",
			RepoURL:     "https://github.com/example/todo-app",
			TechStack:   `["Vue 3","TypeScript","Vite"]`,
			Year:        "2024",
			Status:      1,
			SortOrder:   2,
		},
	}
	for _, work := range works {
		if err := db.Where("slug = ?", work.Slug).FirstOrCreate(&work).Error; err != nil {
			return err
		}
	}

	postSeeds := []struct {
		title      string
		excerpt    string
		content    string
		category   string
		tags       []string
		published  time.Time
		viewCount  int
		commentNum int
	}{
		{
			title:      "Go 多版本管理工具",
			excerpt:    "Go 多版本管理工具介绍。",
			content:    "## 为什么需要多版本\n\n不同项目可能依赖不同的 Go 版本。\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```",
			category:   "go",
			tags:       []string{"go"},
			published:  time.Now().AddDate(0, -2, 0),
			viewCount:  120,
			commentNum: 2,
		},
		{
			title:      "Vue 3 组合式 API 实践",
			excerpt:    "用组合式 API 组织可复用的业务逻辑。",
			content:    "## 组合式 API\n\n`setup` 让我们可以更自然地拆分逻辑。\n\n```ts\nconst count = ref(0)\nconst double = computed(() => count.value * 2)\n```",
			category:   "frontend",
			tags:       []string{"vue", "typescript"},
			published:  time.Now().AddDate(0, -4, 0),
			viewCount:  210,
			commentNum: 1,
		},
		{
			title:      "Markdown 写作与渲染",
			excerpt:    "Markdown 基本语法以及前端安全渲染。",
			content:    "## Markdown\n\n- 列表\n- 代码块\n- 表格\n\n| 字段 | 说明 |\n| --- | --- |\n| title | 标题 |",
			category:   "tools",
			tags:       []string{"markdown", "vue"},
			published:  time.Now().AddDate(0, -6, 0),
			viewCount:  180,
			commentNum: 0,
		},
		{
			title:      "Git 常用工作流",
			excerpt:    "从提交、分支到合并的常见 Git 工作流。",
			content:    "## Git 工作流\n\n```bash\ngit switch -c feature/foo\ngit commit -am \"feat: add foo\"\ngit push -u origin feature/foo\n```",
			category:   "tools",
			tags:       []string{"git"},
			published:  time.Now().AddDate(0, -8, 0),
			viewCount:  96,
			commentNum: 1,
		},
		{
			title:      "Python 推导式",
			excerpt:    "介绍 Python 中列表、字典和集合推导式。",
			content:    "## 列表推导式\n\n```python\nsquares = [x * x for x in range(10)]\n```",
			category:   "python",
			tags:       []string{"python"},
			published:  time.Now().AddDate(0, -10, 0),
			viewCount:  75,
			commentNum: 0,
		},
		{
			title:      "SQL 查询优化清单",
			excerpt:    "从索引、执行计划到 N+1 问题。",
			content:    "## 索引\n\n为高频过滤字段建立合适索引。\n\n## N+1\n\n尽量使用 JOIN 或批量查询。",
			category:   "tools",
			tags:       []string{"sql", "go"},
			published:  time.Now().AddDate(0, -12, 0),
			viewCount:  140,
			commentNum: 2,
		},
	}

	for i := 0; i < 10; i++ {
		postSeeds = append(postSeeds, (struct {
			title      string
			excerpt    string
			content    string
			category   string
			tags       []string
			published  time.Time
			viewCount  int
			commentNum int
		}{
			title:      fmt.Sprintf("title %d", i),
			excerpt:    fmt.Sprintf("excerpt %d", i),
			content:    "## 为什么需要多版本\n\n不同项目可能依赖不同的 Go 版本。\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```",
			category:   "tools",
			tags:       []string{"tools"},
			published:  time.Now().AddDate(0, -2, 0),
			viewCount:  120,
			commentNum: 0,
		}))
	}

	for i, seedPost := range postSeeds {
		var existing int64
		slug := utils.GenerateSlug(seedPost.title)
		if err := db.Model(&model.Post{}).Where("slug = ?", slug).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			continue
		}

		category := categoryMap[seedPost.category]
		post := model.Post{
			Title:       seedPost.title,
			Slug:        slug,
			Content:     seedPost.content,
			Excerpt:     seedPost.excerpt,
			Status:      "published",
			IsTop:       boolToInt(i == 0),
			ViewCount:   seedPost.viewCount,
			PublishedAt: &seedPost.published,
		}
		if category.ID != 0 {
			post.CategoryID = &category.ID
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&post).Error; err != nil {
				return err
			}
			tagModels := make([]model.Tag, 0, len(seedPost.tags))
			for _, tagSlug := range seedPost.tags {
				if tag, ok := tagMap[tagSlug]; ok {
					tagModels = append(tagModels, tag)
				}
			}
			return tx.Model(&post).Association("Tags").Replace(tagModels)
		}); err != nil {
			return err
		}

		for j := 0; j < seedPost.commentNum; j++ {
			status := model.CommentStatusApproved
			if rng.Intn(3) == 0 {
				status = model.CommentStatusPending
			}
			now := time.Now().Add(-time.Duration(rng.Intn(72)) * time.Hour)
			comment := model.Comment{
				PostID:     post.ID,
				ParentID:   0,
				Nickname:   fmt.Sprintf("访客%d", j+1),
				Email:      fmt.Sprintf("visitor%d@example.com", j+1),
				Content:    fmt.Sprintf("第 %d 条自动生成评论：文章写得不错。", j+1),
				Status:     status,
				IsAuthor:   0,
				IP:         "127.0.0.1",
				UserAgent:  "Faker/1.0",
				ReviewedAt: &now,
			}
			if err := db.Create(&comment).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
