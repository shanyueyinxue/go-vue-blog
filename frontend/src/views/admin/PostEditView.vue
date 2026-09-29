<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { createPost, createTag, getAdminPost, listCategories, listTags, updatePost, uploadFile } from '@/api'
import type { Category, Tag } from '@/api'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()

const isEdit = ref(false)
const postId = route.params.id ? Number(route.params.id) : null

const form = ref({
  title: '',
  slug: '',
  content: '',
  excerpt: '',
  cover_image: '',
  category_id: undefined as number | undefined,
  tag_ids: [] as (number | string)[],
  status: 'draft' as 'draft' | 'published',
  is_top: 0,
  published_at: '',
})

const categories = ref<Category[]>([])
const tags = ref<Tag[]>([])
const loading = ref(false)
const saving = ref(false)
const errorMsg = ref('')
const successMsg = ref('')

onMounted(async () => {
  isEdit.value = !!postId
  try {
    const [cats, tgs] = await Promise.all([listCategories(), listTags()])
    categories.value = cats
    tags.value = tgs
  } catch {
    // 分类/标签加载失败不阻塞编辑
  }
  if (postId) {
    loading.value = true
    try {
      const post = await getAdminPost(postId)
      form.value.title = post.title
      form.value.slug = post.slug
      form.value.content = post.content
      form.value.excerpt = post.excerpt
      form.value.cover_image = post.cover_image
      form.value.category_id = post.category_id ?? undefined
      form.value.tag_ids = post.tags.map((t) => t.id)
      form.value.status = post.status
      form.value.is_top = post.is_top
      form.value.published_at = post.published_at || ''
    } catch (e) {
      errorMsg.value = (e as Error).message || '加载文章失败'
    } finally {
      loading.value = false
    }
  }
})

/**
 * 解析标签选择器的值：数字为已有标签 ID；字符串为输入的自定义新标签名。
 * 新标签名先尝试复用同名已有标签，否则调用 createTag 创建后返回其 ID。
 */
async function resolveTagIds(raw: (number | string)[]): Promise<number[]> {
  const ids: number[] = []
  for (const item of raw) {
    if (typeof item === 'number') {
      ids.push(item)
      continue
    }
    const name = item.trim()
    if (!name) continue
    // 已存在同名标签则直接复用 ID，避免重复创建
    const existing = tags.value.find((t) => t.name === name)
    if (existing) {
      ids.push(existing.id)
      continue
    }
    try {
      const created = await createTag({ name })
      ids.push(created.id)
    } catch (e) {
      // 单标签创建失败不阻断整体保存，提示后跳过
      ElMessage.warning(`标签「${name}」创建失败：${(e as Error).message || '未知错误'}`)
    }
  }
  return ids
}

async function save() {
  errorMsg.value = ''
  successMsg.value = ''
  if (!form.value.title.trim()) {
    errorMsg.value = '标题不能为空'
    return
  }
  if (!form.value.content.trim()) {
    errorMsg.value = '内容不能为空'
    return
  }
  saving.value = true
  try {
    const tagIds = await resolveTagIds(form.value.tag_ids)
    const params: Record<string, unknown> = {
      title: form.value.title.trim(),
      slug: form.value.slug.trim() || undefined,
      content: form.value.content,
      excerpt: form.value.excerpt.trim() || null,
      cover_image: form.value.cover_image.trim() || null,
      category_id: form.value.category_id ?? null,
      tag_ids: tagIds,
      status: form.value.status,
      is_top: form.value.is_top,
    }
    if (form.value.published_at) {
      const d = new Date(form.value.published_at)
      params.published_at = d.toISOString()
    }
    if (isEdit.value && postId) {
      await updatePost(postId, params)
      ElMessage.success('文章已保存')
      successMsg.value = '保存成功'
    } else {
      const post = await createPost(params)
      ElMessage.success('文章已创建')
      router.replace(`/admin/posts/${post.id}/edit`)
      isEdit.value = true
      successMsg.value = '创建成功'
    }
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
    const result = await uploadFile(file)
    form.value.cover_image = result.url
    ElMessage.success('封面上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    coverUploading.value = false
  }
}

/**
 * Markdown 编辑器图片上传：接入后台文件上传接口（POST /api/admin/files，需 JWT）。
 * md-editor-v3 的 onUploadImg 约定：files 为待上传文件，callback 接收可访问 URL 数组。
 */
async function handleEditorUpload(files: File[], callback: (urls: string[]) => void) {
  try {
    const urls = await Promise.all(
      files.map(async (file) => {
        const result = await uploadFile(file, 'posts')
        return result.url
      }),
    )
    callback(urls)
  } catch (e) {
    ElMessage.error((e as Error).message || '图片上传失败')
    callback([])
  }
}
</script>

<template>
  <div>
    <div class="page-header">
      <h1>{{ isEdit ? '编辑文章' : '新建文章' }}</h1>
      <el-button @click="$router.push('/admin/posts')">返回列表</el-button>
    </div>

    <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon style="margin-bottom: 16px" />
    <el-alert v-if="successMsg" :title="successMsg" type="success" show-icon style="margin-bottom: 16px" />

    <el-skeleton v-if="loading" :rows="8" animated />
    <el-card v-else>
      <el-form label-position="top">
        <el-form-item label="标题" required>
          <el-input v-model="form.title" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :xs="24" :md="12">
            <el-form-item label="别名 slug">
              <el-input v-model="form.slug" placeholder="留空自动生成" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="封面图 URL">
              <el-input v-model="form.cover_image" placeholder="https://…">
                <template #append>
                  <el-upload :show-file-list="false" accept="image/*" :http-request="handleCoverUpload"
                    :disabled="coverUploading">
                    <el-button :loading="coverUploading" :disabled="coverUploading">上传</el-button>
                  </el-upload>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="6">
            <el-form-item label="分类">
              <el-select v-model="form.category_id" clearable placeholder="无分类" style="width: 100%">
                <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="6">
            <el-form-item label="发布时间">
              <el-date-picker v-model="form.published_at" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss"
                placeholder="选择发布时间" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="6">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="草稿" value="draft" />
                <el-option label="发布" value="published" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="6">
            <el-form-item label="置顶">
              <el-select v-model="form.is_top" style="width: 100%">
                <el-option label="否" :value="0" />
                <el-option label="是" :value="1" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24">
            <el-form-item label="标签（可多选 / 输入新标签名回车创建）">
              <el-select v-model="form.tag_ids" multiple filterable allow-create default-first-option
                :reserve-keyword="false" placeholder="选择已有标签，或输入新标签名后回车" style="width: 100%">
                <el-option v-for="t in tags" :key="t.id" :label="t.name" :value="t.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="摘要">
          <el-input v-model="form.excerpt" type="textarea" :rows="3" />
        </el-form-item>

        <el-form-item label="正文（Markdown）" required>
          <MdEditor v-model="form.content" :preview="true" :on-upload-img="handleEditorUpload" />
        </el-form-item>

        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </el-form>
    </el-card>
  </div>
</template>
