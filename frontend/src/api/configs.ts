/**
 * 站点配置相关接口（docs/api.md 第 7 章）
 *
 * 公开接口：
 * - `GET /api/configs/public`  前台生效配置（内存缓存）
 *
 * 管理接口（需 JWT）：
 * - `GET    /api/admin/configs`              配置版本列表（分页）
 * - `GET    /api/admin/configs/:id`          单个配置
 * - `POST   /api/admin/configs`              新增配置（自动激活）
 * - `PUT    /api/admin/configs/:id`          更新配置
 * - `PUT    /api/admin/configs/:id/activate` 激活指定版本
 * - `DELETE /api/admin/configs/:id`          删除配置（激活中禁止删除）
 */

import { request } from './http'
import type { ConfigUpsertParams, PageParams, PageResult, PublicConfig, SiteConfig } from './types'

/** 前台生效配置（docs/api.md 7.1，无需认证） */
export function getPublicConfig() {
  return request<PublicConfig>({
    url: '/configs/public',
    method: 'get',
  })
}

/** 管理端配置列表（docs/api.md 7.2，分页） */
export function listAdminConfigs(params: PageParams) {
  return request<PageResult<SiteConfig>>({
    url: '/admin/configs',
    method: 'get',
    params,
  })
}

/** 管理端单个配置详情 */
export function getAdminConfig(id: number) {
  return request<SiteConfig>({
    url: `/admin/configs/${id}`,
    method: 'get',
  })
}

/** 新增配置（docs/api.md 7.2，创建后自动置为激活） */
export function createConfig(params: ConfigUpsertParams) {
  return request<SiteConfig>({
    url: '/admin/configs',
    method: 'post',
    data: params,
  })
}

/** 更新配置（docs/api.md 7.2，不改变 is_active 状态） */
export function updateConfig(id: number, params: ConfigUpsertParams) {
  return request<SiteConfig>({
    url: `/admin/configs/${id}`,
    method: 'put',
    data: params,
  })
}

/** 激活指定版本配置（docs/api.md 7.2，同时取消其它版本激活） */
export function activateConfig(id: number) {
  return request<{ message: string }>({
    url: `/admin/configs/${id}/activate`,
    method: 'put',
  })
}

/** 删除配置（docs/api.md 7.2，激活中的配置禁止删除） */
export function deleteConfig(id: number) {
  return request<{ message: string }>({
    url: `/admin/configs/${id}`,
    method: 'delete',
  })
}
