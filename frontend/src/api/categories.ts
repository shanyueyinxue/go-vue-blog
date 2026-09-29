/**
 * 分类相关接口（docs/api.md 第 5 章）
 *
 * 公开接口：
 * - `GET /api/categories`  分类列表（仅启用，附 post_count）
 *
 * 管理接口（需 JWT）：
 * - `GET    /api/admin/categories`      列表（含禁用，分页）
 * - `POST   /api/admin/categories`      创建
 * - `PUT    /api/admin/categories/:id`  更新
 * - `DELETE /api/admin/categories/:id`  删除（分类下存在文章时拒绝）
 */

import { request } from './http'
import type { Category, CategoryUpsertParams, PageParams, PageResult } from './types'

/** 公开分类列表（docs/api.md 5.1，仅返回启用分类） */
export function listCategories() {
  return request<Category[]>({
    url: '/categories',
    method: 'get',
  })
}

/** 管理端分类列表（docs/api.md 5.1，分页，含禁用分类） */
export function listAdminCategories(params: PageParams) {
  return request<PageResult<Category>>({
    url: '/admin/categories',
    method: 'get',
    params,
  })
}

/** 创建分类 */
export function createCategory(params: CategoryUpsertParams) {
  return request<Category>({
    url: '/admin/categories',
    method: 'post',
    data: params,
  })
}

/** 更新分类 */
export function updateCategory(id: number, params: CategoryUpsertParams) {
  return request<Category>({
    url: `/admin/categories/${id}`,
    method: 'put',
    data: params,
  })
}

/** 删除分类（软删除；分类下存在文章时返回 40900） */
export function deleteCategory(id: number) {
  return request<{ message: string }>({
    url: `/admin/categories/${id}`,
    method: 'delete',
  })
}
