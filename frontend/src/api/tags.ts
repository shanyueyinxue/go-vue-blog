/**
 * 标签相关接口（docs/api.md 第 5 章）
 *
 * 公开接口：
 * - `GET /api/tags`  标签列表（仅启用，附 post_count）
 *
 * 管理接口（需 JWT）：
 * - `GET    /api/admin/tags`      列表（分页）
 * - `POST   /api/admin/tags`      创建
 * - `PUT    /api/admin/tags/:id`  更新
 * - `DELETE /api/admin/tags/:id`  删除（并清理 post_tags 关联）
 */

import { request } from './http'
import type { PageParams, PageResult, Tag, TagUpsertParams } from './types'

/** 公开标签列表（docs/api.md 5.2） */
export function listTags() {
  return request<Tag[]>({
    url: '/tags',
    method: 'get',
  })
}

/** 管理端标签列表（docs/api.md 5.2，分页） */
export function listAdminTags(params: PageParams) {
  return request<PageResult<Tag>>({
    url: '/admin/tags',
    method: 'get',
    params,
  })
}

/** 创建标签 */
export function createTag(params: TagUpsertParams) {
  return request<Tag>({
    url: '/admin/tags',
    method: 'post',
    data: params,
  })
}

/** 更新标签 */
export function updateTag(id: number, params: TagUpsertParams) {
  return request<Tag>({
    url: `/admin/tags/${id}`,
    method: 'put',
    data: params,
  })
}

/** 删除标签（物理删除，并清理 post_tags 关联） */
export function deleteTag(id: number) {
  return request<{ message: string }>({
    url: `/admin/tags/${id}`,
    method: 'delete',
  })
}
