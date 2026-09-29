<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getArchive } from '@/api'
import type { ArchiveItem } from '@/api'
import { markPageReady } from '@/composables/usePageReady'
import { formatDate } from '@/utils/format'

const items = ref<ArchiveItem[]>([])
const loading = ref(false)

interface YearGroup {
  year: number
  /** 该年文章总数 */
  total: number
  /** 该年各月（降序） */
  months: ArchiveItem[]
}

// 后端返回的是「月」级别的扁平列表；按年份聚合为 年 > 月 > 文章列表 三层结构，
// 避免年份在每个月上重复出现
const yearGroups = computed<YearGroup[]>(() => {
  const groups: YearGroup[] = []
  const index = new Map<number, YearGroup>()
  for (const item of items.value) {
    let group = index.get(item.year)
    if (!group) {
      group = { year: item.year, total: 0, months: [] }
      index.set(item.year, group)
      groups.push(group)
    }
    group.total += item.count
    group.months.push(item)
  }
  return groups
})

onMounted(async () => {
  loading.value = true
  try {
    items.value = await getArchive()
  } finally {
    loading.value = false
    markPageReady()
  }
})
</script>

<template>
  <div>
    <h1 class="blog-section-title">文章归档</h1>
    <div class="blog-archive-list">
      <div v-if="loading" class="blog-empty">
        <i class="fa-solid fa-spinner fa-spin"></i>
        <span style="margin-left: 8px">加载中…</span>
      </div>
      <div v-else-if="items.length === 0" class="blog-empty">暂无文章</div>
      <template v-else>
        <div v-for="group in yearGroups" :key="group.year" class="blog-archive-card">
          <h2 class="blog-archive-year">
            {{ group.year }} 年
            <span class="blog-archive-count">{{ group.total }} 篇</span>
          </h2>
          <div v-for="item in group.months" :key="item.month">
            <div class="blog-archive-month">
              <i class="fa-solid fa-calendar fa-fw"></i>
              {{ item.month }} 月 · {{ item.count }} 篇
            </div>
            <div v-for="p in item.posts" :key="p.id" class="blog-archive-post">
              <RouterLink :to="`/post/${p.slug}`">{{ p.title }}</RouterLink>
              <time>{{ formatDate(p.published_at) }}</time>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
