/**
 * API 模块统一出口
 *
 * 使用方式：
 * ```ts
 * import { listPosts, login } from '@/api'
 * ```
 */

export * from './types'
export {
  request,
  BizError,
  isUnauthorized,
  getAccessToken,
  getRefreshToken,
  setTokens,
  clearTokens,
  AUTH_EXPIRED_EVENT,
} from './http'

export * from './auth'
export * from './users'
export * from './posts'
export * from './categories'
export * from './tags'
export * from './comments'
export * from './configs'
export * from './friends'
export * from './works'
export * from './files'
export * from './musics'
export * from './stats'
export * from './captcha'
