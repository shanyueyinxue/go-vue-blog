<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listWorks } from '@/api'
import type { Work } from '@/api'
import { markPageReady } from '@/composables/usePageReady'

const works = ref<Work[]>([])
const loading = ref(false)

/** 解析后端返回的技术栈 JSON 字符串数组（解析失败返回空数组） */
function parseTechStack(raw: string): string[] {
  if (!raw) return []
  try {
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? arr.filter((v): v is string => typeof v === 'string') : []
  } catch {
    return []
  }
}

onMounted(async () => {
  loading.value = true
  try {
    works.value = await listWorks()
  } finally {
    loading.value = false
    markPageReady()
  }
})
</script>

<template>
  <div>
    <h1 class="blog-section-title">作品集</h1>
    <div class="blog-works-list">
      <div v-if="loading" class="blog-empty">
        <i class="fa-solid fa-spinner fa-spin"></i>
        <span style="margin-left: 8px">加载中…</span>
      </div>
      <div v-else-if="works.length === 0" class="blog-empty">暂无作品</div>
      <div v-else class="blog-works-grid">
        <div v-for="w in works" :key="w.id" class="blog-work-card">
          <div class="blog-work-cover">
            <img v-if="w.cover" :src="w.cover" :alt="w.name" loading="lazy" />
            <i v-else class="fa-solid fa-diagram-project"></i>
          </div>
          <div class="blog-work-body">
            <div class="blog-work-head">
              <strong class="blog-work-name">{{ w.name }}</strong>
              <span v-if="w.year" class="blog-work-year">{{ w.year }}</span>
            </div>
            <p class="blog-work-desc">{{ w.description || '暂无描述' }}</p>
            <div v-if="parseTechStack(w.tech_stack).length" class="blog-work-techs">
              <span v-for="t in parseTechStack(w.tech_stack)" :key="t" class="blog-work-tech">{{ t }}</span>
            </div>
            <div class="blog-work-actions">
              <a
                v-if="w.demo_url"
                :href="w.demo_url"
                target="_blank"
                rel="noopener noreferrer"
                class="blog-work-btn primary"
              >
                <i class="fa-solid fa-rocket"></i> 演示
              </a>
              <a
                v-if="w.repo_url"
                :href="w.repo_url"
                target="_blank"
                rel="noopener noreferrer"
                class="blog-work-btn"
              >
                <i class="fa-solid fa-code"></i> 源码
              </a>
              <a
                v-if="w.article_url"
                :href="w.article_url"
                target="_blank"
                rel="noopener noreferrer"
                class="blog-work-btn"
              >
                <i class="fa-solid fa-file-lines"></i> 文章
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
