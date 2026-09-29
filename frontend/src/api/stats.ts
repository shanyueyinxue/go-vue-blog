/**
 * 统计相关接口（docs/api.md 第 11 章）
 *
 * - `GET /api/admin/stats/summary`  仪表盘统计（需 JWT）
 */

import { request } from './http'
import type { StatsSummary } from './types'

/** 仪表盘统计（docs/api.md 11.1，需 JWT） */
export function getStatsSummary() {
  return request<StatsSummary>({
    url: '/admin/stats/summary',
    method: 'get',
  })
}
