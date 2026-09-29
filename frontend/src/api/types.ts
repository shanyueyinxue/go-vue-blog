/**
 * API 类型定义
 *
 * 与后端（`backend/internal/model`、`backend/internal/dto`）及
 * `docs/api.md` 文档保持一致。字段命名采用后端 JSON 返回的
 * snake_case 风格。
 */

import type { AxiosRequestConfig } from 'axios'

/* -------------------------------------------------------------------------- */
/*                                通用类型与常量                                */
/* -------------------------------------------------------------------------- */

/** 后端统一响应结构（docs/api.md 1.3） */
export interface ApiResponse<T = unknown> {
  /** 错误码，0 表示成功 */
  code: number
  /** 提示信息，成功时为 "ok" */
  message: string
  /** 业务数据，可为 null */
  data: T
}

/** 分页响应结构（docs/api.md 1.5） */
export interface PageResult<T> {
  /** 当前页数据 */
  list: T[]
  /** 总条数 */
  total: number
  /** 当前页码（从 1 开始） */
  page: number
  /** 每页条数 */
  pageSize: number
}

/** 分页查询参数（docs/api.md 1.5） */
export interface PageParams {
  /** 页码，从 1 开始，默认 1 */
  page?: number
  /** 每页条数，默认 10，最大 100 */
  pageSize?: number
}

/** 错误码常量（docs/api.md 1.4，与 backend/pkg/response 一致） */
export const ApiCode = {
  /** 成功 */
  Success: 0,
  /** 请求参数错误 */
  ParamError: 40000,
  /** 未认证或 Token 已过期/无效 */
  Unauthorized: 40100,
  /** 无权限（非管理员） */
  Forbidden: 40300,
  /** 资源不存在 */
  NotFound: 40400,
  /** 资源冲突（如 slug 已存在） */
  Conflict: 40900,
  /** 请求过于频繁（限流） */
  TooManyRequests: 42900,
  /** 服务器内部错误 */
  ServerError: 50000,
} as const

/** 文章状态（docs/api.md 1.7） */
export type PostStatus = 'draft' | 'published'

/** 评论状态（docs/api.md 1.7） */
export type CommentStatus = 'pending' | 'approved' | 'rejected'

/** 文章排序方式（docs/api.md 4.2） */
export type PostSort = 'latest' | 'views' | 'oldest'

/** 验证码场景（docs/api.md 12.1） */
export type CaptchaScene = 'comment' | 'login' | 'forgot'

/* -------------------------------------------------------------------------- */
/*                                   管理员用户                                 */
/* -------------------------------------------------------------------------- */

/** 管理员用户信息（backend/internal/model/user.go，不含密码） */
export interface User {
  id: number
  /** 登录名（唯一） */
  username: string
  /** 邮箱（可空） */
  email: string
  /** 手机号（可空） */
  phone: string
  /** 头像地址 */
  avatar: string
  /** 是否超级管理员：1 是 / 0 否 */
  is_master: number
  /** 角色，如 admin */
  role: string
  /** 状态：1 启用 / 0 禁用 */
  status: number
  /** 最近登录 IP */
  last_ip: string
  /** 最近登录时间（RFC3339） */
  last_login_at: string | null
  created_at: string
  updated_at: string
}

/** 修改个人信息请求体（PUT /api/admin/profile，仅允许修改基础资料） */
export interface ProfileUpdateParams {
  /** 登录名（留空表示不修改；不能传空串） */
  username?: string
  /** 邮箱（传空串表示清除） */
  email?: string
  /** 手机号（传空串表示清除） */
  phone?: string
  /** 头像地址（传空串表示清除） */
  avatar?: string
}

/** 忘记密码-发送邮箱验证码请求体（POST /api/forgot-password/send-code） */
export interface ForgotPasswordSendParams {
  /** 注册邮箱 */
  email: string
}

