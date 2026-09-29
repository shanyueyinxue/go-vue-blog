/**
 * 作品集相关接口（docs/api.md 第 8 章）
 *
 * 公开接口：
 * - `GET /api/works`  作品列表（仅启用，置顶优先）
 *
 * 管理接口（需 JWT）：
 * - `GET    /api/admin/works`      列表（分页，含禁用）
 * - `POST   /api/admin/works`      创建
 * - `GET    /api/admin/works/:id`  详情
 * - `PUT    /api/admin/works/:id`  更新
 * - `DELETE /api/admin/works/:id`  删除（软删除）
 */

import { request } from './http'
import type { PageParams, PageResult, Work, WorkUpsertParams } from './types'

/** 公开作品列表（docs/api.md 8.1，仅返回启用作品） */
export function listWorks() {
  return request<Work[]>({
    url: '/works',
    method: 'get',
  })
}

/** 管理端作品列表（docs/api.md 8.2，分页，含禁用） */
export function listAdminWorks(params: PageParams) {
  return request<PageResult<Work>>({
    url: '/admin/works',
    method: 'get',
    params,
  })
}

/** 管理端作品详情（docs/api.md 8.2） */
export function getAdminWork(id: number) {
  return request<Work>({
    url: `/admin/works/${id}`,
    method: 'get',
  })
}

/** 创建作品（docs/api.md 8.2，name 必填，slug 留空自动生成） */
export function createWork(params: WorkUpsertParams) {
  return request<Work>({
    url: '/admin/works',
    method: 'post',
    data: params,
  })
}

/** 更新作品（docs/api.md 8.2，字段缺省表示不修改） */
export function updateWork(id: number, params: WorkUpsertParams) {
  return request<Work>({
    url: `/admin/works/${id}`,
    method: 'put',
    data: params,
  })
}

/** 删除作品（docs/api.md 8.2，软删除） */
export function deleteWork(id: number) {
  return request<{ message: string }>({
    url: `/admin/works/${id}`,
    method: 'delete',
  })
}
