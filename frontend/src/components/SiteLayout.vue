<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useSiteStore } from '@/stores/site'
import { useTheme } from '@/composables/useTheme'
import { usePostCover } from '@/composables/usePostCover'
import { renderIcpHtml } from '@/utils/html'
import AudioPlayer from '@/components/AudioPlayer.vue'

const siteStore = useSiteStore()
const { config } = storeToRefs(siteStore)
const { theme, toggleTheme } = useTheme()
const router = useRouter()
const route = useRoute()
const searchQuery = ref('')
const showSearch = ref(false)
const hiddenMenu = ref(false)
const showMenuItems = ref(false)
const isHome = computed(() => route.name === 'home')
const menuColor = ref(isHome.value)
// 悬浮音乐播放器出现时给页脚预留空间，避免遮挡（播放器固定在左下角）
const playerReady = ref(false)
let lastScrollTop = 0

/** 站点配置扩展字段（extra） */
const extra = computed(() => config.value.extra ?? {})

/* ----------------------------- 卡片透明度 / 导航栏透明度 / 页面背景 ----------------------------- */
// extra.card_opacity（0-1，默认 1）：卡片背景透明（文字不透明，color-mix 实现）
const cardOpacity = computed(() => {
  const v = Number(extra.value.card_opacity)
  return Number.isFinite(v) && v >= 0 && v <= 1 ? v : 1
})
// extra.card_text_fade：开启后整卡（含文字）同步变淡
const cardTextFade = computed(() => extra.value.card_text_fade === true)
// extra.nav_opacity（0-1，默认 1）：导航栏背景透明（文字不透明，color-mix 实现）
const navOpacity = computed(() => {
  const v = Number(extra.value.nav_opacity)
  return Number.isFinite(v) && v >= 0 && v <= 1 ? v : 1
})
// extra.nav_text_fade：开启后整条导航栏（含文字）同步变淡
const navTextFade = computed(() => extra.value.nav_text_fade === true)
// extra.post_cover_enabled：开启后文章详情页把封面图作为整页固定背景
const postCoverEnabled = computed(() => extra.value.post_cover_enabled === true)
const postCover = usePostCover()

// extra.nav_bg_image（浅色）/ extra.nav_bg_image_dark（深色）为 {导航label: 图片URL} 映射，
// 作用于 .blog-shell 整页背景（按当前页面切换）；深色主题优先深色图，该页面缺失时回退浅色；
// 路由无匹配导航项（搜索/详情/404）时用默认背景
const navItems = [
  { to: '/', icon: 'fa-house', label: '首页' },
  { to: '/posts', icon: 'fa-file-lines', label: '文章' },
  { to: '/categories', icon: 'fa-folder', label: '分类' },
  { to: '/tags', icon: 'fa-tags', label: '标签' },
  { to: '/archive', icon: 'fa-box-archive', label: '归档' },
  { to: '/works', icon: 'fa-diagram-project', label: '作品' },
  { to: '/friends', icon: 'fa-link', label: '友链' },
  { to: '/about', icon: 'fa-id-card', label: '关于' },
]

const currentNavItem = computed(() => navItems.find((item) => item.to === route.path) ?? null)

const navBgImage = computed(() => {
  const label = currentNavItem.value?.label
  if (!label) return ''
  if (theme.value === 'dark') {
    const dark = extra.value.nav_bg_image_dark
    if (dark && typeof dark === 'object') {
      const v = (dark as Record<string, unknown>)[label]
      if (typeof v === 'string' && v) return v
    }
  }
  const light = extra.value.nav_bg_image
  if (light && typeof light === 'object') {
    const v = (light as Record<string, unknown>)[label]
    return typeof v === 'string' ? v : ''
  }
  return ''
})

