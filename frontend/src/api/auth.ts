/**
 * 认证相关接口（docs/api.md 第 3 章）
 *
 * - `POST /api/admin/login`         管理员登录（无需 JWT）
 * - `POST /api/admin/refresh`       刷新访问 Token
 * - `GET  /api/admin/profile`       获取管理员信息（需 JWT）
 * - `PUT  /api/admin/profile`       修改个人信息（需 JWT）
 * - `PUT  /api/admin/profile/password` 修改管理员密码（需 JWT）
 * - `POST /api/forgot-password/send-code` 忘记密码-发送邮箱验证码（无需 JWT）
 * - `POST /api/forgot-password/reset`     忘记密码-重置密码（无需 JWT）
 */

import { request } from './http'
import type {
  EmailCaptchaResult,
  ForgotPasswordResetParams,
  ForgotPasswordSendParams,
  LoginParams,
  LoginResult,
  ProfileUpdateParams,
  RefreshParams,
  RefreshResult,
  UpdatePasswordParams,
  User,
} from './types'

/** 管理员登录（docs/api.md 3.1） */
export function login(params: LoginParams) {
  return request<LoginResult>({
    url: '/admin/login',
    method: 'post',
    data: params,
  })
}

/** 刷新访问 Token（docs/api.md 3.2） */
export function refreshToken(params: RefreshParams) {
  return request<RefreshResult>({
    url: '/admin/refresh',
    method: 'post',
    data: params,
  })
}

/** 获取管理员信息（docs/api.md 3.3） */
export function getProfile() {
  return request<User>({
    url: '/admin/profile',
    method: 'get',
  })
}

/** 修改个人信息（docs/api.md 3.4） */
export function updateProfile(params: ProfileUpdateParams) {
  return request<User>({
    url: '/admin/profile',
    method: 'put',
    data: params,
  })
}

/** 修改管理员密码（docs/api.md 3.5） */
export function updatePassword(params: UpdatePasswordParams) {
  return request<{ message: string }>({
    url: '/admin/profile/password',
    method: 'put',
    data: params,
  })
}

/** 忘记密码：发送邮箱验证码（docs/api.md 3.6，邮箱未注册也返回成功，防用户枚举） */
export function forgotPasswordSendCode(params: ForgotPasswordSendParams) {
  return request<EmailCaptchaResult>({
    url: '/forgot-password/send-code',
    method: 'post',
    data: params,
  })
}

/** 忘记密码：校验验证码并重置密码（docs/api.md 3.7，成功后旧 Token 全部失效） */
export function forgotPasswordReset(params: ForgotPasswordResetParams) {
  return request<{ message: string }>({
    url: '/forgot-password/reset',
    method: 'post',
    data: params,
  })
}