/** 忘记密码-重置密码请求体（POST /api/forgot-password/reset） */
export interface ForgotPasswordResetParams {
  /** 注册邮箱 */
  email: string
  /** 发送验证码时返回的 token */
  token: string
  /** 邮箱验证码 */
  code: string
  /** 新密码（≥ 8 位） */
  new_password: string
}

/* -------------------------------------------------------------------------- */
/*                                 用户管理（仅超级管理员）                      */
/* -------------------------------------------------------------------------- */

/** 用户管理列表查询参数（GET /api/admin/users） */
export interface AdminUserQueryParams extends PageParams {
  /** 用户名/邮箱关键字模糊搜索 */
  keyword?: string
}

/** 创建用户请求体（POST /api/admin/users，仅 master） */
export interface UserCreateParams {
  /** 登录名（必填，唯一） */
  username: string
  /** 初始密码（必填，≥ 8 位） */
  password: string
  /** 邮箱（可空） */
  email?: string
  /** 手机号（可空） */
  phone?: string
  /** 头像地址（可空） */
  avatar?: string
  /** 角色，默认 admin */
  role?: string
  /** 状态：1 启用 / 0 禁用，默认 1 */
  status?: number
  /** 是否超级管理员：1 / 0，默认 0 */
  is_master?: number
}

/** 更新用户请求体（PUT /api/admin/users/:id，仅 master，缺省字段不修改） */
export interface UserUpdateParams {
  /** 登录名 */
  username?: string
  /** 邮箱（传空串表示清除） */
  email?: string
  /** 手机号（传空串表示清除） */
  phone?: string
  /** 头像地址（传空串表示清除） */
  avatar?: string
  /** 角色 */
  role?: string
  /** 状态：1 启用 / 0 禁用 */
  status?: number
  /** 是否超级管理员：1 / 0 */
  is_master?: number
}

/** 重置用户密码请求体（PUT /api/admin/users/:id/password，仅 master，无需原密码） */
export interface UserPasswordResetParams {
  /** 新密码（≥ 8 位） */
  new_password: string
}

/* -------------------------------------------------------------------------- */
/*                                    分类与标签                                */
/* -------------------------------------------------------------------------- */

/** 分类（backend/internal/model/category.go） */
export interface Category {
  id: number
  name: string
  /** SEO 友好别名（唯一） */
  slug: string
  description: string
  /** 状态：1 启用 / 0 禁用 */
  status: number
  /** 排序权重，越小越靠前 */
  sort_order: number
  /** 文章数量（聚合字段） */
  post_count: number
  created_at: string
  updated_at: string
}

/** 分类创建/更新请求体（backend/internal/dto/category.go） */
export interface CategoryUpsertParams {
  /** 分类名称（必填） */
  name: string
  /** 别名（留空自动生成） */
  slug?: string
  description?: string
  /** 状态：1 启用 / 0 禁用 */
  status?: number
  /** 排序权重 */
  sort_order?: number
}

/** 标签（backend/internal/model/tag.go） */
export interface Tag {
  id: number
  name: string
  /** SEO 友好别名（唯一） */
  slug: string
  /** 文章数量（聚合字段） */
  post_count: number
  created_at: string
  updated_at: string
}

/** 标签创建/更新请求体（backend/internal/dto/tag.go） */
export interface TagUpsertParams {
  /** 标签名称（必填） */
  name: string
  /** 别名（留空自动生成） */
  slug?: string
}

/* -------------------------------------------------------------------------- */
/*                                    文章                                     */
/* -------------------------------------------------------------------------- */

/** 文章完整对象（backend/internal/model/post.go） */
export interface Post {
  id: number
  title: string
  /** SEO 友好 URL 别名（唯一） */
  slug: string
  /** Markdown 正文（公开列表接口不返回） */
  content: string
  /** 摘要（列表页展示） */
  excerpt: string
  /** 封面图地址 */
  cover_image: string
  /** 分类 ID（可空） */
  category_id: number | null
  /** 所属分类（Preload 填充） */
  category: Category | null
  /** 标签列表（多对多） */
  tags: Tag[]
  /** 状态：draft 草稿 / published 发布 */
  status: PostStatus
  /** 是否置顶：1 置顶 / 0 否 */
  is_top: number
  /** 浏览量 */
  view_count: number
  /** 评论数（冗余字段） */
  comment_count: number
  /** 发布时间（可空） */
  published_at: string | null
  created_at: string
  updated_at: string
}

