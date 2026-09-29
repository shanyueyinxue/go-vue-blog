<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { deleteFile, listFiles, uploadFile } from '@/api'
import type { FileItem } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { formatBytes } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'

const keyword = ref('')
const mimeType = ref('')
const { result, loading, load, changePage, reload } = usePagination<FileItem>(
  (p) => {
    const params: Record<string, unknown> = { ...p }
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    if (mimeType.value) params.mime_type = mimeType.value
    return listFiles(params)
  },
  20,
)
const uploading = ref(false)

async function handleUpload(options: { file: File | Blob }) {
  const file = options.file
  if (!(file instanceof File)) return
  // 上传成功后回到第 1 页查看最新文件
  if (result.value.page !== 1) {
    result.value.page = 1
  }
  uploading.value = true
  try {
    await uploadFile(file)
    ElMessage.success('上传成功')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    uploading.value = false
  }
}

/** 优先使用后端返回的完整 URL，兼容旧数据回退到本地 /uploads 前缀 */
function fileUrl(file: FileItem): string {
  if (file.url) return file.url
  if (/^https?:\/\//i.test(file.path)) return file.path
  return `/uploads/${file.path.replace(/^\/+/, '')}`
}

function thumbnailUrl(file: FileItem): string {
  const res = (() => {
    if (!file.thumbnail) return ''
    if (file.thumbnail_url) return file.thumbnail_url
    if (/^https?:\/\//i.test(file.thumbnail)) return file.thumbnail
    return `/uploads/${file.thumbnail.replace(/^\/+/, '')}`
  })()
  return res
}

/* ------------------------------ 预览（内联弹层，替代新窗口打开） ------------------------------ */

// 图片：全屏 el-image-viewer（与缩略图点开效果一致，支持当前页多图左右切换）
const imageViewerVisible = ref(false)
const imageViewerList = ref<string[]>([])
const imageViewerIndex = ref(0)

function previewImage(file: FileItem) {
  const list = result.value.list
    .filter((f) => f.mime_type.startsWith('image/'))
    .map((f) => fileUrl(f))
  if (list.length === 0) return
  imageViewerList.value = list
  imageViewerIndex.value = Math.max(list.indexOf(fileUrl(file)), 0)
  imageViewerVisible.value = true
}

// 视频 / 音频 / 浏览器可渲染文档（PDF、文本、JSON、XML 等）：el-dialog 内联预览
const mediaDialogVisible = ref(false)
const mediaFile = ref<FileItem | null>(null)

/** 预览内容类型：video / audio / embed（iframe 内嵌） */
const mediaKind = computed<'video' | 'audio' | 'embed'>(() => {
  const mime = mediaFile.value?.mime_type || ''
  if (mime.startsWith('video/')) return 'video'
  if (mime.startsWith('audio/')) return 'audio'
  return 'embed'
})

function preview(file: FileItem) {
  const mime = file.mime_type || ''
  if (mime.startsWith('image/')) {
    previewImage(file)
  } else if (
    mime.startsWith('video/') ||
    mime.startsWith('audio/') ||
    mime === 'application/pdf' ||
    mime === 'application/json' ||
    mime === 'application/xml' ||
    mime.startsWith('text/')
  ) {
    mediaFile.value = file
    mediaDialogVisible.value = true
  } else {
    // 浏览器无法内联渲染的类型（Office 文档、压缩包等）：维持新窗口打开
    window.open(fileUrl(file), '_blank', 'noopener,noreferrer')
  }
}

/**
 * 下载文件。
 * 优先：浏览器直接 fetch 文件地址（同源 /uploads 或 OSS/CDN 直连，不经服务器中转）
 *       → blob 下载，可自定义文件名（跨域需存储域名开启 CORS）。
 * 回退：fetch 失败（存储域名未开 CORS / 网络异常）时新窗口打开，
 *       若 OSS 配了 Content-Disposition 也会触发下载。
 */
async function download(file: FileItem) {
  const url = fileUrl(file)
  try {
    const resp = await fetch(url)
    if (!resp.ok) {
      throw new Error(`HTTP ${resp.status}`)
    }
    const blob = await resp.blob()
    const objUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = objUrl
    a.download = file.filename
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(objUrl)
  } catch {
    window.open(url, '_blank', 'noopener,noreferrer')
  }
}

async function copyLink(file: FileItem) {
  const text = fileUrl(file);
  try {
    // 尝试使用 Clipboard API
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text);
      ElMessage.success('链接已复制');
      return;
    }
    // 如果不支持，回退到 execCommand
    fallbackCopy(text);
    ElMessage.success('链接已复制');
  } catch {
    // 如果 Clipboard API 失败，尝试回退
    try {
      fallbackCopy(text);
      ElMessage.success('链接已复制');
    } catch {
      // 回退也失败，提示手动复制
      ElMessage.error('复制失败，请手动复制：' + text);
    }
  }
}

