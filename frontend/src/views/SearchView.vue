<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { listPosts } from '@/api'
import type { PageResult, PostListItem } from '@/api'
import { markPageReady } from '@/composables/usePageReady'
import { formatDate } from '@/utils/format'

const route = useRoute()
const keyword = computed(() => String(route.query.q ?? '').trim())
const posts = ref<PageResult<PostListItem>>({ list: [], total: 0, page: 1, pageSize: 10 })
const page = ref(1)
const pageSize = 10
const loading = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(posts.value.total / pageSize)))

async function load() {
  if (!keyword.value) {
    posts.value = { list: [], total: 0, page: 1, pageSize }
    return
  }
  loading.value = true
  try {
    posts.value = await listPosts({ page: page.value, pageSize, keyword: keyword.value })
  } finally {
    loading.value = false
  }
}

function changePage(next: number) {
  if (next < 1 || next > totalPages.value) return
  page.value = next
  load()
}

watch(keyword, () => {
  page.value = 1
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
  <div class="blog-article">
    <div class="blog-article-card">
      <h1>搜索结果</h1>
      <div v-if="keyword" class="blog-article-meta">
        关键词：{{ keyword }} · 共 {{ posts.total }} 篇
      </div>

      <div v-if="loading" class="blog-empty">
        <i class="fa-solid fa-spinner fa-spin"></i>
        <span style="margin-left: 8px">搜索中…</span>
      </div>
      <div v-else-if="!keyword || posts.list.length === 0" class="blog-empty">暂无搜索结果</div>

      <div v-else>
        <article v-for="post in posts.list" :key="post.id" class="blog-post" style="margin-top: 0">
          <RouterLink :to="`/post/${post.slug}`">
            <h2 class="blog-post-title">{{ post.title }}</h2>
          </RouterLink>
          <div class="blog-post-meta">
            <span v-if="post.category">
              <i class="fa-solid fa-bookmark fa-fw"></i>
              {{ post.category.name }}
            </span>
            <span>
              <i class="fa-solid fa-calendar fa-fw"></i>
              {{ formatDate(post.published_at) }}
            </span>
          </div>
          <div class="blog-post-excerpt">{{ post.excerpt || '暂无摘要' }}</div>
          <RouterLink :to="`/post/${post.slug}`" class="blog-go-post">阅读全文</RouterLink>
        </article>

        <div v-if="posts.total > pageSize" class="blog-pagination">
          <button :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
          <span class="blog-page-current">{{ page }}</span>
          <span>/ {{ totalPages }}</span>
          <button :disabled="page >= totalPages" @click="changePage(page + 1)">下一页</button>
        </div>
      </div>
    </div>
  </div>
</template>
