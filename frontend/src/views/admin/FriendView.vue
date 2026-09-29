<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { createFriend, deleteFriend, listAdminFriends, updateFriend, uploadFile } from '@/api'
import type { Friend } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { ElMessage, ElMessageBox } from 'element-plus'

const { result, loading, load, changePage } = usePagination<Friend>((p) => listAdminFriends(p))
const editing = ref<Friend | null>(null)
const form = ref({ name: '', url: '', icon: '', description: '', status: 1, sort_order: 0 })
const saving = ref(false)
const errorMsg = ref('')

function openNew() {
  editing.value = null
  form.value = { name: '', url: '', icon: '', description: '', status: 1, sort_order: 0 }
  errorMsg.value = ''
}

function openEdit(f: Friend) {
  editing.value = f
  form.value = {
    name: f.name,
    url: f.url,
    icon: f.icon,
    description: f.description,
    status: f.status,
    sort_order: f.sort_order,
  }
  errorMsg.value = ''
}

async function save() {
  errorMsg.value = ''
  if (!form.value.name.trim() || !form.value.url.trim()) {
    errorMsg.value = '名称和地址不能为空'
    return
  }
  saving.value = true
  try {
    const params = {
      name: form.value.name.trim(),
      url: form.value.url.trim(),
      icon: form.value.icon.trim(),
      description: form.value.description,
      status: form.value.status,
      sort_order: form.value.sort_order,
    }
    if (editing.value) {
      await updateFriend(editing.value.id, params)
      ElMessage.success('友链已更新')
    } else {
      await createFriend(params)
      ElMessage.success('友链已创建')
    }
    openNew()
    await load()
  } catch (e) {
    errorMsg.value = (e as Error).message || '保存失败'
  } finally {
    saving.value = false
  }
}

const iconUploading = ref(false)

async function handleIconUpload(options: { file: File | Blob }) {
  const file = options.file
  if (!(file instanceof File)) return
  if (!file.type.startsWith('image/')) {
    ElMessage.warning('请选择图片文件')
    return
  }
  iconUploading.value = true
  try {
    const result = await uploadFile(file)
    form.value.icon = result.url
    ElMessage.success('图标上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    iconUploading.value = false
  }
}

async function remove(f: Friend) {
  try {
    await ElMessageBox.confirm(`确定删除友链「${f.name}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteFriend(f.id)
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
      <h1>友链管理</h1>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="10">
        <el-card>
          <template #header>{{ editing ? '编辑友链' : '新建友链' }}</template>
          <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon style="margin-bottom: 16px" />
          <el-form label-position="top">
            <el-form-item label="名称" required>
              <el-input v-model="form.name" />
            </el-form-item>
            <el-form-item label="地址" required>
              <el-input v-model="form.url" placeholder="https://…" />
            </el-form-item>
            <el-form-item label="图标 URL">
              <el-input v-model="form.icon" placeholder="https://…">
                <template #append>
                  <el-upload :show-file-list="false" accept="image/*" :http-request="handleIconUpload"
                    :disabled="iconUploading">
                    <el-button :loading="iconUploading" :disabled="iconUploading">上传</el-button>
                  </el-upload>
                </template>
              </el-input>
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
          <template #header>友链列表</template>
          <el-table v-loading="loading" :data="result.list">
            <el-table-column prop="name" label="名称" min-width="120" />
            <el-table-column label="地址" min-width="220">
              <template #default="{ row }">
                <a :href="row.url" target="_blank" rel="noopener noreferrer">{{ row.url }}</a>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="sort_order" label="排序" width="80" />
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
