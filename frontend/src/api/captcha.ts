/**
 * 验证码相关接口（docs/api.md 第 12 章，公开接口）
 *
 * - `POST /api/captcha/email`   发送邮箱验证码
 * - `POST /api/captcha/verify`  校验验证码
 */

import { request } from './http'
import type { EmailCaptchaParams, EmailCaptchaResult, VerifyCaptchaParams, VerifyCaptchaResult } from './types'

/** 发送邮箱验证码（docs/api.md 12.1，防评论垃圾） */
export function sendEmailCaptcha(params: EmailCaptchaParams) {
  return request<EmailCaptchaResult>({
    url: '/captcha/email',
    method: 'post',
    data: params,
  })
}

/** 校验验证码（docs/api.md 12.2，校验通过后验证码即作废） */
export function verifyCaptcha(params: VerifyCaptchaParams) {
  return request<VerifyCaptchaResult>({
    url: '/captcha/verify',
    method: 'post',
    data: params,
  })
}
