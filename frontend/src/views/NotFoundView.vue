<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { markPageReady } from '@/composables/usePageReady'

// 静态页无需数据加载，直接上报就绪（结束 SiteLayout 初始 loading）
onMounted(markPageReady)
</script>

<template>
  <div class="blog-404">
    <div class="blog-404-code">404</div>
    <h1 class="blog-404-title">页面不存在</h1>
    <p class="blog-404-desc">您访问的页面可能已被移除、更名，或地址输入有误。</p>
    <RouterLink to="/" class="blog-404-btn">
      <i class="fa-solid fa-house"></i>&ensp;返回首页
    </RouterLink>
  </div>
</template>