/** 公开列表中的文章（不含正文 content，见 docs/api.md 4.1） */
export type PostListItem = Omit<Post, 'content'>

/** 文章创建/更新请求体（backend/internal/dto/post.go） */
export interface PostUpsertParams {
  /** 标题（创建必填） */
  title?: string
  /** 别名（留空自动生成，重复时自动追加后缀） */
  slug?: string
  /** Markdown 正文 */
  content?: string
  /** 摘要（null 表示不修改） */
  excerpt?: string | null
  /** 封面图 URL（null 表示不修改） */
  cover_image?: string | null
  /** 分类 ID（0 表示清空分类） */
  category_id?: number | null
  /** 标签 ID 数组 */
  tag_ids?: number[]
  /** 状态：draft / published */
  status?: PostStatus
  /** 是否置顶：1 / 0 */
  is_top?: number
  /** 发布时间（RFC3339） */
  published_at?: string | null
}

/** 公开文章列表查询参数（docs/api.md 4.2） */
export interface PostListParams extends PageParams {
  /** 按分类筛选 */
  category_id?: number
  /** 按标签筛选 */
  tag_id?: number
  /** 标题/摘要关键字搜索 */
  keyword?: string
  /** 排序：latest（默认）/ views / oldest */
  sort?: PostSort
  /** 是否置顶优先（默认 true） */
  top_first?: boolean
}

/** 管理端文章列表查询参数（额外支持按状态筛选） */
export interface AdminPostListParams extends PageParams {
  /** 按状态筛选：draft / published，不传则全部 */
  status?: PostStatus
  /** 按分类筛选 */
  category_id?: number
  /** 按标签筛选 */
  tag_id?: number
  /** 置顶筛选：'1' 仅置顶 / '0' 仅非置顶，不传为全部 */
  is_top?: '0' | '1'
  /** 标题/摘要关键字搜索 */
  keyword?: string
}

/** 归档文章摘要（backend/internal/dto/post.go） */
export interface ArchivePost {
  id: number
  title: string
  slug: string
  published_at: string | null
}

/** 归档条目（按年月分组） */
export interface ArchiveItem {
  year: number
  month: number
  /** 该月文章数 */
  count: number
  posts: ArchivePost[]
}

/* -------------------------------------------------------------------------- */
/*                                    评论                                     */
/* -------------------------------------------------------------------------- */

/** 公开评论对象（不含 email/phone/ip/user_agent 等隐私字段，见 docs/api.md 6.1） */
export interface Comment {
  id: number
  post_id: number
  /** 父评论 ID，0 表示顶层评论（支持楼中楼） */
  parent_id: number
  /** 访客昵称 */
  nickname: string
  content: string
  status: CommentStatus
  /** 是否博主评论：1 是 / 0 否 */
  is_author: number
  /** 回复数（冗余字段） */
  reply_count: number
  likes: number
  created_at: string
  /** 子评论列表（楼中楼） */
  replies?: Comment[]
}

/** 管理端评论对象（含隐私字段） */
export interface AdminComment extends Comment {
  /** 邮箱（公开接口不回传） */
  email: string
  /** 手机号（公开接口不回传） */
  phone: string
  /** 评论者 IP（公开接口不回传） */
  ip: string
  /** 用户代理（公开接口不回传） */
  user_agent: string
  /** 审核时间（可空） */
  reviewed_at: string | null
  /** 关联文章标题（管理端聚合字段） */
  post_title?: string
  /** 父评论昵称（管理端聚合字段） */
  parent_nickname?: string
  updated_at: string
}

/** 获取文章评论查询参数（仅分页，docs/api.md 6.2） */
export type CommentListParams = PageParams

