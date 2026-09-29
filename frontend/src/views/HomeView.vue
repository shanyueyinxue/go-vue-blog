<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { listPosts } from '@/api'
import type { PageResult, PostListItem, SocialLink } from '@/api'
import { useSiteStore } from '@/stores/site'
import { useTheme } from '@/composables/useTheme'
import { markPageReady } from '@/composables/usePageReady'
import BlogPagination from '@/components/BlogPagination.vue'
import BlogPostCard from '@/components/BlogPostCard.vue'

const posts = ref<PageResult<PostListItem>>({ list: [], total: 0, page: 1, pageSize: 10 })
const siteStore = useSiteStore()
const { config } = storeToRefs(siteStore)
const { theme } = useTheme()
const route = useRoute()
const router = useRouter()
const loading = ref(false)

function parsePage(value: unknown): number {
  const page = Number(value)
  return Number.isInteger(page) && page > 0 ? page : 1
}

function parseId(value: unknown): number | undefined {
  const id = Number(value)
  return Number.isInteger(id) && id > 0 ? id : undefined
}

const page = ref(parsePage(route.query.page))
const pageSize = ref(10)
const sort = ref<'latest' | 'views' | 'oldest'>('latest')

// 支持 ?tag_id= / ?category_id= 筛选（点标签/分类进入首页时生效）
const tagId = computed(() => parseId(route.query.tag_id))
const categoryId = computed(() => parseId(route.query.category_id))

const heroBg = ref<HTMLElement | null>(null)
const postsWrap = ref<HTMLElement | null>(null)

/** 站点配置扩展字段（extra） */
const extra = computed(() => config.value.extra ?? {})

/** 侧栏头像：黑暗主题下优先使用 extra.dark_avatar，未设置回退 config.avatar */
const avatar = computed(() => {
  if (theme.value === 'dark') {
    const dark = extra.value.dark_avatar
    if (typeof dark === 'string' && dark) return dark
  }
  return config.value.avatar
})

/** 首页背景图列表：黑暗主题下优先使用 extra.dark_home_images，未设置回退 home_images */
function resolveHomeImages(): string[] {
  const base = (config.value.home_images ?? []).filter(Boolean)
  if (theme.value === 'dark') {
    const dark = extra.value.dark_home_images
    if (Array.isArray(dark)) {
      const darkList = dark.filter((v): v is string => typeof v === 'string' && v.length > 0)
      if (darkList.length) return darkList
    }
  }
  return base
}

const heroGradients = [
  'linear-gradient(135deg, #a3ddfb 0%, #ffbbf4 100%)',
  'linear-gradient(135deg, #9abbf7 0%, #ffb7c5 100%)',
  'linear-gradient(135deg, #f6d5f7 0%, #fbe9d7 100%)',
  'linear-gradient(135deg, #bde0fe 0%, #cdb4db 100%)',
  'linear-gradient(135deg, #d8f3dc 0%, #95d5b2 100%)',
  'linear-gradient(135deg, #fbc4ab 0%, #ffc8dd 100%)',
]

const totalPages = computed(() => Math.max(1, Math.ceil(posts.value.total / pageSize.value)))

// 社交链接：数组顺序即显示顺序；兼容旧的对象格式 {"name":"url"}（按 key 排序）
const socialLinks = computed<SocialLink[]>(() => {
  const raw = config.value.social_links
  if (Array.isArray(raw)) {
    return raw
      .filter(
        (item): item is SocialLink =>
          !!item && typeof item === 'object' && typeof (item as SocialLink).url === 'string',
      )
      .map((item) => ({ name: String(item.name ?? ''), url: item.url }))
  }
  return Object.entries((raw ?? {}) as Record<string, unknown>).map(([name, url]) => ({
    name,
    url: String(url),
  }))
})

async function load() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: page.value,
      pageSize: pageSize.value,
      sort: sort.value,
    }
    if (tagId.value) params.tag_id = tagId.value
    if (categoryId.value) params.category_id = categoryId.value
    posts.value = await listPosts(params)
  } finally {
    loading.value = false
  }
}

