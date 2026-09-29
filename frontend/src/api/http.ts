/**
 * axios 实例与拦截器封装
 *
 * 职责：
 * 1. 创建统一 baseURL（`VITE_API_BASE_URL`，默认 `/api`）的 axios 实例；
 * 2. 请求拦截器自动注入 `Authorization: Bearer <token>`；
 * 3. 响应拦截器统一处理后端 `{ code, message, data }` 响应结构，
 *    业务错误（code !== 0）抛出 {@link BizError}；
 * 4. 访问令牌过期（40100）时自动用刷新令牌换新并重放原请求，
 *    刷新失败则清理本地凭证并派发 `auth:expired` 事件。
 */

import axios, {
  AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from 'axios'
import { ApiCode, type ApiResponse, type RefreshResult, type RequestConfig } from './types'

/* -------------------------------------------------------------------------- */
/*                               Token 本地存储管理                             */
/* -------------------------------------------------------------------------- */

/** 访问令牌 localStorage 键名 */
const ACCESS_TOKEN_KEY = 'blog_access_token'
/** 刷新令牌 localStorage 键名 */
const REFRESH_TOKEN_KEY = 'blog_refresh_token'

/** 读取访问令牌 */
export function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

/** 读取刷新令牌 */
export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

/** 写入访问令牌与刷新令牌 */
export function setTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

/** 清理全部本地凭证（登出/登录失效时调用） */
export function clearTokens(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

/* -------------------------------------------------------------------------- */
/*                                  业务错误类型                                */
/* -------------------------------------------------------------------------- */

/** 后端返回业务错误（code !== 0）时抛出的错误 */
export class BizError extends Error {
  /** 后端业务错误码（见 ApiCode） */
  code: number
  /** 原始请求配置（可用于重试） */
  config?: RequestConfig

  constructor(code: number, message: string, config?: RequestConfig) {
    super(message)
    this.name = 'BizError'
    this.code = code
    this.config = config
  }
}

/**
 * 判断错误是否为登录失效（未认证）类型：
 * 后端业务码 40100 或 HTTP 401。
 */
export function isUnauthorized(error: unknown): boolean {
  if (error instanceof BizError) {
    return error.code === ApiCode.Unauthorized
  }
  return error instanceof AxiosError && error.response?.status === 401
}

/* -------------------------------------------------------------------------- */
/*                                axios 实例与拦截器                            */
/* -------------------------------------------------------------------------- */

/** API 基础路径：优先使用环境变量，默认 `/api`（配合 Vite 代理使用） */
const BASE_URL: string = import.meta.env.VITE_API_BASE_URL ?? '/api'

/** 统一的 axios 实例 */
const service: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 15000,
  // 后端统一返回 JSON，声明以便 axios 正确解析
  headers: { 'Content-Type': 'application/json' },
})

/** 请求拦截器：自动注入访问令牌 */
service.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

/**
 * 登录失效事件名。
 * 当刷新令牌也失效时触发，业务层（如路由守卫/全局提示）可监听并跳转登录页：
 * `window.addEventListener('auth:expired', handler)`
 */
export const AUTH_EXPIRED_EVENT = 'auth:expired'

/** 是否正在刷新令牌（防止并发重复刷新） */
let refreshing = false
/** 刷新期间挂起的请求队列 */
let pendingQueue: Array<(token: string) => void> = []

/** 冲刷等待队列：向所有等待请求分发新令牌 */
function flushQueue(token: string): void {
  pendingQueue.forEach((resolve) => resolve(token))
  pendingQueue = []
}

/**
 * 使用刷新令牌换取新的访问令牌。
 * 注意：使用裸 axios 请求，避免再次进入本实例拦截器造成循环调用。
 */
async function refreshAccessToken(): Promise<string> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    throw new BizError(ApiCode.Unauthorized, '未找到刷新令牌，请重新登录')
  }
  const { data } = await axios.post<ApiResponse<RefreshResult>>(
    `${BASE_URL}/admin/refresh`,
    { refresh_token: refreshToken },
  )
  if (data.code !== ApiCode.Success) {
    throw new BizError(data.code, data.message)
  }
  // 刷新成功：更新本地凭证
  setTokens(data.data.token, data.data.refresh_token)
  return data.data.token
}

/** 响应拦截器：统一处理业务码、错误码与自动刷新 */
service.interceptors.response.use(
  (response) => {
    // 后端所有接口均返回 { code, message, data }
    const res = response.data as ApiResponse
    if (res.code !== ApiCode.Success) {
      // 业务层面返回的错误码（如 40400、40900 等），统一抛出
      return Promise.reject(new BizError(res.code, res.message, response.config))
    }
    return response
  },
  async (error: AxiosError<ApiResponse>) => {
    const config = error.config as InternalAxiosRequestConfig | undefined
    const data = error.response?.data
    const bizCode = data?.code

    // 未认证（HTTP 401 或业务码 40100）且存在刷新令牌时，尝试刷新并重放请求
    if (
      (error.response?.status === 401 || bizCode === ApiCode.Unauthorized) &&
      config &&
      !config.url?.includes('/admin/refresh') &&
      getRefreshToken()
    ) {
      if (!refreshing) {
        refreshing = true
        try {
          const token = await refreshAccessToken()
          flushQueue(token)
          // 重放当前请求
          config.headers = config.headers ?? {}
          config.headers.Authorization = `Bearer ${token}`
          return service(config)
        } catch (refreshError) {
          // 刷新失败：清空凭证并通知业务层登录已失效
          flushQueue('')
          clearTokens()
          window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
          return Promise.reject(refreshError)
        } finally {
          refreshing = false
        }
      }
      // 已有刷新流程进行中：挂起等待，刷新完成后复用新令牌重放
      return new Promise((resolve, reject) => {
        pendingQueue.push((token) => {
          if (token) {
            config.headers = config.headers ?? {}
            config.headers.Authorization = `Bearer ${token}`
            resolve(service(config))
          } else {
            reject(error)
          }
        })
      })
    }

    // 后端返回了统一响应结构（即使 HTTP 非 2xx，如 400/500），优先使用其中的 message
    if (data && typeof data.message === 'string' && data.message) {
      return Promise.reject(new BizError(bizCode ?? 0, data.message, config))
    }
    // 无响应体的网络错误（断网/超时等）：给出友好提示
    if (!error.response) {
      const msg =
        error.code === 'ECONNABORTED' ? '请求超时，请稍后再试' : '网络异常，请检查网络连接'
      return Promise.reject(new BizError(0, msg, config))
    }
    // 其它 HTTP 状态错误（响应体不是统一结构）
    return Promise.reject(
      new BizError(bizCode ?? error.response.status, `请求失败（HTTP ${error.response.status}）`, config),
    )
  },
)

/* -------------------------------------------------------------------------- */
/*                                   请求封装                                  */
/* -------------------------------------------------------------------------- */

/**
 * 泛型请求封装：自动解包统一响应结构 `{ code, message, data }`，
 * 直接返回业务数据 `data`，类型安全。
 *
 * @example
 * const page = await request<PageResult<PostListItem>>({ url: '/posts', method: 'get', params })
 */
export async function request<T>(config: RequestConfig): Promise<T> {
  const response = await service.request<ApiResponse<T>>(config)
  return response.data.data
}
