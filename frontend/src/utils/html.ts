/**
 * 管理员配置的 HTML 渲染工具（安全防护）
 *
 * 站点配置中的备案号等字段允许填写 HTML 标签（如
 * `<a href="https://beian.miit.gov.cn/" target="_blank">京ICP备xxx号</a>`），
 * 渲染前必须经 DOMPurify 消毒：
 * - 仅允许 `<a>` 标签与 href/target/rel/title 属性；
 * - DOMPurify 默认阻止 javascript: 等危险协议；
 * - 统一给 `<a>` 补 `rel="noopener noreferrer"`（缺 target 时补 `_blank`），
 *   防止 target="_blank" 的反向标签钓鱼（tabnabbing）。
 */
import DOMPurify from 'dompurify'

// 全局钩子：所有经 DOMPurify 输出的链接统一加 rel，避免重复在各处处理
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node.tagName === 'A') {
    node.setAttribute('rel', 'noopener noreferrer')
    if (!node.getAttribute('target')) {
      node.setAttribute('target', '_blank')
    }
  }
})

/** 渲染管理员的 HTML：仅允许安全链接，返回消毒后的 HTML */
export function renderAdminHtml(html: string): string {
  if (!html) return ''
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['a'],
    ALLOWED_ATTR: ['href', 'target', 'rel', 'title'],
    ALLOW_DATA_ATTR: false,
  })
}

/** HTML 转义（用于纯文本回退时包进链接，防止文本本身被当作标签） */
function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

/**
 * 渲染备案号字段：
 * - 已包含 `<a>` 标签时按 HTML 渲染（经消毒）；
 * - 纯文本时回退为指向工信部备案查询的默认链接（兼容旧配置）。
 */
export function renderIcpHtml(raw: string, fallbackHref = 'https://beian.miit.gov.cn/'): string {
  const sanitized = renderAdminHtml(raw)
  if (/<a[\s>]/i.test(sanitized)) {
    return sanitized
  }
  const href = escapeHtml(fallbackHref)
  return `<a href="${href}" target="_blank" rel="noopener noreferrer">${escapeHtml(raw)}</a>`
}
