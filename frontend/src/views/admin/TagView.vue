<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { createTag, deleteTag, listAdminTags, updateTag } from '@/api'
import type { Tag } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { ElMessage, ElMessageBox } from 'element-plus'

const { result, loading, load, changePage } = usePagination<Tag>((p) => listAdminTags(p))
const editing = ref<Tag | null>(null)
const form = ref({ name: '', slug: '' })
const saving = ref(false)
const errorMsg = ref('')

function openNew() {
  editing.value = null
  form.value = { name: '', slug: '' }
  errorMsg.value = ''
}

function openEdit(t: Tag) {
  editing.value = t
  form.value = { name: t.name, slug: t.slug }
  errorMsg.value = ''
}

async function save() {
  errorMsg.value = ''
  if (!form.value.name.trim()) {
    errorMsg.value = '名称不能为空'
    return
  }
  saving.value = true
  try {
    const params = {
      name: form.value.name.trim(),
      slug: form.value.slug.trim() || undefined,
    }
    if (editing.value) {
      await updateTag(editing.value.id, params)
      ElMessage.success('标签已更新')
    } else {
      await createTag(params)
      ElMessage.success('标签已创建')
    }
    openNew()
    await load()
  } catch (e) {
    errorMsg.value = (e as Error).message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function remove(t: Tag) {
  try {
    await ElMessageBox.confirm(`确定删除标签「${t.name}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteTag(t.id)
    ElMessage.success('删除成功')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>标签管理</h1>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="10">
        <el-card>
          <template #header>{{ editing ? '编辑标签' : '新建标签' }}</template>
          <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon style="margin-bottom: 16px" />
          <el-form label-position="top">
            <el-form-item label="名称" required>
              <el-input v-model="form.name" />
            </el-form-item>
            <el-form-item label="别名 slug">
              <el-input v-model="form.slug" placeholder="留空自动生成" />
            </el-form-item>
            <el-button type="primary" :loading="saving" @click="save">保存</el-button>
            <el-button v-if="editing" @click="openNew">取消</el-button>
          </el-form>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="14">
        <el-card>
          <template #header>标签列表</template>
          <el-table v-loading="loading" :data="result.list">
            <el-table-column prop="name" label="名称" min-width="140" />
            <el-table-column prop="slug" label="slug" min-width="130" />
            <el-table-column prop="post_count" label="文章数" width="90" />
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
                <el-button link type="danger" @click="remove(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-if="result.total > result.pageSize"
            style="margin-top: 16px; justify-content: flex-end"
            background
            layout="prev, pager, next, total"
            :total="result.total"
            :page-size="result.pageSize"
            :current-page="result.page"
            @current-change="changePage"
          />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
