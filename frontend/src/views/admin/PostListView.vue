<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { deletePost, listAdminCategories, listAdminPosts, listAdminTags } from '@/api'
import type { Category, PostListItem, Tag } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { formatDate } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'

const status = ref('')
const keyword = ref('')
const categoryId = ref<number | undefined>(undefined)
const tagId = ref<number | undefined>(undefined)
const isTop = ref<'0' | '1' | ''>('')
const categories = ref<Category[]>([])
const tags = ref<Tag[]>([])
const deletingId = ref<number | null>(null)
const { result: posts, loading, load, changePage, reload } = usePagination<PostListItem>(
  (p) => {
    const params: Record<string, unknown> = { ...p }
    if (status.value) params.status = status.value
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    if (categoryId.value) params.category_id = categoryId.value
    if (tagId.value) params.tag_id = tagId.value
    if (isTop.value !== '') params.is_top = isTop.value
    return listAdminPosts(params)
  },
)

async function loadFilters() {
  try {
    const [categoryRes, tagRes] = await Promise.all([
      listAdminCategories({ page: 1, pageSize: 100 }),
      listAdminTags({ page: 1, pageSize: 100 }),
    ])
    categories.value = categoryRes.list
    tags.value = tagRes.list
  } catch {
    // 筛选选项加载失败不阻塞文章列表
  }
}

async function remove(id: number) {
  try {
    await ElMessageBox.confirm('确定删除该文章？删除后不可恢复。', '提示', { type: 'warning' })
  } catch {
    return
  }
  deletingId.value = id
  try {
    await deletePost(id)
    ElMessage.success('删除成功')
    load()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  } finally {
    deletingId.value = null
  }
}

// 状态/分类/标签/置顶筛选变化时回到第 1 页并重新加载
watch([status, categoryId, tagId, isTop], reload)

onMounted(() => {
  load()
  loadFilters()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h1>文章管理</h1>
      <el-button type="primary" @click="$router.push('/admin/posts/new')">新建文章</el-button>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-input
          v-model="keyword"
          placeholder="搜索标题…"
          clearable
          style="width: 240px"
          @keyup.enter="reload"
        />
        <el-select v-model="status" placeholder="状态" clearable style="width: 130px">
          <el-option label="已发布" value="published" />
          <el-option label="草稿" value="draft" />
        </el-select>
        <el-select v-model="categoryId" placeholder="分类" clearable filterable style="width: 140px">
          <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
        <el-select v-model="tagId" placeholder="标签" clearable filterable style="width: 140px">
          <el-option v-for="t in tags" :key="t.id" :label="t.name" :value="t.id" />
        </el-select>
        <el-select v-model="isTop" placeholder="置顶" clearable style="width: 110px">
          <el-option label="仅置顶" value="1" />
          <el-option label="仅非置顶" value="0" />
        </el-select>
        <el-button type="primary" @click="reload">搜索</el-button>
      </div>

      <el-table v-loading="loading" :data="posts.list">
        <el-table-column label="标题" min-width="280">
          <template #default="{ row }">
            <RouterLink :to="`/post/${row.slug}`" target="_blank">{{ row.title }}</RouterLink>
            <el-tag v-if="row.is_top === 1" type="warning" size="small" style="margin-left: 8px">置顶</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="140">
          <template #default="{ row }">{{ row.category?.name || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="row.status === 'published' ? 'success' : 'info'">
              {{ row.status === 'published' ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="view_count" label="浏览量" width="100" />
        <el-table-column label="发布时间" width="190">
          <template #default="{ row }">{{ formatDate(row.published_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="$router.push(`/admin/posts/${row.id}/edit`)">编辑</el-button>
            <el-button link type="danger" :loading="deletingId === row.id" @click="remove(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="posts.total > posts.pageSize"
        style="margin-top: 16px"
        background
        layout="prev, pager, next, total"
        :total="posts.total"
        :page-size="posts.pageSize"
        :current-page="posts.page"
        @current-change="changePage"
      />
    </el-card>
  </div>
</template>