const shellStyle = computed<Record<string, string>>(() => {
  const styles: Record<string, string> = {}
  if (cardOpacity.value !== 1) {
    styles['--blog-card-opacity'] = String(cardOpacity.value)
  }
  if (navOpacity.value !== 1) {
    styles['--blog-nav-opacity'] = String(navOpacity.value)
  }
  // 按页面切换的整页背景（nav_bg_image / nav_bg_image_dark）
  if (navBgImage.value) {
    styles.backgroundImage = `url(${navBgImage.value})`
    styles.backgroundSize = 'cover'
    styles.backgroundPosition = 'center'
    styles.backgroundRepeat = 'no-repeat'
    styles.backgroundAttachment = 'fixed' // 固定在页面顶部，避免滚动时背景图跟随滚动
  }
  // 文章详情封面固定背景（extra.post_cover_enabled 开启且有封面时，优先级高于页面背景）
  if (postCoverEnabled.value && postCover.value) {
    styles.backgroundImage = `url(${postCover.value})`
    styles.backgroundAttachment = 'fixed'
    styles.backgroundSize = 'cover'
    styles.backgroundPosition = 'center'
    styles.backgroundRepeat = 'no-repeat'
    styles.backgroundAttachment = 'fixed' // 固定在页面顶部，避免滚动时背景图跟随滚动
  }
  return styles
})

/* ----------------------------- 页脚背景色 / 回到顶部 ----------------------------- */
// extra.footer_inner_bg：页脚内层背景色（支持带透明度颜色），为空则不设置
const footerInnerBg = computed(() => {
  const v = extra.value.footer_inner_bg
  return typeof v === 'string' && v ? v : ''
})

const showBackTop = ref(false)

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

/* ----------------------------- 导航/菜单交互 ----------------------------- */

// 站点配置 extra.hideNavItems（数组，元素为导航项 label 名称）中列出的项不显示
const hiddenNavItems = computed<Set<string>>(() => {
  const raw = config.value.extra?.hideNavItems
  if (!Array.isArray(raw)) return new Set()
  return new Set(raw.filter((v): v is string => typeof v === 'string'))
})

const visibleNavItems = computed(() => navItems.filter((item) => !hiddenNavItems.value.has(item.label)))

function handleScroll() {
  const top = document.documentElement.scrollTop || document.body.scrollTop
  hiddenMenu.value = top > lastScrollTop && top > 80
  menuColor.value = isHome.value && top <= window.innerHeight - 100
  showBackTop.value = top > 300
  lastScrollTop = top
}

watch(isHome, (value) => {
  menuColor.value = value && (document.documentElement.scrollTop || document.body.scrollTop) <= window.innerHeight - 100
})

