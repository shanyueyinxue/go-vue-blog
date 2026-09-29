/**
 * 首屏数据加载完成标记（模块级共享状态）
 *
 * 前台数据页在首屏数据加载完成后调用 markPageReady()；
 * SiteLayout 的初始 loading 同时满足「window load 已触发 + 页面已上报」才隐藏，
 * 保证慢网络下 loading 覆盖真实加载过程而非一闪而过。
 */
import { ref } from 'vue'

const pageReady = ref(false)

/** 读取页面就绪状态（SiteLayout 监听） */
export function usePageReady() {
  return pageReady
}

/** 当前页面首屏数据加载完成（无论成败都调用，放 finally 中） */
export function markPageReady() {
  pageReady.value = true
}