function changePage(p: number) {
  if (p < 1 || p > totalPages.value) return
  page.value = p
  router.replace({ query: { ...route.query, page: p } })
  load()
}

function scrollToPosts() {
  postsWrap.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

// 首屏内容随滚动逐渐上移：blog-home-wrap 的 top 从 0 平滑过渡到 -80px
function updatePostsWrapTop() {
  const el = postsWrap.value
  if (!el) return
  const progress = Math.min(Math.max((window.scrollY / window.innerHeight) * 2, 0), 1)
  el.style.top = `${Math.round(-80 * progress)}px`
}

function setRandomHero() {
  const images = resolveHomeImages()
  const source = images.length ? images : heroGradients
  const index = Math.floor(Math.random() * source.length)
  if (heroBg.value) {
    const value = source[index] ?? source[0] ?? ''
    if (images.length) {
      heroBg.value.style.backgroundImage = `url(${value})`
    } else {
      heroBg.value.style.backgroundImage = value
    }
  }
}

// 主题切换（白天↔黑暗）时重新随机背景，使 dark_home_images 立即生效
watch(theme, () => setRandomHero())

watch(
  () => route.query.page,
  (value) => {
    const next = parsePage(value)
    if (next !== page.value) {
      page.value = next
      load()
    }
  },
)

// 标签/分类筛选变化时回到第 1 页并重新加载
watch([tagId, categoryId], () => {
  if (page.value !== 1) {
    page.value = 1
    router.replace({ query: { ...route.query, page: 1 } })
  }
  load()
})

onMounted(async () => {
  setRandomHero()
  updatePostsWrapTop()
  window.addEventListener('scroll', updatePostsWrapTop, { passive: true })
  try {
    await Promise.all([load(), siteStore.fetchConfig()])
    setRandomHero()
  } catch {
    // 配置加载失败不阻塞文章列表
  } finally {
    markPageReady()
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', updatePostsWrapTop)
})
</script>

<template>
  <div>
    <section class="blog-hero" @click="scrollToPosts">
      <div ref="heroBg" class="blog-hero-bg"></div>
      <div class="blog-hero-info">
        <span class="blog-hero-loop"></span>
        <span class="blog-hero-loop"></span>
        <span class="blog-hero-loop"></span>
        <span class="blog-hero-loop"></span>
        <div class="blog-hero-card">
          <div style="padding: 30px">
            <h1>{{ config.title || '我的博客' }}</h1>
            <h5>{{ config.subtitle || config.description || '记录生活、学习与工作中的思考' }}</h5>
          </div>
        </div>
      </div>
    </section>

    <div ref="postsWrap" class="blog-home-wrap">
      <section class="blog-posts">
        <div v-if="loading" class="blog-empty">
          <i class="fa-solid fa-spinner fa-spin"></i>
          <span style="margin-left: 8px">加载中…</span>
        </div>

        <div v-else-if="posts.list.length === 0" class="blog-empty">暂无文章</div>

        <template v-else>
          <BlogPostCard v-for="post in posts.list" :key="post.id" :post="post" />

          <BlogPagination v-if="posts.total > pageSize" :total="posts.total" :page="page" :page-size="pageSize"
            @change="changePage" />
        </template>
      </section>

      <aside class="blog-side-card">
        <div class="blog-side-card-inner">
          <div class="blog-side-body">
            <div class="blog-side-avatar">
              <img v-if="avatar" :src="avatar" :alt="config.title || 'avatar'" />
              <span v-else>{{ (config.title || '博')[0] }}</span>
            </div>
            <div class="blog-side-name">{{ config.side_name || config.title || '我的博客' }}</div>
            <div class="blog-side-desc">{{ config.description || config.subtitle || '个人博客' }}</div>
            <div v-if="socialLinks.length" class="blog-side-links">
              <a v-for="link in socialLinks" :key="link.name" :href="link.url" target="_blank"
                rel="noopener noreferrer">
                {{ link.name }}
              </a>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>
