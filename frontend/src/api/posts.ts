/**
 * 文章相关接口（docs/api.md 第 4 章）
 *
 * 公开接口（无需认证）：
 * - `GET  /api/posts`              文章列表（分页/分类/标签/搜索）
 * - `GET  /api/posts/archive`      文章归档（按年月）
 * - `GET  /api/posts/:slug`        文章详情
 * - `POST /api/posts/:slug/view`   浏览量 +1
 *
 * 管理接口（需 JWT）：
 * - `GET    /api/admin/posts`          列表（含草稿）
 * - `POST   /api/admin/posts`          创建
 * - `GET    /api/admin/posts/:id`      详情
 * - `PUT    /api/admin/posts/:id`      更新
 * - `DELETE /api/admin/posts/:id`      删除（软删除）
 */

import { request } from './http'
import type {
  AdminPostListParams,
  ArchiveItem,
  PageResult,
  Post,
  PostListItem,
  PostListParams,
  PostUpsertParams,
  ViewResult,
} from './types'

/** 公开文章列表（docs/api.md 4.2，仅返回已发布文章） */
export function listPosts(params: PostListParams) {
  return request<PageResult<PostListItem>>({
    url: '/posts',
    method: 'get',
    params,
  })
}

/** 文章归档（按年月分组，docs/api.md 4.5） */
export function getArchive() {
  return request<ArchiveItem[]>({
    url: '/posts/archive',
    method: 'get',
  })
}

/** 公开文章详情（docs/api.md 4.3，仅返回已发布文章，含完整正文） */
export function getPostBySlug(slug: string) {
  return request<Post>({
    url: `/posts/${encodeURIComponent(slug)}`,
    method: 'get',
  })
}

/** 文章浏览量 +1（docs/api.md 4.4） */
export function increasePostView(slug: string) {
  return request<ViewResult>({
    url: `/posts/${encodeURIComponent(slug)}/view`,
    method: 'post',
  })
}

/* ------------------------------- 管理接口 ------------------------------- */

/** 管理端文章列表（docs/api.md 4.6，含草稿） */
export function listAdminPosts(params: AdminPostListParams) {
  return request<PageResult<PostListItem>>({
    url: '/admin/posts',
    method: 'get',
    params,
  })
}

/** 创建文章（docs/api.md 4.6） */
export function createPost(params: PostUpsertParams) {
  return request<Post>({
    url: '/admin/posts',
    method: 'post',
    data: params,
  })
}

/** 管理端文章详情（docs/api.md 4.6，含草稿的完整数据） */
export function getAdminPost(id: number) {
  return request<Post>({
    url: `/admin/posts/${id}`,
    method: 'get',
  })
}

/** 更新文章（docs/api.md 4.6，字段可选，null 表示不修改） */
export function updatePost(id: number, params: PostUpsertParams) {
  return request<Post>({
    url: `/admin/posts/${id}`,
    method: 'put',
    data: params,
  })
}

/** 删除文章（docs/api.md 4.6，软删除） */
export function deletePost(id: number) {
  return request<{ message: string }>({
    url: `/admin/posts/${id}`,
    method: 'delete',
  })
}
