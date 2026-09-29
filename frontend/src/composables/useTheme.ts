/**
 * 主题（浅色 / 黑暗）切换组合式函数
 *
 * 主题通过 <html data-theme="light|dark"> 属性生效（见 theme-vars.css）。
 * 仅用户**手动切换**时才写入 localStorage；未手动选择时依次跟随
 * 站点配置 theme 与系统 prefers-color-scheme，且不会产生持久化记录。
 */
import { ref, watch } from 'vue'

export type Theme = 'light' | 'dark'

const STORAGE_KEY = 'blog-theme'

function systemTheme(): Theme {
  return typeof window !== 'undefined' &&
    window.matchMedia?.('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light'
}

/** 读取当前生效主题（供组件初始化状态使用） */
function readTheme(): Theme {
  return document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light'
}

/** 仅应用主题到 <html>，不持久化 */
function applyTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme
}

/** 用户手动选择时持久化（读取时优先于站点配置与系统偏好） */
function persistTheme(theme: Theme) {
  try {
    localStorage.setItem(STORAGE_KEY, theme)
  } catch {
    // 隐私模式等写入失败时静默忽略
  }
}

const current = ref<Theme>(readTheme())
// 仅监听变化应用主题；持久化只在用户手动切换时进行（见 useTheme）
watch(current, (theme) => applyTheme(theme))

/** 在应用挂载前调用一次，避免首屏闪烁（见 main.ts） */
export function initTheme() {
  const saved = localStorage.getItem(STORAGE_KEY)
  const theme: Theme = saved === 'light' || saved === 'dark' ? saved : systemTheme()
  applyTheme(theme)
  current.value = theme
  return theme
}

/**
 * 应用站点配置中的主题值（站点默认主题）。
 * 优先级：用户手动选择（localStorage 有记录）> 站点配置 theme > 系统偏好。
 * 仅接受安全的 CSS 标识符（字母/数字/连字符），其余值忽略。
 */
export function applySiteTheme(theme: string | undefined) {
  if (!theme) return
  if (!/^[a-z0-9-]+$/i.test(theme)) return
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved === 'light' || saved === 'dark') return // 用户手动切换过，保留用户选择
  document.documentElement.dataset.theme = theme
  current.value = readTheme()
}

export function useTheme() {
  /** 用户手动设置主题：应用并持久化 */
  function setTheme(theme: Theme) {
    current.value = theme
    persistTheme(theme)
  }

  /** 用户手动切换主题：应用并持久化 */
  function toggleTheme() {
    const next: Theme = current.value === 'dark' ? 'light' : 'dark'
    current.value = next
    persistTheme(next)
  }

  return { theme: current, setTheme, toggleTheme }
}
