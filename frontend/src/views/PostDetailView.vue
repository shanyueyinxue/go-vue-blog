<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useHead } from '@vueuse/head'
import MdRenderer from '@/components/MdRenderer.vue'
import CommentItem from '@/components/CommentItem.vue'
import { createComment, getPostBySlug, increasePostView, listComments, sendEmailCaptcha } from '@/api'
import type { Comment, PageResult, Post } from '@/api'
import { useSiteStore } from '@/stores/site'
import { usePostCover } from '@/composables/usePostCover'
import { markPageReady } from '@/composables/usePageReady'
import { formatDate } from '@/utils/format'

const route = useRoute()
const slug = computed(() => String(route.params.slug ?? ''))

const siteStore = useSiteStore()
const postCover = usePostCover()

const post = ref<Post | null>(null)
const comments = ref<PageResult<Comment>>({ list: [], total: 0, page: 1, pageSize: 50 })
const loading = ref(false)
const notFound = ref(false)
const commentFormRef = ref<HTMLElement | null>(null)

// 依据后端返回的文章数据动态改写 <head>：标题 = 文章标题 + 站点名，
// 描述用文章摘要 excerpt，关键词取文章 tags；未加载/不存在时回退路由 meta。
// 离开本页（组件卸载）后 useHead 自动回退到 App.vue 的全局设置
useHead({
  title: computed(() => {
    const suffix = siteStore.title
    const page = route.meta.title
    if (post.value) return `${post.value.title} - ${suffix}`
    return page ? `${page} - ${suffix}` : suffix
  }),
  meta: computed(() => {
    const entries: Array<{ name: string; content: string }> = []
    if (notFound.value) {
      entries.push({ name: 'description', content: '文章不存在或未发布。' })
      return entries
    }
    const excerpt = (post.value?.excerpt ?? '').trim()
    if (post.value && excerpt) {
      entries.push({ name: 'description', content: excerpt.length > 160 ? `${excerpt.slice(0, 160)}…` : excerpt })
    } else {
      entries.push({ name: 'description', content: route.meta.description || siteStore.description })
    }
    const keywords = (post.value?.tags ?? []).map((t) => t.name).join(', ')
    if (post.value && keywords) {
      entries.push({ name: 'keywords', content: keywords })
    }
    return entries
  }),
})

const form = ref({
  nickname: '',
  email: '',
  content: '',
  parent_id: 0,
  captcha_token: '',
  captcha_code: '',
})
const submitting = ref(false)
const submitMsg = ref<{ type: 'success' | 'error'; text: string } | null>(null)
const captchaCountdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

/** 站点是否开启评论邮箱验证码（配置未加载时按开启处理，与后端默认一致） */
const needCaptcha = computed(() => siteStore.config.comment_need_captcha !== 0)

function clearCountdown() {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
  captchaCountdown.value = 0
}

async function sendCaptcha() {
  submitMsg.value = null
  if (!form.value.email.trim()) {
    submitMsg.value = { type: 'error', text: '请先填写邮箱' }
    return
  }
  try {
    const res = await sendEmailCaptcha({ email: form.value.email.trim(), scene: 'comment' })
    form.value.captcha_token = res.token
    submitMsg.value = { type: 'success', text: `验证码已发送至邮箱，${res.expires_in} 秒内有效` }
    captchaCountdown.value = 60
    clearCountdown()
    countdownTimer = setInterval(() => {
      captchaCountdown.value -= 1
      if (captchaCountdown.value <= 0) clearCountdown()
    }, 1000)
  } catch (e) {
    submitMsg.value = { type: 'error', text: (e as Error).message || '验证码发送失败' }
  }
}

async function loadPost() {
  loading.value = true
  notFound.value = false
  try {
    post.value = await getPostBySlug(slug.value)
    // 写入封面背景（SiteLayout 根据 extra.post_cover_enabled 决定是否显示）
    postCover.value = post.value?.cover_image ?? ''
    increasePostView(slug.value).catch(() => {})
  } catch {
    notFound.value = true
    postCover.value = ''
  } finally {
    loading.value = false
  }
}

async function loadComments() {
  try {
    comments.value = await listComments(slug.value, { page: 1, pageSize: 50 })
  } catch {
    // 评论加载失败不影响正文
  }
}

async function submitComment() {
  submitMsg.value = null
  if (!form.value.nickname.trim() || !form.value.content.trim()) {
    submitMsg.value = { type: 'error', text: '昵称和评论内容不能为空' }
    return
  }
  if (!form.value.email.trim()) {
    submitMsg.value = { type: 'error', text: '请填写邮箱' }
    return
  }
  if (needCaptcha.value && (!form.value.captcha_token || !form.value.captcha_code.trim())) {
    submitMsg.value = { type: 'error', text: '请先获取并填写邮箱验证码' }
    return
  }
  submitting.value = true
  try {
    const res = await createComment({
      slug: slug.value,
      nickname: form.value.nickname.trim(),
      email: form.value.email.trim(),
      content: form.value.content.trim(),
      parent_id: form.value.parent_id || 0,
      captcha_token: form.value.captcha_token,
      captcha_code: form.value.captcha_code.trim(),
    })
    if (res.status === 'pending') {
      submitMsg.value = { type: 'success', text: '评论提交成功，审核通过后将展示' }
    } else {
      submitMsg.value = { type: 'success', text: '评论提交成功' }
      await loadComments()
    }
    form.value.content = ''
    form.value.parent_id = 0
    form.value.captcha_token = ''
    form.value.captcha_code = ''
    clearCountdown()
  } catch (e) {
    submitMsg.value = { type: 'error', text: (e as Error).message || '评论提交失败' }
  } finally {
    submitting.value = false
  }
}

