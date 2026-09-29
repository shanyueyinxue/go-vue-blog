import { defineStore } from 'pinia'
// 直连子模块而非 '@/api' 桶：站点 store 位于入口 chunk（main.ts 静态导入），
// 走桶会连带全部 API 模块进入入口，拖长白屏期
import { getPublicConfig } from '@/api/configs'
import type { PublicConfig } from '@/api/types'
import { applySiteTheme } from '@/composables/useTheme'

/**
 * 移除 CSS 值中的注释：
 * - 块注释 /* ... *​/ 直接移除
 * - 行注释 // ... 到行尾移除（引号内与 url(...) 内的 // 保留）
 */
export function stripCssComments(value: string): string {
  const v = value.replace(/\/\*[\s\S]*?\*\//g, '')
  let result = ''
  let i = 0
  let quote: string | null = null
  let inUrl = false
  while (i < v.length) {
    const ch = v[i]
    // 转义字符：整体拷贝两个字符
    if (ch === '\\') {
      result += ch + (v[i + 1] ?? '')
      i += 2
      continue
    }
    // 引号内：原样保留，直到闭合引号
    if (quote) {
      result += ch
      if (ch === quote) quote = null
      i++
      continue
    }
    if (ch === '"' || ch === "'") {
      quote = ch
      result += ch
      i++
      continue
    }
    // url( 起始（含空白与引号变体），内部 // 不算注释
    const lower = v.slice(i).toLowerCase()
    const urlStart = lower.match(/^url\(\s*['"]?/)
    if (urlStart) {
      inUrl = true
      result += v.slice(i, i + urlStart[0].length)
      i += urlStart[0].length
      continue
    }
    if (inUrl) {
      result += ch
      if (ch === ')') inUrl = false
      i++
      continue
    }
    // 行注释：// 到行尾，直接截断
    if (ch === '/' && v[i + 1] === '/') {
      return result
    }
    result += ch
    i++
  }
  return result
}

/** 解析平坦 CSS 变量文本（每行一个 --name: value;），返回 [变量名, 值] 列表 */
export function parseCustomCssVars(text: string): Array<[string, string]> {
  const result: Array<[string, string]> = []
  for (const raw of text.split(/\r?\n/)) {
    const line = raw.trim()
    if (!line || line.startsWith('//') || line.startsWith('/*')) continue
    // 注入前先移除值后面的注释，避免注释文本混入变量值
    const cleaned = stripCssComments(line)
    const m = cleaned.match(/^(--[\w-]+)\s*:\s*(.+?);?\s*$/)
    if (m?.[1] && m[2]) {
      result.push([m[1].trim(), m[2].trim()])
    }
  }
  return result
}

/** 注入自定义 CSS 变量的 <style> 元素（:root 通用 + [data-theme='dark'] 黑暗主题专属） */
let customVarsStyleEl: HTMLStyleElement | null = null

/**
 * 将站点配置中的自定义 CSS 变量注入页面：
 * - text 为通用变量，写入 :root（所有主题生效）
 * - darkText 为黑暗主题专属变量，写入 [data-theme='dark']（仅黑暗主题生效）
 * 均通过一个 <style> 标签实现，切换主题时由 CSS 自动决定生效范围。
 */
function applyCustomCssVars(
  text: string | undefined,
  darkText: string | undefined,
  enabled: number | undefined,
) {
  if (!customVarsStyleEl) {
    customVarsStyleEl = document.createElement('style')
    customVarsStyleEl.id = 'blog-custom-css-vars'
    document.head.appendChild(customVarsStyleEl)
  }

  const blocks: string[] = []
  if (enabled === 1 && text) {
    const vars = parseCustomCssVars(text)
    if (vars.length) {
      blocks.push(`:root{\n${vars.map(([n, v]) => `${n}:${v};`).join('\n')}\n}`)
    }
  }
  if (enabled === 1 && darkText) {
    const darkVars = parseCustomCssVars(darkText)
    if (darkVars.length) {
      blocks.push(`[data-theme='dark']{\n${darkVars.map(([n, v]) => `${n}:${v};`).join('\n')}\n}`)
    }
  }
  customVarsStyleEl.textContent = blocks.join('\n')
}

interface SiteState {
  config: Partial<PublicConfig>
  loading: boolean
  loaded: boolean
}

let pendingConfigRequest: Promise<Partial<PublicConfig>> | null = null

export const useSiteStore = defineStore('site', {
  state: (): SiteState => ({
    config: {},
    loading: false,
    loaded: false,
  }),

  getters: {
    title: (state) => state.config.title || '我的博客',
    description: (state) => state.config.description || state.config.subtitle || '个人博客',
  },

  actions: {
    async fetchConfig(force = false) {
      if (this.loaded && !force) return this.config
      if (pendingConfigRequest) return pendingConfigRequest

      this.loading = true
      pendingConfigRequest = (async () => {
        try {
          this.config = await getPublicConfig()
          this.loaded = true
          // 应用后台配置的主题值（站点默认主题）与自定义 CSS 变量
          applySiteTheme(this.config.theme)
          applyCustomCssVars(
            this.config.custom_css_vars,
            this.config.custom_css_vars_dark,
            this.config.custom_css_enabled,
          )
          return this.config
        } finally {
          this.loading = false
          pendingConfigRequest = null
        }
      })()
      return pendingConfigRequest
    },
  },
})