function submitSearch() {
  const q = searchQuery.value.trim()
  showSearch.value = false
  router.push({ name: 'search', query: q ? { q } : {} })
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true })
  // 导航栏依赖站点配置（标题、hideNavItems、背景图等），进入任意页面都需拉取
  siteStore.fetchConfig().catch(() => { })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <div
    class="blog-shell"
    :style="shellStyle"
    :class="{ 'card-text-fade': cardTextFade, 'nav-text-fade': navTextFade }"
  >
    <header class="blog-menu" :class="{ hidden: hiddenMenu, 'menu-color': menuColor }">
      <nav class="blog-menu-desktop">
        <RouterLink class="blog-menu-title" to="/">
          <span>{{ config.title || '我的博客' }}</span>
        </RouterLink>
        <RouterLink v-for="item in visibleNavItems" :key="item.to" :to="item.to">
          <i class="fa-solid fa-fw" :class="item.icon"></i>
          <span>&ensp;{{ item.label }}</span>
        </RouterLink>
        <button class="blog-nav-search-icon" type="button" aria-label="搜索" @click="showSearch = true">
          <i class="fa-solid fa-magnifying-glass"></i>
        </button>
        <button class="blog-nav-search-icon blog-theme-toggle" type="button"
          :aria-label="theme === 'dark' ? '切换到浅色模式' : '切换到黑暗模式'" :title="theme === 'dark' ? '切换到浅色模式' : '切换到黑暗模式'"
          @click="toggleTheme">
          <i class="fa-solid" :class="theme === 'dark' ? 'fa-sun' : 'fa-moon'"></i>
        </button>
      </nav>

      <nav class="blog-menu-mobile">
        <div class="blog-menu-mobile-title" @click="showMenuItems = !showMenuItems">
          <i class="fa-solid fa-bars fa-fw"></i>
          <span>&emsp;{{ config.title || '我的博客' }}</span>
        </div>
        <Transition name="blog-slide">
          <div v-show="showMenuItems" class="blog-menu-mobile-items">
            <RouterLink v-for="item in visibleNavItems" :key="item.to" :to="item.to" @click="showMenuItems = false">
              <div class="blog-menu-mobile-item">
                <div class="blog-menu-icon">
                  <i class="fa-solid fa-fw" :class="item.icon"></i>
                </div>
                <div class="blog-menu-label">{{ item.label }}</div>
              </div>
            </RouterLink>
            <button class="blog-nav-search-icon" type="button" aria-label="搜索" @click="showSearch = true">
              <i class="fa-solid fa-magnifying-glass"></i>
            </button>
            <button class="blog-nav-search-icon blog-theme-toggle" type="button"
              :aria-label="theme === 'dark' ? '切换到浅色模式' : '切换到黑暗模式'" :title="theme === 'dark' ? '切换到浅色模式' : '切换到黑暗模式'"
              @click="toggleTheme">
              <i class="fa-solid" :class="theme === 'dark' ? 'fa-sun' : 'fa-moon'"></i>
            </button>
          </div>
        </Transition>
      </nav>
    </header>

    <Transition name="blog-fade">
      <div v-if="showMenuItems" class="blog-menu-curtain" @click="showMenuItems = false"></div>
    </Transition>

    <Transition name="blog-fade">
      <div v-if="showSearch" class="blog-search-modal" @click.self="showSearch = false">
        <form class="blog-search-popup" @submit.prevent="submitSearch">
          <button type="button" class="blog-search-close" @click="showSearch = false">
            <i class="fa-solid fa-xmark"></i>
          </button>
          <h2>搜索文章</h2>
          <div class="blog-search-box">
            <i class="fa-solid fa-magnifying-glass"></i>
            <input v-model="searchQuery" type="search" placeholder="输入标题或摘要关键字…" autofocus />
          </div>
          <button type="submit" class="blog-btn">搜索</button>
        </form>
      </div>
    </Transition>

    <main class="blog-main">
      <RouterView />
    </main>

    <footer class="blog-footer" :class="{ 'blog-footer-with-player': playerReady }">
      <div class="blog-footer-inner" :style="footerInnerBg ? { backgroundColor: footerInnerBg } : undefined">
        <div>
          &copy; {{ new Date().getFullYear() }} {{ config.title || '我的博客' }}
          <span v-if="config.subtitle"> · {{ config.subtitle }}</span>
        </div>
        <div class="blog-footer-icp">
          <!-- 备案号/公安备案号支持 HTML 链接，渲染前经 DOMPurify 消毒（见 utils/html.ts） -->
          <span v-if="config.icp" v-html="renderIcpHtml(config.icp)"></span>
          <span v-if="config.police_icp" v-html="renderIcpHtml(config.police_icp)"></span>
        </div>
      </div>
    </footer>

    <!-- 回到顶部按钮：右下角贴底，滚动超过 300px 时显示 -->
    <Transition name="blog-fade">
      <button v-if="showBackTop" class="blog-back-top" type="button" aria-label="回到顶部" title="回到顶部"
        @click="scrollToTop">
        <i class="fa-solid fa-arrow-up"></i>
      </button>
    </Transition>

    <!-- 全局音乐播放器（站点配置 extra.aplayer.enabled 开启且存在启用曲目时显示） -->
    <AudioPlayer @ready="playerReady = true" @unready="playerReady = false" />
  </div>
</template>