function replyTo(comment: Comment) {
  form.value.parent_id = comment.id
  form.value.content = `@${comment.nickname} `
  submitMsg.value = null
  // 滚动到评论表单（而非页面顶部），方便直接输入回复
  commentFormRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

watch(slug, () => {
  if (slug.value) {
    loadPost()
    loadComments()
  }
})

onMounted(async () => {
  try {
    await Promise.allSettled([loadPost(), loadComments()])
  } finally {
    markPageReady()
  }
  siteStore.fetchConfig().catch(() => {})
})

onBeforeUnmount(() => {
  clearCountdown()
  // 离开详情页时清除封面背景
  postCover.value = ''
})
</script>

<template>
  <div v-if="loading" class="blog-empty">
    <i class="fa-solid fa-spinner fa-spin"></i>
    <span style="margin-left: 8px">加载中…</span>
  </div>

  <div v-else-if="notFound" class="blog-article">
    <div class="blog-article-card">
      <div class="blog-empty">文章不存在或未发布</div>
    </div>
  </div>

  <article v-else-if="post" class="blog-article">
    <div class="blog-article-card">
      <h1>{{ post.title }}</h1>
      <div class="blog-article-meta">
        <RouterLink
          v-if="post.category"
          :to="{ path: '/posts', query: { category_id: post.category.id } }"
          :title="`查看「${post.category.name}」分类的文章`"
        >
          <i class="fa-solid fa-bookmark fa-fw"></i>
          {{ post.category.name }}
        </RouterLink>
        <RouterLink v-for="t in post.tags" :key="t.id" :to="{ path: '/posts', query: { tag_id: t.id } }">
          <i class="fa-solid fa-tag fa-fw"></i>
          {{ t.name }}
        </RouterLink>
        <span>
          <i class="fa-solid fa-calendar fa-fw"></i>
          {{ formatDate(post.published_at) }}
        </span>
        <span>
          <i class="fa-solid fa-eye fa-fw"></i>
          {{ post.view_count }}
        </span>
        <span>
          <i class="fa-solid fa-comment fa-fw"></i>
          {{ post.comment_count }}
        </span>
      </div>

      <div class="blog-markdown">
        <MdRenderer :content="post.content" />
      </div>

      <hr style="margin: 40px 0" />

      <div class="blog-comments">
        <h2 style="margin-bottom: 20px">评论 ({{ comments.total }})</h2>

        <div v-if="comments.list.length === 0" class="blog-empty">暂无评论</div>
        <div v-else>
          <CommentItem
            v-for="c in comments.list"
            :key="c.id"
            :comment="c"
            :depth="1"
            @reply="replyTo"
          />
        </div>

        <h2 ref="commentFormRef" style="margin: 35px 0 18px">发表评论</h2>
        <div
          v-if="submitMsg"
          :class="submitMsg.type === 'success' ? 'alert alert-success' : 'alert alert-error'"
        >
          {{ submitMsg.text }}
        </div>

        <form @submit.prevent="submitComment">
          <div class="grid-2">
            <div class="blog-form-group">
              <label class="blog-form-label">昵称 *</label>
              <input v-model="form.nickname" class="blog-form-input" maxlength="50" />
            </div>
            <div class="blog-form-group">
              <label class="blog-form-label">邮箱 *</label>
              <input v-model="form.email" class="blog-form-input" type="email" />
            </div>
          </div>
          <div v-if="needCaptcha" class="blog-form-group">
            <label class="blog-form-label">邮箱验证码 *</label>
            <div style="display: flex; gap: 8px">
              <input
                v-model="form.captcha_code"
                class="blog-form-input"
                maxlength="6"
                placeholder="输入验证码"
                style="flex: 1"
              />
              <button
                type="button"
                class="btn"
                :disabled="captchaCountdown > 0 || !form.email.trim()"
                @click="sendCaptcha"
              >
                {{ captchaCountdown > 0 ? `${captchaCountdown}s 后重发` : '获取验证码' }}
              </button>
            </div>
          </div>
          <div class="blog-form-group">
            <label class="blog-form-label">内容 *</label>
            <textarea v-model="form.content" class="blog-form-textarea" maxlength="1000"></textarea>
            <p v-if="form.parent_id" class="form-hint">
              正在回复评论 #{{ form.parent_id }}
              <button type="button" class="btn btn-sm" @click="form.parent_id = 0">取消回复</button>
            </p>
          </div>
          <button class="btn btn-primary" :disabled="submitting">提交评论</button>
        </form>
      </div>
    </div>
  </article>
</template>