/** 提交评论请求体（docs/api.md 6.3） */
export interface CommentCreateParams {
  /** 文章 ID（与 slug 二选一） */
  post_id?: number
  /** 文章别名（与 post_id 二选一） */
  slug?: string
  /** 父评论 ID，0/省略为顶层评论 */
  parent_id?: number
  /** 昵称（必填，XSS 过滤） */
  nickname: string
  /** 邮箱（必填，用于接收验证码与通知） */
  email?: string
  /** 手机号（可选） */
  phone?: string
  /** 评论内容（必填，XSS 过滤） */
  content: string
  /** 验证码 token（发送验证码时返回，站点开启验证码时必填） */
  captcha_token?: string
  /** 验证码内容（站点开启验证码时必填） */
  captcha_code?: string
}

/** 提交评论响应（docs/api.md 6.3） */
export interface CommentCreateResult {
  /** 新评论对象 */
  comment: Comment
  /** 评论状态：pending 待审核 / approved 已通过 */
  status: CommentStatus
}

/** 管理端评论列表查询参数（backend/internal/dto/comment.go） */
export interface AdminCommentListParams extends PageParams {
  /** 审核状态筛选：pending / approved / rejected */
  status?: CommentStatus
  /** 按文章筛选 */
  post_id?: number
  /** 昵称/内容关键字 */
  keyword?: string
}

/** 管理端修改评论请求体（仅允许修改 status / is_author，其余字段后端一律忽略） */
export interface CommentUpdateParams {
  /** 审核状态：pending / approved / rejected（缺省不修改） */
  status?: CommentStatus
  /** 是否博主评论：1 是 / 0 否（缺省不修改） */
  is_author?: number
}

/* -------------------------------------------------------------------------- */
/*                                  站点配置                                   */
/* -------------------------------------------------------------------------- */

/** 社交链接项（social_links 为数组，数组顺序即前台显示顺序） */
export interface SocialLink {
  /** 名称（前台直接展示） */
  name: string
  /** 链接地址 */
  url: string
}

/** 前台生效配置（backend/internal/dto/config.go PublicConfig） */
export interface PublicConfig {
  title: string
  subtitle: string
  /** 侧栏卡片名称（留空回退站点标题） */
  side_name: string
  description: string
  /** 博客头像 URL */
  avatar: string
  /** ICP 备案号 */
  icp: string
  /** 警察 ICP 备案号 */
  police_icp: string
  theme: string
  /** About 页面内容 */
  about_content: string
  /** 首页随机背景图 URL 列表 */
  home_images: string[]
  /** 社交链接（JSON 数组，顺序即显示顺序） */
  social_links: SocialLink[]
  /** 评论是否需要审核：1 是 / 0 否 */
  comment_need_review: number
  /** 评论是否需要邮箱验证码：1 是 / 0 否 */
  comment_need_captcha: number
  /** 自定义 CSS 变量（每行一个 --name: value;，仅平坦变量） */
  custom_css_vars: string
  /** 黑暗主题专属的自定义 CSS 变量（仅 data-theme=dark 时生效） */
  custom_css_vars_dark: string
  /** 是否启用自定义 CSS 变量：1 是 / 0 否 */
  custom_css_enabled: number
  /** 扩展字段（JSON 对象） */
  extra: Record<string, unknown>
}

/** 站点配置管理对象（backend/internal/model/site_config.go，social_links/extra 为 JSON 字符串） */
export interface SiteConfig {
  id: number
  /** 版本号（如 1.0.0） */
  version: string
  /** 是否生效：1 生效 / 0 否 */
  is_active: number
  title: string
  subtitle: string
  /** 侧栏卡片名称（留空回退站点标题） */
  side_name: string
  description: string
  /** 博客头像 URL */
  avatar: string
  icp: string
  police_icp: string
  theme: string
  about_content: string
  /** 首页随机背景图（JSON 字符串） */
  home_images: string
  /** 社交链接（JSON 字符串） */
  social_links: string
  comment_need_review: number
  /** 评论是否需要邮箱验证码：1 是 / 0 否 */
  comment_need_captcha: number
  /** 自定义 CSS 变量（平坦变量文本） */
  custom_css_vars: string
  /** 黑暗主题专属的自定义 CSS 变量（平坦变量文本） */
  custom_css_vars_dark: string
  /** 是否启用自定义 CSS 变量：1 是 / 0 否 */
  custom_css_enabled: number
  /** 扩展字段（JSON 字符串） */
  extra: string
  created_by: string
  created_at: string
  updated_at: string
}

