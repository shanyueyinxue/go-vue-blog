/**
 * 友链相关接口（docs/api.md 第 9 章）
 *
 * 公开接口：
 * - `GET /api/friends`  友链列表（仅启用）
 *
 * 管理接口（需 JWT）：
 * - `GET    /api/admin/friends`      列表（分页，返回全部）
 * - `POST   /api/admin/friends`      创建
 * - `PUT    /api/admin/friends/:id`  更新
 * - `DELETE /api/admin/friends/:id`  删除（软删除）
 */

import { request } from './http'
import type { Friend, FriendUpsertParams, PageParams, PageResult } from './types'

/** 公开友链列表（docs/api.md 9.1，仅返回启用友链） */
export function listFriends() {
  return request<Friend[]>({
    url: '/friends',
    method: 'get',
  })
}

/** 管理端友链列表（docs/api.md 9.2，分页） */
export function listAdminFriends(params: PageParams) {
  return request<PageResult<Friend>>({
    url: '/admin/friends',
    method: 'get',
    params,
  })
}

/** 创建友链（docs/api.md 9.2） */
export function createFriend(params: FriendUpsertParams) {
  return request<Friend>({
    url: '/admin/friends',
    method: 'post',
    data: params,
  })
}

/** 更新友链 */
export function updateFriend(id: number, params: FriendUpsertParams) {
  return request<Friend>({
    url: `/admin/friends/${id}`,
    method: 'put',
    data: params,
  })
}

/** 删除友链（软删除） */
export function deleteFriend(id: number) {
  return request<{ message: string }>({
    url: `/admin/friends/${id}`,
    method: 'delete',
  })
}
