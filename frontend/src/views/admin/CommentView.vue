<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { deleteComment, listAdminComments, updateComment } from '@/api'
import type { AdminComment } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { formatDate } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'

const status = ref('')
const keyword = ref('')
const detailVisible = ref(false)
const detail = ref<AdminComment | null>(null)
const saving = ref(false)
const { result, loading, load, changePage, reload } = usePagination<AdminComment>(
  (p) => {
    const params: Record<string, unknown> = { ...p }
    if (status.value) params.status = status.value
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    return listAdminComments(params)
  },
  10,
)

/** 保存详情中的修改（仅 status / is_author 两个字段，其余字段后端一律忽略） */
async function saveDetail() {
  if (!detail.value) return
  saving.value = true
  try {
    await updateComment(detail.value.id, {
      status: detail.value.status,
      is_author: detail.value.is_author,
    })
    ElMessage.success('评论已更新')
    detailVisible.value = false
    load()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function remove(c: AdminComment) {
  try {
    await ElMessageBox.confirm('确定删除该评论？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteComment(c.id)
    ElMessage.success('删除成功')
    load()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

function showDetail(c: AdminComment) {
  detail.value = c
  detailVisible.value = true
}

// 状态筛选变化时回到第 1 页并重新加载
watch(status, reload)

onMounted(load)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>评论管理</h1>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-input
          v-model="keyword"
          placeholder="搜索昵称 / 内容…"
          clearable
          style="width: 320px"
          @keyup.enter="reload"
        />
        <el-select v-model="status" style="width: 160px" @change="reload">
          <el-option label="全部状态" value="" />
          <el-option label="待审核" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
        <el-button type="primary" @click="reload">搜索</el-button>
      </div>

      <el-table v-loading="loading" :data="result.list">
        <el-table-column label="昵称" width="130">
          <template #default="{ row }">
            {{ row.nickname }}
            <el-tag v-if="row.is_author === 1" size="small" style="margin-left: 6px">博主</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="内容" min-width="220" show-overflow-tooltip />
        <el-table-column label="邮箱 / 电话" width="170">
          <template #default="{ row }">{{ row.email || row.phone || '-' }}</template>
        </el-table-column>
        <el-table-column label="IP" width="130">
          <template #default="{ row }">{{ row.ip || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'approved' ? 'success' : row.status === 'pending' ? 'warning' : 'danger'">
              {{ row.status === 'approved' ? '已通过' : row.status === 'pending' ? '待审核' : '已拒绝' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="190">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">详情</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="result.total > result.pageSize"
        style="margin-top: 16px"
        background
        layout="prev, pager, next, total"
        :total="result.total"
        :page-size="result.pageSize"
        :current-page="result.page"
        @current-change="changePage"
      />
    </el-card>

    <el-dialog v-model="detailVisible" title="评论详情" width="640px">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="昵称">
          {{ detail.nickname }}
          <el-tag v-if="detail.is_author === 1" size="small" style="margin-left: 6px">博主</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="关联文章">
          {{ detail.post_title || `文章 #${detail.post_id}` }}
        </el-descriptions-item>
        <el-descriptions-item label="父级评论">
          {{ detail.parent_nickname || '无' }}
        </el-descriptions-item>
        <el-descriptions-item label="邮箱 / 电话">
          {{ detail.email || detail.phone || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="IP">
          {{ detail.ip || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="评论内容">
          <div style="white-space: pre-wrap">{{ detail.content }}</div>
        </el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">修改</el-divider>
      <el-form v-if="detail" label-width="90px">
        <el-form-item label="审核状态">
          <el-select v-model="detail.status" style="width: 220px">
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
          <p class="form-hint">改为「已通过」时自动通知评论者并同步文章评论数</p>
        </el-form-item>
        <el-form-item label="博主评论">
          <el-switch v-model="detail.is_author" :active-value="1" :inactive-value="0" />
          <p class="form-hint">标记为博主评论后前台展示博主标识</p>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="detailVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveDetail">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
