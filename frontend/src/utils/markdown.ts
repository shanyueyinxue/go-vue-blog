import DOMPurify from 'dompurify'
import hljs from 'highlight.js/lib/common'
// 代码高亮配色不再引用固定主题样式（github.css），
// 改为使用 theme-vars.css 的 --hljs-* 变量（见 theme.css .hljs 规则），
// 以便深色模式下自动切换为深色配色。
import MarkdownIt from 'markdown-it'

const md: InstanceType<typeof MarkdownIt> = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: false,
  highlight(str: string, lang: string): string {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return `<pre class="hljs"><code>${hljs.highlight(str, { language: lang, ignoreIllegals: true }).value}</code></pre>`
      } catch {
        // ignore highlighting errors and fall through to escaped code block
      }
    }
    return `<pre class="hljs"><code>${md.utils.escapeHtml(str)}</code></pre>`
  },
})

// 参数类型由 markdown-it 的 RenderRule 推断（避免显式 any）
const defaultLinkOpen = md.renderer.rules.link_open

md.renderer.rules.link_open = (tokens, idx, options, env, self) => {
  const token = tokens[idx]
  if (token) {
    token.attrSet('target', '_blank')
    token.attrSet('rel', 'noopener noreferrer')
  }
  return defaultLinkOpen
    ? defaultLinkOpen(tokens, idx, options, env, self)
    : self.renderToken(tokens, idx, options)
}

export function renderMarkdown(src: string): string {
  if (!src) return ''
  const html = md.render(src)
  return DOMPurify.sanitize(html, {
    ADD_ATTR: ['target', 'rel', 'class'],
    USE_PROFILES: { html: true },
  })
}
