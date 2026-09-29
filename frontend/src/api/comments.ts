/**
 * 评论相关接口（docs/api.md 第 6 章）
 *
 * 公开接口：
 * - `GET  /api/posts/:slug/comments`  获取文章已通过评论（含楼中楼）
 * - `POST /api/comments`              提交评论（匿名）
 *
 * 管理接口（需 JWT）：
 * - `GET    /api/admin/comments`              列表（含待审核）
 * - `PUT    /api/admin/comments/:id`          修改（仅 status / is_author）
 * - `DELETE /api/admin/comments/:id`          删除（软删除）
 */

import { request } from './http'
import type {
  AdminComment,
  AdminCommentListParams,
  Comment,
  CommentCreateParams,
  CommentCreateResult,
  CommentListParams,
  CommentUpdateParams,
  PageResult,
} from './types'

/** 获取某文章已通过的评论（docs/api.md 6.2，仅返回 approved） */
export function listComments(slug: string, params: CommentListParams) {
  return request<PageResult<Comment>>({
    url: `/posts/${encodeURIComponent(slug)}/comments`,
    method: 'get',
    params,
  })
}

/** 提交评论（docs/api.md 6.3，匿名提交） */
export function createComment(params: CommentCreateParams) {
  return request<CommentCreateResult>({
    url: '/comments',
    method: 'post',
    data: params,
  })
}

/** 管理端评论列表（docs/api.md 6.4，含待审核，返回完整隐私字段） */
export function listAdminComments(params: AdminCommentListParams) {
  return request<PageResult<AdminComment>>({
    url: '/admin/comments',
    method: 'get',
    params,
  })
}

/** 修改评论（docs/api.md 6.4，仅允许修改 status / is_author） */
export function updateComment(id: number, params: CommentUpdateParams) {
  return request<AdminComment>({
    url: `/admin/comments/${id}`,
    method: 'put',
    data: params,
  })
}

/** 删除评论（docs/api.md 6.4，软删除，级联处理子评论） */
export function deleteComment(id: number) {
  return request<{ message: string }>({
    url: `/admin/comments/${id}`,
    method: 'delete',
  })
}
