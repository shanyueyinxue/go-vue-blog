<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getStatsSummary } from '@/api'
import type { StatsSummary } from '@/api'
import { formatDate } from '@/utils/format'

const stats = ref<StatsSummary | null>(null)
const loading = ref(false)

const statCards = computed(() => {
  const s = stats.value
  if (!s) return []
  return [
    { label: '文章总数', value: s.post_count },
    { label: '已发布', value: s.published_count },
    { label: '草稿', value: s.draft_count },
    { label: '分类', value: s.category_count },
    { label: '标签', value: s.tag_count },
    { label: '评论', value: s.comment_count },
    { label: '待审核', value: s.pending_comment_count },
    { label: '总浏览', value: s.total_views },
  ]
})

onMounted(async () => {
  loading.value = true
  try {
    stats.value = await getStatsSummary()
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="page-header">
      <h1>仪表盘</h1>
    </div>
    <el-skeleton v-if="loading" :rows="8" animated />
    <template v-else-if="stats">
      <el-row :gutter="16">
        <el-col v-for="card in statCards" :key="card.label" :xs="12" :sm="8" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-card-value">{{ card.value }}</div>
            <div class="stat-card-label">{{ card.label }}</div>
          </el-card>
        </el-col>
      </el-row>

      <el-card style="margin-top: 20px">
        <template #header>最近文章</template>
        <el-empty v-if="stats.recent_posts.length === 0" description="暂无文章" />
        <el-table v-else :data="stats.recent_posts">
          <el-table-column label="标题" min-width="280">
            <template #default="{ row }">
              <RouterLink :to="`/admin/posts/${row.id}/edit`">{{ row.title }}</RouterLink>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="row.status === 'published' ? 'success' : 'info'">
                {{ row.status === 'published' ? '已发布' : '草稿' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="发布时间" width="190">
            <template #default="{ row }">{{ formatDate(row.published_at) }}</template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>
  </div>
</template>
