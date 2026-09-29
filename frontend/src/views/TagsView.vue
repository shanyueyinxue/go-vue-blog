<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { listTags } from '@/api'
import type { Tag } from '@/api'
import { markPageReady } from '@/composables/usePageReady'

const tags = ref<Tag[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    tags.value = await listTags()
  } finally {
    loading.value = false
    markPageReady()
  }
})
</script>

<template>
  <div>
    <h1 class="blog-section-title">标签</h1>
    <div class="blog-tags-page">
      <div v-if="loading" class="blog-empty">
        <i class="fa-solid fa-spinner fa-spin"></i>
        <span style="margin-left: 8px">加载中…</span>
      </div>
      <div v-else-if="tags.length === 0" class="blog-empty">暂无标签</div>
      <div v-else class="blog-tags-cloud">
        <RouterLink
          v-for="tag in tags"
          :key="tag.id"
          :to="{ path: '/posts', query: { tag_id: tag.id } }"
          class="blog-tag-chip"
          :title="`查看「${tag.name}」标签的文章`"
        >
          <i class="fa-solid fa-tag fa-fw"></i>
          {{ tag.name }}
          <span class="blog-tag-count">{{ tag.post_count }}</span>
        </RouterLink>
      </div>
    </div>
  </div>
</template>
