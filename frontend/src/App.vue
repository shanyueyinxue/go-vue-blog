<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { useHead } from '@vueuse/head'
import { useSiteStore } from '@/stores/site'

const siteStore = useSiteStore()

// 应用启动即加载全局配置，公开页和后台页都能复用（main.ts 已预取，此处命中缓存）
siteStore.fetchConfig()

// 依据当前路由 meta 输出标题与描述；站点级 title/description 作为兜底。
// 页面组件内更深的 useHead（如 PostDetailView）优先级更高，可再覆盖 keywords/description
const route = useRoute()
useHead({
  title: computed(() => {
    const page = route.meta.title
    return page ? `${page} - ${siteStore.title}` : siteStore.title
  }),
  meta: computed(() => [
    {
      name: 'description',
      content: route.meta.description || siteStore.description,
    },
  ]),
})
</script>

<template>
  <RouterView />
</template>
