<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listFriends } from '@/api'
import type { Friend } from '@/api'
import { markPageReady } from '@/composables/usePageReady'

const friends = ref<Friend[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    friends.value = await listFriends()
  } finally {
    loading.value = false
    markPageReady()
  }
})
</script>

<template>
  <div>
    <h1 class="blog-section-title">友情链接</h1>
    <div class="blog-friends-list">
      <div v-if="loading" class="blog-empty">
        <i class="fa-solid fa-spinner fa-spin"></i>
        <span style="margin-left: 8px">加载中…</span>
      </div>
      <div v-else-if="friends.length === 0" class="blog-empty">暂无友链</div>
      <div v-else class="blog-friends-grid">
        <a
          v-for="f in friends"
          :key="f.id"
          :href="f.url"
          target="_blank"
          rel="noopener noreferrer"
          class="blog-friend-card"
        >
          <span class="blog-friend-icon">
            <img v-if="f.icon" :src="f.icon" :alt="f.name" />
            <i v-else class="fa-solid fa-link"></i>
          </span>
          <div class="blog-friend-info">
            <strong>{{ f.name }}</strong>
            <p>{{ f.description || '暂无描述' }}</p>
          </div>
        </a>
      </div>
    </div>
  </div>
</template>
