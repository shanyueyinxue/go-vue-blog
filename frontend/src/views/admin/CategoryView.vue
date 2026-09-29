<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { createCategory, deleteCategory, listAdminCategories, updateCategory } from '@/api'
import type { Category } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { ElMessage, ElMessageBox } from 'element-plus'

const { result, loading, load, changePage } = usePagination<Category>((p) => listAdminCategories(p))
const editing = ref<Category | null>(null)
const form = ref({ name: '', slug: '', description: '', status: 1, sort_order: 0 })
const saving = ref(false)
const errorMsg = ref('')

function openNew() {
  editing.value = null
  form.value = { name: '', slug: '', description: '', status: 1, sort_order: 0 }
  errorMsg.value = ''
}

function openEdit(c: Category) {
  editing.value = c
  form.value = {
    name: c.name,
    slug: c.slug,
    description: c.description,
    status: c.status,
    sort_order: c.sort_order,
  }
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
      description: form.value.description,
      status: form.value.status,
      sort_order: form.value.sort_order,
    }
    if (editing.value) {
      await updateCategory(editing.value.id, params)
      ElMessage.success('分类已更新')
    } else {
      await createCategory(params)
      ElMessage.success('分类已创建')
    }
    openNew()
    await load()
  } catch (e) {
    errorMsg.value = (e as Error).message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function remove(c: Category) {
  try {
    await ElMessageBox.confirm(`确定删除分类「${c.name}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteCategory(c.id)
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
      <h1>分类管理</h1>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="10">
        <el-card>
          <template #header>{{ editing ? '编辑分类' : '新建分类' }}</template>
          <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon style="margin-bottom: 16px" />
          <el-form label-position="top">
            <el-form-item label="名称" required>
              <el-input v-model="form.name" />
            </el-form-item>
            <el-form-item label="别名 slug">
              <el-input v-model="form.slug" placeholder="留空自动生成" />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="form.description" type="textarea" />
            </el-form-item>
            <el-row :gutter="12">
              <el-col :span="12">
                <el-form-item label="状态">
                  <el-select v-model="form.status" style="width: 100%">
                    <el-option label="启用" :value="1" />
                    <el-option label="禁用" :value="0" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="排序">
                  <el-input-number v-model="form.sort_order" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-button type="primary" :loading="saving" @click="save">保存</el-button>
            <el-button v-if="editing" @click="openNew">取消</el-button>
          </el-form>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="14">
        <el-card>
          <template #header>分类列表</template>
          <el-table v-loading="loading" :data="result.list">
            <el-table-column prop="name" label="名称" min-width="140" />
            <el-table-column prop="slug" label="slug" min-width="130" />
            <el-table-column prop="post_count" label="文章数" width="90" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
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
