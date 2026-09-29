<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listPosts } from '@/api'
import type { PageResult, PostListItem } from '@/api'
import { markPageReady } from '@/composables/usePageReady'
import BlogPagination from '@/components/BlogPagination.vue'
import BlogPostCard from '@/components/BlogPostCard.vue'

const posts = ref<PageResult<PostListItem>>({ list: [], total: 0, page: 1, pageSize: 10 })
const route = useRoute()
const router = useRouter()

function parsePage(value: unknown): number {
  const page = Number(value)
  return Number.isInteger(page) && page > 0 ? page : 1
}

function parseId(value: unknown): number | undefined {
  const id = Number(value)
  return Number.isInteger(id) && id > 0 ? id : undefined
}

const page = ref(parsePage(route.query.page))
const pageSize = 10
const loading = ref(false)

// 支持 ?tag_id= / ?category_id= 筛选（从分类/标签等入口进入时生效）
const tagId = computed(() => parseId(route.query.tag_id))
const categoryId = computed(() => parseId(route.query.category_id))

const totalPages = computed(() => Math.max(1, Math.ceil(posts.value.total / pageSize)))

async function load() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, pageSize, sort: 'latest' }
    if (tagId.value) params.tag_id = tagId.value
    if (categoryId.value) params.category_id = categoryId.value
    posts.value = await listPosts(params)
  } finally {
    loading.value = false
  }
}

function changePage(next: number) {
  if (next < 1 || next > totalPages.value) return
  page.value = next
  router.replace({ query: { ...route.query, page: next } })
  load()
}

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

// 分类/标签筛选变化时回到第 1 页并重新加载
watch([tagId, categoryId], () => {
  if (page.value !== 1) {
    page.value = 1
    router.replace({ query: { ...route.query, page: 1 } })
  }
  load()
})

onMounted(async () => {
  try {
    await load()
  } finally {
    markPageReady()
  }
})
</script>

<template>
  <div>
    <h1 class="blog-section-title">文章</h1>
    <div class="blog-posts" style="margin: auto">
      <div v-if="loading" class="blog-empty">
        <i class="fa-solid fa-spinner fa-spin"></i>
        <span style="margin-left: 8px">加载中…</span>
      </div>

      <div v-else-if="posts.list.length === 0" class="blog-empty">暂无文章</div>

      <template v-else>
        <BlogPostCard v-for="post in posts.list" :key="post.id" :post="post" />

        <BlogPagination
          v-if="posts.total > pageSize"
          :total="posts.total"
          :page="page"
          :page-size="pageSize"
          @change="changePage"
        />
      </template>
    </div>
  </div>
</template>