/** 站点配置创建/更新请求体（backend/internal/dto/config.go，创建时 version 必填） */
export interface ConfigUpsertParams {
  /** 版本号（创建必填） */
  version?: string
  title?: string
  subtitle?: string
  /** 侧栏卡片名称（空串表示清除，回退站点标题） */
  side_name?: string
  description?: string
  /** 博客头像 URL */
  avatar?: string
  icp?: string
  police_icp?: string
  theme?: string
  about_content?: string
  /** 首页随机背景图 URL 列表 */
  home_images?: string[]
  /** 社交链接（JSON 数组，顺序即显示顺序） */
  social_links?: SocialLink[]
  /** 评论是否需要审核：1 是 / 0 否 */
  comment_need_review?: number
  /** 评论是否需要邮箱验证码：1 是 / 0 否 */
  comment_need_captcha?: number
  /** 自定义 CSS 变量（平坦变量文本，空字符串表示清除） */
  custom_css_vars?: string
  /** 黑暗主题专属的自定义 CSS 变量（平坦变量文本，空字符串表示清除） */
  custom_css_vars_dark?: string
  /** 是否启用自定义 CSS 变量：1 是 / 0 否 */
  custom_css_enabled?: number
  /** 扩展字段（JSON 对象） */
  extra?: Record<string, unknown>
  created_by?: string
}

/* -------------------------------------------------------------------------- */
/*                                    友链                                     */
/* -------------------------------------------------------------------------- */

/** 友链（backend/internal/model/friend.go） */
export interface Friend {
  id: number
  name: string
  /** 友链地址 */
  url: string
  /** 友链图标（图片 URL 或 Font Awesome 类名） */
  icon: string
  description: string
  /** 状态：1 启用 / 0 禁用 */
  status: number
  /** 排序权重，越小越靠前 */
  sort_order: number
  created_at: string
  updated_at: string
}

/** 友链创建/更新请求体（backend/internal/dto/friend.go） */
export interface FriendUpsertParams {
  /** 友链名称（必填） */
  name: string
  /** 友链地址（必填） */
  url: string
  /** 友链图标（图片 URL 或 Font Awesome 类名） */
  icon?: string
  description?: string
  /** 状态：1 启用 / 0 禁用 */
  status?: number
  /** 排序权重 */
  sort_order?: number
}

/* -------------------------------------------------------------------------- */
/*                                    作品集                                   */
/* -------------------------------------------------------------------------- */

/** 作品（backend/internal/model/work.go） */
export interface Work {
  id: number
  /** 作品名 */
  name: string
  /** SEO 友好别名（唯一） */
  slug: string
  /** 作品描述 */
  description: string
  /** 封面图 URL */
  cover: string
  /** 演示链接 */
  demo_url: string
  /** 文章链接 */
  article_url: string
  /** 源码链接 */
  repo_url: string
  /** 技术栈（JSON 字符串数组，如 ["Vue 3","Go"]） */
  tech_stack: string
  /** 年份/时间（如 2025、2024-06） */
  year: string
  /** 是否置顶/推荐：1 是 / 0 否 */
  is_top: number
  /** 状态：1 启用 / 0 禁用 */
  status: number
  /** 排序权重，越小越靠前 */
  sort_order: number
  /** 扩展字段（JSON 字符串） */
  extra: string
  created_at: string
  updated_at: string
}