function fallbackCopy(text: string) {
  const textArea = document.createElement('textarea');
  textArea.value = text;
  // 使 textarea 不可见
  textArea.style.position = 'fixed';
  textArea.style.left = '-9999px';
  textArea.style.top = '-9999px';
  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();
  try {
    const successful = document.execCommand('copy');
    if (!successful) {
      throw new Error('execCommand failed');
    }
  } finally {
    document.body.removeChild(textArea);
  }
}
async function remove(file: FileItem) {
  try {
    await ElMessageBox.confirm(`确定删除文件「${file.filename}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteFile(file.id)
    ElMessage.success('删除成功')
    // 删除后若当前页空了则回退一页
    if (result.value.list.length === 1 && result.value.page > 1) {
      result.value.page -= 1
    }
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
      <h1>文件管理</h1>
      <el-upload :show-file-list="false" :http-request="handleUpload" :disabled="uploading">
        <el-button type="primary" :loading="uploading" :disabled="uploading">上传文件</el-button>
      </el-upload>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-input v-model="keyword" placeholder="搜索文件名…" clearable style="width: 240px" @keyup.enter="reload"
          @clear="reload" />
        <el-select v-model="mimeType" placeholder="MIME 类型" clearable style="width: 160px" @change="reload">
          <el-option label="图片 (image/*)" value="image/" />
          <el-option label="文档 (application/*)" value="application/" />
          <el-option label="文本 (text/*)" value="text/" />
          <el-option label="音频 (audio/*)" value="audio/" />
          <el-option label="视频 (video/*)" value="video/" />
        </el-select>
        <el-button type="primary" @click="reload">搜索</el-button>
      </div>

      <el-table v-loading="loading" :data="result.list">
        <el-table-column label="缩略图" width="90">
          <template #default="{ row }">
            <el-image v-if="row.thumbnail" :src="thumbnailUrl(row)" :preview-src-list="[thumbnailUrl(row)]"
              :z-index="2333" :preview-teleported="true" crossorigin="anonymous" fit="cover"
              style="width: 56px; height: 56px; border-radius: 8px; z-index: 1;" />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="filename" label="文件名" min-width="220" show-overflow-tooltip />
        <el-table-column prop="mime_type" label="MIME 类型" width="180" show-overflow-tooltip />
        <el-table-column label="大小" width="120">
          <template #default="{ row }">{{ formatBytes(row.size) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="preview(row)">查看</el-button>
            <el-button link type="primary" @click="download(row)">下载</el-button>
            <el-button link type="primary" @click="copyLink(row)">复制链接</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination v-if="result.total > result.pageSize" style="margin-top: 16px; justify-content: flex-end"
        background layout="prev, pager, next, total" :total="result.total" :page-size="result.pageSize"
        :current-page="result.page" @current-change="changePage" />
    </el-card>

    <!-- 图片全屏预览（与缩略图一致的 el-image-viewer 效果） -->
    <el-image-viewer
      v-if="imageViewerVisible"
      :url-list="imageViewerList"
      :initial-index="imageViewerIndex"
      :z-index="3000"
      teleported
      hide-on-click-modal
      @close="imageViewerVisible = false"
    />

    <!-- 视频 / 音频 / 文档（PDF、文本等）内联预览 -->
    <el-dialog
      v-model="mediaDialogVisible"
      :title="mediaFile?.filename || '预览'"
      width="min(960px, 92vw)"
      top="5vh"
      destroy-on-close
      @closed="mediaFile = null"
    >
      <video
        v-if="mediaKind === 'video' && mediaFile"
        :src="fileUrl(mediaFile)"
        controls
        autoplay
        style="width: 100%; max-height: 75vh; border-radius: 8px; background: #000"
      />
      <audio
        v-else-if="mediaKind === 'audio' && mediaFile"
        :src="fileUrl(mediaFile)"
        controls
        autoplay
        style="width: 100%"
      />
      <iframe
        v-else-if="mediaFile"
        :src="fileUrl(mediaFile)"
        style="width: 100%; height: 72vh; border: 0; border-radius: 8px"
      />
    </el-dialog>
  </div>
</template>
