/**
 * 应用入口
 *
 * 加载策略：入口 chunk 只静态导入轻量核心（vue / pinia / 站点 store / 主题变量等），
 * 并立即用「独立 Vue 实例」渲染全局 loading（AppLoading.vue），
 * 之后再动态导入重资源（FontAwesome、element-plus、theme.css、App.vue、router、head）。
 * 这样 loading 能真正覆盖「资源下载 + 配置预取 + 首屏数据」全过程；
 * 若全部静态导入，这些资源会在 main.ts 执行前就已加载完（ES module 提升），
 * loading 出现时资源加载早已结束（这正是 App.vue 内渲染 loading 的缺陷）。
 */
import { createApp, watch } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { initTheme } from './composables/useTheme'
import { useSiteStore } from './stores/site'
import { usePageReady } from './composables/usePageReady'
import AppLoading from './components/AppLoading.vue'
// 轻量基础样式随入口加载：提供 --blog-* CSS 变量，AppLoading 与后续页面都依赖它
import './assets/theme-vars.css'
import './assets/main.css'

// 挂载前先应用主题（系统/用户偏好），避免页面闪烁
initTheme()

const pinia = createPinia()
setActivePinia(pinia)
const siteStore = useSiteStore()

/* ----------------------------- 1. 独立 loading 实例 ----------------------------- */
// 在加载任何重资源之前先渲染 loading（入口 chunk 很小，白屏期≈入口下载期）
const loadingApp = createApp(AppLoading)
loadingApp.use(pinia)
loadingApp.mount('#app-loading')

// loading 结束条件：window load 已触发 且 当前页面已调用 markPageReady()；10s 兜底
const pageReady = usePageReady()
let windowLoaded = document.readyState === 'complete'
let loadingHidden = false

function hideLoading() {
  if (loadingHidden) return
  loadingHidden = true
  loadingApp.unmount()
}

function maybeHideLoading() {
  if (pageReady.value && windowLoaded) {
    hideLoading()
  }
}

watch(pageReady, maybeHideLoading)
if (!windowLoaded) {
  window.addEventListener(
    'load',
    () => {
      windowLoaded = true
      maybeHideLoading()
    },
    { once: true },
  )
}
setTimeout(hideLoading, 10000)

/* ----------------------------- 2. 并行：取配置 + 动态加载重资源 ----------------------------- */
// fetchConfig 内部应用站点默认主题与自定义 CSS 变量（AppLoading 会自动跟随配色）
const configPromise = siteStore.fetchConfig().catch(() => {})

// 动态加载重资源（loading 显示期间下载）；CSS 依赖顺序保持：theme.css 在 element-plus 之后
const cssLoaded = (async () => {
  await import('@fortawesome/fontawesome-free/css/all.min.css')
  await import('element-plus/dist/index.css')
  await import('./assets/theme.css')
})()

const [appModule, routerModule, headModule] = await Promise.all([
  import('./App.vue'),
  import('./router'),
  import('@vueuse/head'),
])
const { default: App } = appModule
const { default: router } = routerModule
const { createHead } = headModule

await cssLoaded
await configPromise

/* ----------------------------- 3. 创建并挂载主应用 ----------------------------- */
const app = createApp(App)

// 仅注册实际用到的图标（后台菜单通过字符串名动态渲染 <component :is="...">）。
// 新增图标时需在此补充注册，避免整包引入导致体积膨胀。
import {
  ChatDotRound,
  CollectionTag,
  Document,
  Expand,
  FolderOpened,
  Fold,
  Headset,
  Link,
  Moon,
  Odometer,
  PictureFilled,
  PriceTag,
  Sunny,
  Tools,
  User,
  Postcard
} from '@element-plus/icons-vue'
const icons = {
  ChatDotRound,
  CollectionTag,
  Document,
  Expand,
  FolderOpened,
  Fold,
  Headset,
  Link,
  Moon,
  Odometer,
  PictureFilled,
  PriceTag,
  Sunny,
  Tools,
  User,
  Postcard
}
for (const [key, component] of Object.entries(icons)) {
  app.component(key, component)
}

app.use(pinia)
app.use(createHead())
app.use(router)

app.mount('#app')