/** 作品创建/更新请求体（backend/internal/dto/work.go） */
export interface WorkUpsertParams {
  /** 作品名（必填） */
  name: string
  /** 别名（留空自动生成，重复时自动追加后缀） */
  slug?: string
  /** 描述（null 表示不修改） */
  description?: string | null
  /** 封面图 URL（null 表示不修改） */
  cover?: string | null
  /** 演示链接（null 表示不修改） */
  demo_url?: string | null
  /** 文章链接（null 表示不修改） */
  article_url?: string | null
  /** 源码链接（null 表示不修改） */
  repo_url?: string | null
  /** 技术栈数组（空数组表示清空） */
  tech_stack?: string[]
  /** 年份/时间（null 表示不修改） */
  year?: string | null
  /** 是否置顶：1 / 0 */
  is_top?: number
  /** 状态：1 启用 / 0 禁用 */
  status?: number
  /** 排序权重 */
  sort_order?: number
}

/* -------------------------------------------------------------------------- */
/*                                    文件                                     */
/* -------------------------------------------------------------------------- */

/** 文件记录（backend/internal/model/file.go） */
export interface FileItem {
  id: number
  /** 上传者 ID（单用户系统可为空） */
  user_id: number | null
  /** 原始文件名 */
  filename: string
  /** 存储路径（相对路径） */
  path: string
  /** 缩略图路径（仅图片文件） */
  thumbnail: string
  /** MIME 类型 */
  mime_type: string
  /** 文件大小（字节） */
  size: number
  /** 完整访问 URL（后端按存储类型拼装） */
  url?: string
  /** 缩略图完整访问 URL */
  thumbnail_url?: string
  created_at: string
  updated_at: string
}

/** 文件上传结果（docs/api.md 10.1） */
export interface UploadResult {
  id: number
  filename: string
  /** 可访问 URL */
  url: string
  size: number
  mime_type: string
}

/** 文件列表查询参数（backend/internal/dto/file.go） */
export interface FileListParams extends PageParams {
  /** 按 MIME 类型筛选（前缀匹配，如 image/） */
  mime_type?: string
  /** 按原始文件名模糊搜索 */
  keyword?: string
}

/* -------------------------------------------------------------------------- */
/*                                    音乐                                     */
/* -------------------------------------------------------------------------- */

/** 音乐（backend/internal/model/music.go，音频文件本体在 files 表） */
export interface Music {
  id: number
  /** 歌名（APlayer name） */
  name: string
  /** 歌手（APlayer artist） */
  artist: string
  /** 封面图 URL（可空） */
  cover: string
  /** 关联文件 ID（files 表） */
  file_id: number
  /** 排序权重，越小越靠前 */
  sort_order: number
  /** 状态：1 启用 / 0 禁用 */
  status: number
  /** 歌词文本（预留） */
  lrc: string
  /** 音频完整访问 URL（后端按存储类型拼装，浏览器直连） */
  url?: string
  created_at: string
  updated_at: string
}

/** 公开音乐（backend/internal/dto/music.go PublicMusic，不含 file_id 等内部字段） */
export interface PublicMusic {
  id: number
  name: string
  artist: string
  cover: string
  /** 音频完整访问 URL */
  url: string
  /** 歌词文本 */
  lrc: string
}

/** 音乐创建/更新请求体（backend/internal/dto/music.go） */
export interface MusicUpsertParams {
  /** 歌名（必填） */
  name?: string
  /** 歌手（null 表示不修改，空串表示清除） */
  artist?: string | null
  /** 封面图 URL（null 表示不修改，空串表示清除） */
  cover?: string | null
  /** 关联文件 ID（创建必填；更新时提供且变更会同步清理旧音频文件） */
  file_id?: number
  /** 排序权重 */
  sort_order?: number
  /** 状态：1 启用 / 0 禁用 */
  status?: number
  /** 歌词文本（null 表示不修改，空串表示清除） */
  lrc?: string | null
}

/** 管理端音乐列表查询参数 */
export interface MusicQueryParams extends PageParams {
  /** 按歌名/歌手模糊搜索 */
  keyword?: string
}

