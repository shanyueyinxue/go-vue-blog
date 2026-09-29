<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { listCategories } from '@/api'
import type { Category } from '@/api'
import { markPageReady } from '@/composables/usePageReady'

const categories = ref<Category[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    categories.value = await listCategories()
  } finally {
    loading.value = false
    markPageReady()
  }
})
</script>

<template>
  <div>
    <h1 class="blog-section-title">分类</h1>
    <div class="blog-categories-page">
      <div v-if="loading" class="blog-empty">
        <i class="fa-solid fa-spinner fa-spin"></i>
        <span style="margin-left: 8px">加载中…</span>
      </div>
      <div v-else-if="categories.length === 0" class="blog-empty">暂无分类</div>
      <div v-else class="blog-categories-grid">
        <RouterLink
          v-for="c in categories"
          :key="c.id"
          :to="{ path: '/posts', query: { category_id: c.id } }"
          class="blog-category-card"
          :title="`查看「${c.name}」分类的文章`"
        >
          <i class="fa-solid fa-folder fa-fw"></i>
          <span class="blog-category-name">{{ c.name }}</span>
          <span class="blog-category-count">{{ c.post_count }} 篇</span>
        </RouterLink>
      </div>
    </div>
  </div>
</template>
