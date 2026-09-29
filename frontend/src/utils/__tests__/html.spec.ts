import { describe, expect, it } from 'vitest'
import { renderAdminHtml, renderIcpHtml } from '@/utils/html'

describe('renderAdminHtml 安全消毒', () => {
  it('允许安全链接', () => {
    expect(
      renderAdminHtml('<a href="https://beian.miit.gov.cn/" target="_blank">京ICP备xxx号</a>'),
    ).toContain('href="https://beian.miit.gov.cn/"')
  })

  it('脚本标签被移除', () => {
    expect(renderAdminHtml('<script>alert(1)</script>')).not.toContain('script')
    expect(renderAdminHtml('<script>alert(1)</script>')).toBe('')
  })

  it('javascript: 协议被移除', () => {
    const out = renderAdminHtml('<a href="javascript:alert(1)">点我</a>')
    expect(out).not.toContain('javascript:')
    expect(out).not.toContain('href=')
  })

  it('非 a 标签被移除', () => {
    expect(renderAdminHtml('<img src=x onerror=alert(1)><a href="/">链接</a>')).not.toContain('img')
  })

  it('强制 rel=noopener 并补 target=_blank', () => {
    expect(renderAdminHtml('<a href="https://a.com">x</a>')).toContain('rel="noopener noreferrer"')
    expect(renderAdminHtml('<a href="https://a.com">x</a>')).toContain('target="_blank"')
  })
})

describe('renderIcpHtml 备案号渲染', () => {
  it('HTML 链接直接渲染（消毒后）', () => {
    expect(renderIcpHtml('<a href="https://beian.miit.gov.cn/" target="_blank">京ICP备xxx号</a>')).toContain(
      'href="https://beian.miit.gov.cn/"',
    )
  })

  it('纯文本回退为默认链接', () => {
    const out = renderIcpHtml('京ICP备xxx号')
    expect(out).toContain('href="https://beian.miit.gov.cn/"')
    expect(out).toContain('rel="noopener noreferrer"')
    expect(out).toContain('京ICP备xxx号')
  })

  it('纯文本中的标签被转义，不会变成可执行 HTML', () => {
    const out = renderIcpHtml('<script>alert(1)</script>')
    expect(out).not.toContain('<script>')
    expect(out).toContain('&lt;script&gt;')
  })

  it('危险协议链接被消毒为无害链接', () => {
    const out = renderIcpHtml('<a href="javascript:alert(1)">x</a>')
    expect(out).not.toContain('javascript:')
    expect(out).toContain('x')
  })
})