/** APlayer 选项（存于站点配置 extra.aplayer；播放器形态固定 fixed，不允许配置修改） */
export interface APlayerOptions {
  /** 是否启用播放器 */
  enabled?: boolean
  /** 自动播放（浏览器可能拦截） */
  autoplay?: boolean
  /** 主题色 */
  theme?: string
  /** 播放顺序：list / random */
  order?: 'list' | 'random'
  /** 循环：all / one / none */
  loop?: 'all' | 'one' | 'none'
  /** 音量（0-1） */
  volume?: number
  /** 默认收起播放列表 */
  listFolded?: boolean
  /** 播放列表最大高度（px） */
  listMaxHeight?: number
}

/* -------------------------------------------------------------------------- */
/*                                    统计                                     */
/* -------------------------------------------------------------------------- */

/** 仪表盘统计（docs/api.md 11.1） */
export interface StatsSummary {
  /** 文章总数（含草稿） */
  post_count: number
  /** 已发布文章数 */
  published_count: number
  /** 草稿数 */
  draft_count: number
  /** 分类数 */
  category_count: number
  /** 标签数 */
  tag_count: number
  /** 评论总数 */
  comment_count: number
  /** 待审核评论数 */
  pending_comment_count: number
  /** 总浏览量 */
  total_views: number
  /** 最近发布文章（约 5 篇） */
  recent_posts: PostListItem[]
}

/* -------------------------------------------------------------------------- */
/*                                   认证接口                                   */
/* -------------------------------------------------------------------------- */

/** 管理员登录请求体（docs/api.md 3.1） */
export interface LoginParams {
  /** 管理员用户名 */
  username: string
  /** 密码（bcrypt 校验） */
  password: string
  /** 图形验证码 ID（可选） */
  captcha_id?: string
  /** 图形验证码内容（可选） */
  captcha_code?: string
}

/** 登录成功响应（docs/api.md 3.1） */
export interface LoginResult {
  /** 访问令牌（JWT） */
  token: string
  /** 刷新令牌 */
  refresh_token: string
  /** 访问令牌有效期（秒） */
  expires_in: number
  /** 用户信息 */
  user: User
}

/** 刷新访问 Token 响应（docs/api.md 3.2） */
export interface RefreshResult {
  token: string
  refresh_token: string
  expires_in: number
}

/** 刷新访问 Token 请求体 */
export interface RefreshParams {
  /** 登录时返回的刷新令牌 */
  refresh_token: string
}

/** 修改密码请求体（docs/api.md 3.4） */
export interface UpdatePasswordParams {
  /** 原密码 */
  old_password: string
  /** 新密码（建议长度 ≥ 8，含字母数字） */
  new_password: string
}

/* -------------------------------------------------------------------------- */
/*                                   验证码                                    */
/* -------------------------------------------------------------------------- */

/** 发送邮箱验证码请求体（docs/api.md 12.1） */
export interface EmailCaptchaParams {
  /** 目标邮箱 */
  email: string
  /** 场景：comment / login */
  scene?: CaptchaScene
}

/** 发送邮箱验证码响应 */
export interface EmailCaptchaResult {
  /** 校验用 token，提交评论时携带 */
  token: string
  /** 验证码过期时间（秒） */
  expires_in: number
}

/** 校验验证码请求体（docs/api.md 12.2） */
export interface VerifyCaptchaParams {
  /** 发送验证码时返回的 token */
  token: string
  /** 验证码内容 */
  code: string
}

/** 校验验证码响应 */
export interface VerifyCaptchaResult {
  /** 是否校验通过（通过后验证码即作废） */
  valid: boolean
}

/* -------------------------------------------------------------------------- */
/*                                 通用辅助类型                                 */
/* -------------------------------------------------------------------------- */

/** 请求配置类型别名：供各 API 模块内部使用 */
export type RequestConfig = AxiosRequestConfig

/** 浏览量 +1 响应（docs/api.md 4.4） */
export interface ViewResult {
  /** 最新浏览量 */
  view_count: number
}
