<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { createMusic, deleteMusic, listMusics, updateMusic, uploadFile } from '@/api'
import type { Music } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { ElMessage, ElMessageBox } from 'element-plus'

const { result, loading, load, changePage } = usePagination<Music>((p) => listMusics(p))
const editing = ref<Music | null>(null)
const form = ref({
  name: '',
  artist: '',
  cover: '',
  file_id: 0,
  url: '', // 当前音频 URL（仅预览用，不入参）
  sort_order: 0,
  status: 1,
})
const saving = ref(false)
const errorMsg = ref('')

function resetForm() {
  form.value = {
    name: '',
    artist: '',
    cover: '',
    file_id: 0,
    url: '',
    sort_order: 0,
    status: 1,
  }
}

function openNew() {
  editing.value = null
  resetForm()
  errorMsg.value = ''
}

function openEdit(m: Music) {
  editing.value = m
  form.value = {
    name: m.name,
    artist: m.artist,
    cover: m.cover,
    file_id: m.file_id,
    url: m.url || '',
    sort_order: m.sort_order,
    status: m.status,
  }
  errorMsg.value = ''
}

async function save() {
  errorMsg.value = ''
  if (!form.value.name.trim()) {
    errorMsg.value = '歌名不能为空'
    return
  }
  if (!form.value.file_id) {
    errorMsg.value = '请先上传音频文件'
    return
  }
  saving.value = true
  try {
    const params = {
      name: form.value.name.trim(),
      artist: form.value.artist,
      cover: form.value.cover,
      file_id: form.value.file_id,
      sort_order: form.value.sort_order,
      status: form.value.status,
    }
    if (editing.value) {
      await updateMusic(editing.value.id, params)
      ElMessage.success('音乐已更新')
    } else {
      await createMusic(params)
      ElMessage.success('音乐已创建')
    }
    openNew()
    await load()
  } catch (e) {
    errorMsg.value = (e as Error).message || '保存失败'
  } finally {
    saving.value = false
  }
}

const audioUploading = ref(false)

async function handleAudioUpload(options: { file: File | Blob }) {
  const file = options.file
  if (!(file instanceof File)) return
  // m4a（MP4 容器）浏览器可能识别为 video/mp4，一并放行；最终以后端校验为准
  if (!file.type.startsWith('audio/') && file.type !== 'video/mp4') {
    ElMessage.warning('请选择音频文件')
    return
  }
  audioUploading.value = true
  try {
    const result = await uploadFile(file, 'audios')
    form.value.file_id = result.id
    form.value.url = result.url
    ElMessage.success('音频上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    audioUploading.value = false
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
    const result = await uploadFile(file, 'images')
    form.value.cover = result.url
    ElMessage.success('封面上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    coverUploading.value = false
  }
}

async function remove(m: Music) {
  try {
    await ElMessageBox.confirm(`确定删除音乐「${m.name}」？将同步删除关联的音频文件。`, '提示', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await deleteMusic(m.id)
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
      <h1>音乐管理</h1>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="10">
        <el-card>
          <template #header>{{ editing ? '编辑音乐' : '新建音乐' }}</template>
          <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon style="margin-bottom: 16px" />
          <el-form label-position="top">
            <el-form-item label="歌名" required>
              <el-input v-model="form.name" placeholder="必填" />
            </el-form-item>
            <el-form-item label="歌手">
              <el-input v-model="form.artist" placeholder="如 周杰伦" />
            </el-form-item>
            <el-form-item label="音频文件" required>
              <div style="display: flex; gap: 8px; align-items: flex-start; width: 100%">
                <div style="flex: 1; min-width: 0">
                  <audio v-if="form.url" :src="form.url" controls preload="none"
                    style="width: 100%; height: 36px" />
                  <el-input v-else :model-value="form.file_id ? `已上传文件 #${form.file_id}` : '未上传音频'"
                    disabled />
                </div>
                <el-upload :show-file-list="false" accept="audio/*,.mp3,.ogg,.wav,.flac,.m4a,.aac,.opus,.webm"
                  :http-request="handleAudioUpload" :disabled="audioUploading">
                  <el-button type="primary" :loading="audioUploading" :disabled="audioUploading">
                    {{ form.file_id ? '重新上传' : '上传' }}
                  </el-button>
                </el-upload>
              </div>
              <p class="form-hint">音频上传至文件库（audios 目录），替换音频会同步清理旧文件</p>
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
            <el-row :gutter="12">
              <el-col :span="12">
                <el-form-item label="排序">
                  <el-input-number v-model="form.sort_order" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="状态">
                  <el-select v-model="form.status" style="width: 100%">
                    <el-option label="启用" :value="1" />
                    <el-option label="禁用" :value="0" />
                  </el-select>
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
          <template #header>音乐列表</template>
          <el-table v-loading="loading" :data="result.list">
            <el-table-column label="封面" width="70">
              <template #default="{ row }">
                <el-image
                  v-if="row.cover"
                  :src="row.cover"
                  :preview-src-list="[row.cover]"
                  preview-teleported
                  fit="cover"
                  style="width: 44px; height: 44px; border-radius: 6px"
                />
                <span v-else class="el-icon" style="font-size: 20px"><i class="fa-solid fa-music"></i></span>
              </template>
            </el-table-column>
            <el-table-column label="歌名 / 歌手" min-width="160">
              <template #default="{ row }">
                <div>{{ row.name }}</div>
                <div class="text-muted" style="font-size: 12px">{{ row.artist || '—' }}</div>
              </template>
            </el-table-column>
            <el-table-column label="试听" min-width="180">
              <template #default="{ row }">
                <audio v-if="row.url" :src="row.url" controls preload="none" style="width: 100%; height: 30px" />
                <span v-else class="text-muted">—</span>
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
