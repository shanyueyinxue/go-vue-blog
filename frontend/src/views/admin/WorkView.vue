<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { createWork, deleteWork, listAdminWorks, updateWork, uploadFile } from '@/api'
import type { Work } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { ElMessage, ElMessageBox } from 'element-plus'

const { result, loading, load, changePage } = usePagination<Work>((p) => listAdminWorks(p))
const editing = ref<Work | null>(null)
const form = ref({
  name: '',
  slug: '',
  description: '',
  cover: '',
  demo_url: '',
  article_url: '',
  repo_url: '',
  tech_stack: [] as string[],
  year: '',
  is_top: 0,
  status: 1,
  sort_order: 0,
})
const saving = ref(false)
const errorMsg = ref('')

/** 解析后端返回的技术栈 JSON 字符串数组 */
function parseTechStack(raw: string): string[] {
  if (!raw) return []
  try {
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? arr.filter((v): v is string => typeof v === 'string') : []
  } catch {
    return []
  }
}

function resetForm() {
  form.value = {
    name: '',
    slug: '',
    description: '',
    cover: '',
    demo_url: '',
    article_url: '',
    repo_url: '',
    tech_stack: [],
    year: '',
    is_top: 0,
    status: 1,
    sort_order: 0,
  }
}

function openNew() {
  editing.value = null
  resetForm()
  errorMsg.value = ''
}

function openEdit(w: Work) {
  editing.value = w
  form.value = {
    name: w.name,
    slug: w.slug,
    description: w.description,
    cover: w.cover,
    demo_url: w.demo_url,
    article_url: w.article_url,
    repo_url: w.repo_url,
    tech_stack: parseTechStack(w.tech_stack),
    year: w.year,
    is_top: w.is_top,
    status: w.status,
    sort_order: w.sort_order,
  }
  errorMsg.value = ''
}

async function save() {
  errorMsg.value = ''
  if (!form.value.name.trim()) {
    errorMsg.value = '作品名不能为空'
    return
  }
  saving.value = true
  try {
    const params = {
      name: form.value.name.trim(),
      slug: form.value.slug.trim() || undefined,
      description: form.value.description,
      cover: form.value.cover,
      demo_url: form.value.demo_url,
      article_url: form.value.article_url,
      repo_url: form.value.repo_url,
      tech_stack: form.value.tech_stack,
      year: form.value.year,
      is_top: form.value.is_top,
      status: form.value.status,
      sort_order: form.value.sort_order,
    }
    if (editing.value) {
      await updateWork(editing.value.id, params)
      ElMessage.success('作品已更新')
    } else {
      await createWork(params)
      ElMessage.success('作品已创建')
    }
    openNew()
    await load()
  } catch (e) {
    errorMsg.value = (e as Error).message || '保存失败'
  } finally {
    saving.value = false
  }
}

const coverUploading = ref(false)

async function handleCoverUpload(options: { file: File | Blob }) {
  const file = options.file
  if (!(file instanceof File)) return
  if (!file.type.startsWith('image/')) {
    ElMessage.warning('请选择图片文件')
    return
  }
  coverUploading.value = true
  try {
    const result = await uploadFile(file, 'works')
    form.value.cover = result.url
    ElMessage.success('封面上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    coverUploading.value = false
  }
}

async function remove(w: Work) {
  try {
    await ElMessageBox.confirm(`确定删除作品「${w.name}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteWork(w.id)
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
      <h1>作品管理</h1>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="10">
        <el-card>
          <template #header>{{ editing ? '编辑作品' : '新建作品' }}</template>
          <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon style="margin-bottom: 16px" />
          <el-form label-position="top">
            <el-form-item label="作品名" required>
              <el-input v-model="form.name" placeholder="必填" />
            </el-form-item>
            <el-form-item label="别名 slug">
              <el-input v-model="form.slug" placeholder="留空自动生成" />
            </el-form-item>
            <el-form-item label="封面">
              <el-input v-model="form.cover" placeholder="https://…">
                <template #append>
                  <el-upload :show-file-list="false" accept="image/*" :http-request="handleCoverUpload"
                    :disabled="coverUploading">
                    <el-button :loading="coverUploading" :disabled="coverUploading">上传</el-button>
                  </el-upload>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="form.description" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item label="演示链接">
              <el-input v-model="form.demo_url" placeholder="https://…" />
            </el-form-item>
            <el-form-item label="文章链接">
              <el-input v-model="form.article_url" placeholder="https://…" />
            </el-form-item>
            <el-form-item label="源码链接">
              <el-input v-model="form.repo_url" placeholder="https://…" />
            </el-form-item>
            <el-form-item label="技术栈">
              <el-select
                v-model="form.tech_stack"
                multiple
                filterable
                allow-create
                default-first-option
                placeholder="输入后回车添加，如 Vue 3 / Go"
                style="width: 100%"
              />
            </el-form-item>
            <el-row :gutter="12">
              <el-col :span="12">
                <el-form-item label="年份">
                  <el-input v-model="form.year" placeholder="如 2025" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="排序">
                  <el-input-number v-model="form.sort_order" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
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
                <el-form-item label="置顶推荐">
                  <el-switch v-model="form.is_top" :active-value="1" :inactive-value="0" />
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
          <template #header>作品列表</template>
          <el-table v-loading="loading" :data="result.list">
            <el-table-column label="封面" width="80">
              <template #default="{ row }">
                <el-image
                  v-if="row.cover"
                  :src="row.cover"
                  :preview-src-list="[row.cover]"
                  preview-teleported
                  fit="cover"
                  style="width: 48px; height: 36px; border-radius: 6px"
                />
                <span v-else class="el-icon" style="font-size: 20px"><i class="fa-solid fa-diagram-project"></i></span>
              </template>
            </el-table-column>
            <el-table-column label="名称" min-width="120">
              <template #default="{ row }">
                <span v-if="row.is_top === 1" style="margin-right: 4px">⭐</span>
                {{ row.name }}
              </template>
            </el-table-column>
            <el-table-column label="链接" min-width="180">
              <template #default="{ row }">
                <a
                  v-if="row.demo_url"
                  :href="row.demo_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  style="margin-right: 8px"
                  >演示</a
                >
                <a
                  v-if="row.repo_url"
                  :href="row.repo_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  style="margin-right: 8px"
                  >源码</a
                >
                <a
                  v-if="row.article_url"
                  :href="row.article_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  >文章</a
                >
                <span v-if="!row.demo_url && !row.repo_url && !row.article_url">—</span>
              </template>
            </el-table-column>
            <el-table-column prop="year" label="年份" width="80" />
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
