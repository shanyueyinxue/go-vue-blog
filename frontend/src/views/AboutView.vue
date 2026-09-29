<script setup lang="ts">
import { onMounted, ref } from 'vue'
import MdRenderer from '@/components/MdRenderer.vue'
import { storeToRefs } from 'pinia'
import { useSiteStore } from '@/stores/site'
import { markPageReady } from '@/composables/usePageReady'

const siteStore = useSiteStore()
const { config } = storeToRefs(siteStore)
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    await siteStore.fetchConfig()
  } finally {
    loading.value = false
    markPageReady()
  }
})
</script>

<template>
  <div class="blog-article">
    <div class="blog-article-card">
      <h1>关于</h1>
      <div v-if="loading" class="blog-empty">
        <i class="fa-solid fa-spinner fa-spin"></i>
        <span style="margin-left: 8px">加载中…</span>
      </div>
      <div v-else-if="!config.about_content" class="blog-empty">暂未填写关于内容</div>
      <div v-else class="blog-markdown">
        <MdRenderer :content="config.about_content" />
      </div>
    </div>
  </div>
</template>
