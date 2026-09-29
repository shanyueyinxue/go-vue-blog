/**
 * 用户管理接口（仅超级管理员 is_master=1 可访问，docs/api.md 第 14 章）
 *
 * - `GET    /api/admin/users`           用户列表（分页 + 关键字搜索）
 * - `POST   /api/admin/users`           创建用户
 * - `GET    /api/admin/users/:id`       用户详情
 * - `PUT    /api/admin/users/:id`       修改用户信息
 * - `PUT    /api/admin/users/:id/password` 重置用户密码（无需原密码）
 * - `DELETE /api/admin/users/:id`       删除用户
 */

import { request } from './http'
import type {
  AdminUserQueryParams,
  PageResult,
  User,
  UserCreateParams,
  UserPasswordResetParams,
  UserUpdateParams,
} from './types'

/** 用户列表（docs/api.md 14.1） */
export function listAdminUsers(params: AdminUserQueryParams) {
  return request<PageResult<User>>({
    url: '/admin/users',
    method: 'get',
    params,
  })
}

/** 创建用户（docs/api.md 14.2） */
export function createUser(params: UserCreateParams) {
  return request<User>({
    url: '/admin/users',
    method: 'post',
    data: params,
  })
}

/** 用户详情（docs/api.md 14.3） */
export function getUser(id: number) {
  return request<User>({
    url: `/admin/users/${id}`,
    method: 'get',
  })
}

/** 修改用户信息（docs/api.md 14.4） */
export function updateUser(id: number, params: UserUpdateParams) {
  return request<User>({
    url: `/admin/users/${id}`,
    method: 'put',
    data: params,
  })
}

/** 重置用户密码（docs/api.md 14.5，无需原密码，改后旧 Token 失效） */
export function resetUserPassword(id: number, params: UserPasswordResetParams) {
  return request<{ message: string }>({
    url: `/admin/users/${id}/password`,
    method: 'put',
    data: params,
  })
}

/** 删除用户（docs/api.md 14.6，不能删除当前登录账号） */
export function deleteUser(id: number) {
  return request<{ message: string }>({
    url: `/admin/users/${id}`,
    method: 'delete',
  })
}
