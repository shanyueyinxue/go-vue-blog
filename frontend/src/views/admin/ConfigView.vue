<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { activateConfig, createConfig, deleteConfig, listAdminConfigs, updateConfig, uploadFile } from '@/api'
import type { SiteConfig } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { formatDate } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'

const { result, loading, load, changePage } = usePagination<SiteConfig>((p) => listAdminConfigs(p))
const editing = ref<SiteConfig | null>(null)
const form = ref({
  version: '',
  title: '',
  subtitle: '',
  side_name: '',
  description: '',
  avatar: '',
  icp: '',
  police_icp: '',
  theme: '',
  about_content: '',
  home_images: '',
  comment_need_review: 1,
  comment_need_captcha: 1,
  custom_css_vars: '',
  custom_css_vars_dark: '',
  custom_css_enabled: 0,
  social_links: '[]',
  extra: '{}',
})
const saving = ref(false)
const errorMsg = ref('')

function parseJson(str: string): Record<string, unknown> {
  try {
    const v = JSON.parse(str || '{}')
    return v && typeof v === 'object' ? v : {}
  } catch {
    return {}
  }
}

/** 解析社交链接 JSON（数组格式），兼容旧的对象格式，统一返回 [{name, url}] 数组 */
function parseSocialLinks(str: string): Array<{ name: string; url: string }> {
  try {
    const v = JSON.parse(str || '[]')
    if (Array.isArray(v)) {
      return v
        .filter((item): item is { name?: unknown; url?: unknown } => !!item && typeof item === 'object')
        .map((item) => ({ name: String(item.name ?? ''), url: String(item.url ?? '') }))
        .filter((item) => item.name || item.url)
    }
    if (v && typeof v === 'object') {
      return Object.entries(v as Record<string, unknown>).map(([name, url]) => ({ name, url: String(url) }))
    }
  } catch {
    // 解析失败返回空数组
  }
  return []
}

/** 社交链接（任意格式）→ 格式化 JSON 数组字符串（编辑时展示） */
function formatSocialLinks(raw: string): string {
  return JSON.stringify(parseSocialLinks(raw), null, 2)
}

/** 解析导航栏背景图 JSON（{导航label: 图片URL}），仅保留字符串值 */
function parseNavImages(str: string): Record<string, string> {
  try {
    const v = JSON.parse(str || '{}')
    if (v && typeof v === 'object' && !Array.isArray(v)) {
      const out: Record<string, string> = {}
      for (const [key, value] of Object.entries(v as Record<string, unknown>)) {
        if (typeof value === 'string' && value) out[key] = value
      }
      return out
    }
  } catch {
    // 非法 JSON 返回空对象
  }
  return {}
}

/* ----------------------------- 扩展字段（extra）设置 ----------------------------- */
// 以下设置项存于站点配置 extra JSON，随 PublicConfig.extra 透出前台：
// - hideNavItems           隐藏指定导航项（SiteLayout 按导航 label 匹配）
// - dark_avatar            黑暗主题头像（HomeView 优先使用，未设置回退 avatar）
// - dark_home_images       黑暗主题首页背景图（HomeView 优先使用，未设置回退 home_images）
// - aplayer                音乐播放器选项（播放器形态固定 fixed，仅播放行为类选项）
// - card_opacity           卡片背景透明度（0-1，控制 blog-article-card / blog-post 背景）
// - card_text_fade         卡片文字是否同步变淡（true 时整卡 opacity）
// - nav_opacity            导航栏背景透明度（0-1，控制 blog-menu 背景）
// - nav_text_fade          导航栏文字是否同步变淡（true 时整条导航栏 opacity）
// - post_cover_enabled     文章详情封面是否作为整页固定背景（SiteLayout + PostDetailView）
// - nav_bg_image           页面背景图（{导航label: 图片URL}，作用于 blog-shell，浅色主题）
// - nav_bg_image_dark      页面背景图（{导航label: 图片URL}，作用于 blog-shell，深色主题，缺失回退浅色）
// - footer_inner_bg        页脚内层背景色（blog-footer-inner，支持带透明度颜色）
// 这些键由下方设置区维护，其余键可在「其他配置（JSON）」中手写兜底。

/** 前台导航项 label（与 SiteLayout.vue navItems 保持一致） */
const NAV_ITEMS = ['首页', '文章', '分类', '标签', '归档', '作品', '友链', '关于']

const hideNavItemsForm = ref<string[]>([])
const darkAvatarForm = ref('')
const darkHomeImagesForm = ref('')
const darkAvatarUploading = ref(false)
const darkHomeImageUploading = ref(false)

const aplayerForm = ref({
  enabled: false,
  autoplay: false,
  theme: '#42b983',
  order: 'list' as 'list' | 'random',
  loop: 'all' as 'all' | 'one' | 'none',
  volume: 0.7,
  listFolded: false,
  listMaxHeight: 320,
})

// 卡片透明度 / 封面背景 / 导航栏透明度 / 导航栏背景图 / 页脚背景色
const cardOpacityForm = ref(1)
const cardTextFadeForm = ref(false)
const postCoverEnabledForm = ref(false)
const navOpacityForm = ref(1)
const navTextFadeForm = ref(false)
const navBgImageForm = ref('{}')
const navBgImageDarkForm = ref('{}')
const footerInnerBgForm = ref('')

function resetAplayerForm() {
  aplayerForm.value = {
    enabled: false,
    autoplay: false,
    theme: '#42b983',
    order: 'list',
    loop: 'all',
    volume: 0.7,
    listFolded: false,
    listMaxHeight: 320,
  }
}

/** 重置全部扩展字段设置区（新建/复制配置时调用） */
function resetExtraFields() {
  hideNavItemsForm.value = []
  darkAvatarForm.value = ''
  darkHomeImagesForm.value = ''
  resetAplayerForm()
  cardOpacityForm.value = 1
  cardTextFadeForm.value = false
  postCoverEnabledForm.value = false
  navOpacityForm.value = 1
  navTextFadeForm.value = false
  navBgImageForm.value = '{}'
  navBgImageDarkForm.value = '{}'
  footerInnerBgForm.value = ''
}

/** dark_home_images 文本域（每行一个 URL）→ 字符串数组 */
function darkHomeImagesToArray(): string[] {
  return darkHomeImagesForm.value
    .split(/\r?\n/)
    .map((url) => url.trim())
    .filter(Boolean)
}

/**
 * 从 extra JSON 中加载各设置区维护的键（hideNavItems / dark_avatar / dark_home_images /
 * aplayer / card_opacity / card_text_fade / post_cover_enabled / nav_bg_image /
 * nav_bg_image_dark / footer_inner_bg），并将这些键从「其他配置」文本域中剥离
 * （设置区为唯一入口，避免与手写 JSON 冲突）。返回剩余 extra 的格式化 JSON 字符串。
 */
function loadExtraFields(extraJson: string): string {
  hideNavItemsForm.value = []
  darkAvatarForm.value = ''
  darkHomeImagesForm.value = ''
  resetAplayerForm()
  cardOpacityForm.value = 1
  cardTextFadeForm.value = false
  postCoverEnabledForm.value = false
  navOpacityForm.value = 1
  navTextFadeForm.value = false
  navBgImageForm.value = '{}'
  navBgImageDarkForm.value = '{}'
  footerInnerBgForm.value = ''

  const extra = parseJson(extraJson)

  const hidden = extra.hideNavItems
  if (Array.isArray(hidden)) {
    hideNavItemsForm.value = hidden.filter((v): v is string => typeof v === 'string')
  }
  if (typeof extra.dark_avatar === 'string') {
    darkAvatarForm.value = extra.dark_avatar
  }
  const darkImages = extra.dark_home_images
  if (Array.isArray(darkImages)) {
    darkHomeImagesForm.value = darkImages
      .filter((v): v is string => typeof v === 'string')
      .join('\n')
  }

  const raw = extra.aplayer
  if (raw && typeof raw === 'object') {
    const o = raw as Record<string, unknown>
    if (typeof o.enabled === 'boolean') aplayerForm.value.enabled = o.enabled
    if (typeof o.autoplay === 'boolean') aplayerForm.value.autoplay = o.autoplay
    if (typeof o.theme === 'string' && o.theme) aplayerForm.value.theme = o.theme
    if (o.order === 'random' || o.order === 'list') aplayerForm.value.order = o.order
    if (o.loop === 'all' || o.loop === 'one' || o.loop === 'none') aplayerForm.value.loop = o.loop
    if (typeof o.volume === 'number') aplayerForm.value.volume = Math.min(Math.max(o.volume, 0), 1)
    if (typeof o.listFolded === 'boolean') aplayerForm.value.listFolded = o.listFolded
    if (typeof o.listMaxHeight === 'number') aplayerForm.value.listMaxHeight = o.listMaxHeight
  }

  // 卡片透明度（0-1）
  if (typeof extra.card_opacity === 'number' && extra.card_opacity >= 0 && extra.card_opacity <= 1) {
    cardOpacityForm.value = extra.card_opacity
  }
  if (extra.card_text_fade === true) cardTextFadeForm.value = true
  if (extra.post_cover_enabled === true) postCoverEnabledForm.value = true
  // 导航栏透明度（0-1）
  if (typeof extra.nav_opacity === 'number' && extra.nav_opacity >= 0 && extra.nav_opacity <= 1) {
    navOpacityForm.value = extra.nav_opacity
  }
  if (extra.nav_text_fade === true) navTextFadeForm.value = true
  // 导航栏背景图（{label: url} 映射，展示为格式化 JSON）
  if (extra.nav_bg_image && typeof extra.nav_bg_image === 'object') {
    navBgImageForm.value = JSON.stringify(extra.nav_bg_image, null, 2)
  }
  if (extra.nav_bg_image_dark && typeof extra.nav_bg_image_dark === 'object') {
    navBgImageDarkForm.value = JSON.stringify(extra.nav_bg_image_dark, null, 2)
  }
  if (typeof extra.footer_inner_bg === 'string') {
    footerInnerBgForm.value = extra.footer_inner_bg
  }

  delete extra.hideNavItems
  delete extra.dark_avatar
  delete extra.dark_home_images
  delete extra.aplayer
  delete extra.card_opacity
  delete extra.card_text_fade
  delete extra.post_cover_enabled
  delete extra.nav_opacity
  delete extra.nav_text_fade
  delete extra.nav_bg_image
  delete extra.nav_bg_image_dark
  delete extra.footer_inner_bg
  return JSON.stringify(extra, null, 2)
}

/** 将各设置区合并进扩展字段对象（设置区为这些键的唯一入口） */
function mergeExtraFields(extra: Record<string, unknown>): Record<string, unknown> {
  return {
    ...extra,
    hideNavItems: hideNavItemsForm.value,
    dark_avatar: darkAvatarForm.value,
    dark_home_images: darkHomeImagesToArray(),
    aplayer: { ...aplayerForm.value },
    card_opacity: cardOpacityForm.value,
    card_text_fade: cardTextFadeForm.value,
    post_cover_enabled: postCoverEnabledForm.value,
    nav_opacity: navOpacityForm.value,
    nav_text_fade: navTextFadeForm.value,
    nav_bg_image: parseNavImages(navBgImageForm.value),
    nav_bg_image_dark: parseNavImages(navBgImageDarkForm.value),
    footer_inner_bg: footerInnerBgForm.value,
  }
}

async function handleDarkAvatarUpload(options: { file: File | Blob }) {
  const file = options.file
  if (!(file instanceof File) || !isImageFile(file)) {
    ElMessage.warning('请选择图片文件')
    return
  }
  darkAvatarUploading.value = true
  try {
    const result = await uploadFile(file)
    darkAvatarForm.value = result.url
    ElMessage.success('黑暗主题头像上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    darkAvatarUploading.value = false
  }
}

async function handleDarkHomeImageUpload(options: { file: File | Blob }) {
  const file = options.file
  if (!(file instanceof File) || !isImageFile(file)) {
    ElMessage.warning('请选择图片文件')
    return
  }
  darkHomeImageUploading.value = true
  try {
    const result = await uploadFile(file)
    const urls = darkHomeImagesToArray()
    urls.push(result.url)
    darkHomeImagesForm.value = urls.join('\n')
    ElMessage.success('背景图上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    darkHomeImageUploading.value = false
  }
}

function openNew() {
  editing.value = null
  form.value = {
    version: '',
    title: '',
    subtitle: '',
    side_name: '',
    description: '',
    avatar: '',
    icp: '',
    police_icp: '',
    theme: '',
    about_content: '',
    home_images: '',
    comment_need_review: 1,
    comment_need_captcha: 1,
    custom_css_vars: '',
    custom_css_vars_dark: '',
    custom_css_enabled: 0,
    social_links: '[]',
    extra: '{}',
  }
  resetExtraFields()
  errorMsg.value = ''
}

function openEdit(c: SiteConfig) {
  editing.value = c
  form.value = {
    version: c.version,
    title: c.title,
    subtitle: c.subtitle,
    side_name: c.side_name || '',
    description: c.description,
    avatar: c.avatar,
    icp: c.icp,
    police_icp: c.police_icp,
    theme: c.theme,
    about_content: c.about_content,
    home_images: formatHomeImages(c.home_images),
    comment_need_review: c.comment_need_review,
    comment_need_captcha: c.comment_need_captcha,
    custom_css_vars: c.custom_css_vars || '',
    custom_css_vars_dark: c.custom_css_vars_dark || '',
    custom_css_enabled: c.custom_css_enabled,
    social_links: formatSocialLinks(c.social_links || '[]'),
    extra: loadExtraFields(c.extra || '{}'),
  }
  errorMsg.value = ''
}

/** 复制当前配置到新建表单（版本号清空） */
function copyConfig(c: SiteConfig) {
  editing.value = null
  form.value = {
    version: '',
    title: c.title,
    subtitle: c.subtitle,
    side_name: c.side_name || '',
    description: c.description,
    avatar: c.avatar,
    icp: c.icp,
    police_icp: c.police_icp,
    theme: c.theme,
    about_content: c.about_content,
    home_images: formatHomeImages(c.home_images),
    comment_need_review: c.comment_need_review,
    comment_need_captcha: c.comment_need_captcha,
    custom_css_vars: c.custom_css_vars || '',
    custom_css_vars_dark: c.custom_css_vars_dark || '',
    custom_css_enabled: c.custom_css_enabled,
    social_links: formatSocialLinks(c.social_links || '[]'),
    extra: loadExtraFields(c.extra || '{}'),
  }
  errorMsg.value = ''
  ElMessage.success('已复制当前配置，请填写新版本号后保存')
}

function formatHomeImages(raw: string): string {
  try {
    const value = JSON.parse(raw || '[]')
    if (Array.isArray(value)) {
      return value.filter((item): item is string => typeof item === 'string').join('\n')
    }
  } catch {
    // 非 JSON 时按普通字符串处理
  }
  return raw || ''
}

function isImageFile(file: File): boolean {
  return file.type.startsWith('image/')
}

const avatarUploading = ref(false)
const homeImageUploading = ref(false)

async function handleAvatarUpload(options: { file: File | Blob }) {
  const file = options.file
  if (!(file instanceof File) || !isImageFile(file)) {
    ElMessage.warning('请选择图片文件')
    return
  }
  avatarUploading.value = true
  try {
    const result = await uploadFile(file)
    form.value.avatar = result.url
    ElMessage.success('头像上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    avatarUploading.value = false
  }
}

async function handleHomeImageUpload(options: { file: File | Blob }) {
  const file = options.file
  if (!(file instanceof File) || !isImageFile(file)) {
    ElMessage.warning('请选择图片文件')
    return
  }
  homeImageUploading.value = true
  try {
    const result = await uploadFile(file)
    const urls = form.value.home_images
      .split(/\r?\n/)
      .map((url) => url.trim())
      .filter(Boolean)
    urls.push(result.url)
    form.value.home_images = urls.join('\n')
    ElMessage.success('背景图上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    homeImageUploading.value = false
  }
}

async function save() {
  errorMsg.value = ''
  if (!form.value.version.trim()) {
    errorMsg.value = '版本号不能为空'
    return
  }
  saving.value = true
  try {
    const params = {
      version: form.value.version.trim(),
      title: form.value.title,
      subtitle: form.value.subtitle,
      side_name: form.value.side_name,
      description: form.value.description,
      avatar: form.value.avatar,
      icp: form.value.icp,
      police_icp: form.value.police_icp,
      theme: form.value.theme,
      about_content: form.value.about_content,
      home_images: form.value.home_images
        .split(/\r?\n/)
        .map((url) => url.trim())
        .filter(Boolean),
      comment_need_review: form.value.comment_need_review,
      comment_need_captcha: form.value.comment_need_captcha,
      custom_css_vars: form.value.custom_css_vars,
      custom_css_vars_dark: form.value.custom_css_vars_dark,
      custom_css_enabled: form.value.custom_css_enabled,
      social_links: parseSocialLinks(form.value.social_links),
      extra: mergeExtraFields(parseJson(form.value.extra)),
    }
    if (editing.value) {
      await updateConfig(editing.value.id, params)
      ElMessage.success('配置已更新')
    } else {
      await createConfig(params)
      ElMessage.success('配置已创建')
    }
    openNew()
    await load()
  } catch (e) {
    errorMsg.value = (e as Error).message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function activate(c: SiteConfig) {
  try {
    await activateConfig(c.id)
    ElMessage.success('配置已激活')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message || '激活失败')
  }
}

async function remove(c: SiteConfig) {
  try {
    await ElMessageBox.confirm('确定删除该配置？', '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteConfig(c.id)
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
      <h1>站点配置</h1>
      <el-button @click="openNew">清空</el-button>
    </div>

    <el-card style="margin-bottom: 20px">
      <template #header>{{ editing ? `编辑配置 v${editing.version}` : '新建配置' }}</template>
      <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon style="margin-bottom: 16px" />

      <el-form label-position="top">
        <el-row :gutter="16">
          <el-col :xs="24" :md="12">
            <el-form-item label="版本号" required>
              <el-input v-model="form.version" placeholder="如 1.0.0" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="主题">
              <el-select v-model="form.theme" style="width: 100%">
                <el-option label="跟随系统" value="" />
                <el-option label="浅色" value="light" />
                <el-option label="黑暗" value="dark" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="标题">
              <el-input v-model="form.title" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="副标题">
              <el-input v-model="form.subtitle" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="侧栏卡片名称">
              <el-input v-model="form.side_name" placeholder="留空则使用站点标题" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="描述">
              <el-input v-model="form.description" type="textarea" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="博客头像 URL">
              <el-input v-model="form.avatar" placeholder="https://…">
                <template #append>
                  <el-upload :show-file-list="false" accept="image/*" :http-request="handleAvatarUpload"
                    :disabled="avatarUploading">
                    <el-button :loading="avatarUploading" :disabled="avatarUploading">上传</el-button>
                  </el-upload>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="ICP 备案号">
              <el-input v-model="form.icp"
                placeholder="如 <a href=&quot;https://beian.miit.gov.cn/&quot; target=&quot;_blank&quot;>京ICP备xxx号</a>" />
              <p class="form-hint">纯文本或 HTML 链接标签，前台渲染前会做安全消毒（仅允许安全链接）</p>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="公安备案号">
              <el-input v-model="form.police_icp" placeholder="纯文本或 HTML 链接" />
              <p class="form-hint">纯文本或 HTML 链接标签，前台渲染前会做安全消毒（仅允许安全链接）</p>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="评论审核">
              <el-select v-model="form.comment_need_review" style="width: 100%">
                <el-option label="需要审核" :value="1" />
                <el-option label="直接通过" :value="0" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="评论验证码">
              <el-select v-model="form.comment_need_captcha" style="width: 100%">
                <el-option label="需要邮箱验证码" :value="1" />
                <el-option label="不需要" :value="0" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="About 内容（Markdown）">
          <el-input v-model="form.about_content" type="textarea" :rows="10" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :xs="24" :md="12">
            <el-form-item label="首页随机背景图（每行一个 URL）">
              <div style="display: flex; gap: 8px; align-items: flex-start; width: 100%">
                <el-input v-model="form.home_images" type="textarea" :rows="4"
                  placeholder="https://example.com/bg1.jpg" />
                <el-upload :show-file-list="false" accept="image/*" :http-request="handleHomeImageUpload"
                  :disabled="homeImageUploading">
                  <el-button :loading="homeImageUploading" :disabled="homeImageUploading">上传</el-button>
                </el-upload>
              </div>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="社交链接（JSON）">
              <el-input v-model="form.social_links" type="textarea" :rows="4" class="mono-input" />
              <p class="form-hint">
                数组顺序即前台显示顺序，如 [{"name":"github","url":"https://github.com/xx"}]；旧的对象格式
                {"github":"url"} 仍兼容（自动按键排序转换）
              </p>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :xs="24" :md="12">
            <el-form-item label="自定义 CSS 变量">
              <div style="display: flex; gap: 12px; align-items: flex-start; width: 100%">
                <el-input v-model="form.custom_css_vars" type="textarea" :rows="4" class="mono-input"
                  placeholder="--blog-primary: #66afef;&#10;--font-family-sans: 'Noto Sans SC', sans-serif;"
                  style="flex: 1" />
              </div>
              <p class="form-hint">
                仅支持平坦 CSS 变量（每行一个 <code>--变量名: 值;</code>），可覆盖
                <code>theme-vars.css</code> 中的任意变量（如 --blog-primary、--font-family-sans），
                用于调整主题样式；值后面的注释（<code>/* ... */</code> 或 <code>// ...</code>）会被忽略。
              </p>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="自定义 CSS 变量（黑暗主题）">
              <div style="display: flex; gap: 12px; align-items: flex-start; width: 100%">
                <el-input v-model="form.custom_css_vars_dark" type="textarea" :rows="4" class="mono-input"
                  placeholder="--blog-bg: #0f1115;&#10;--hljs-base: #c9d1d9;" />
                <el-select v-model="form.custom_css_enabled" style="width: 140px">
                  <el-option label="已启用" :value="1" />
                  <el-option label="未启用" :value="0" />
                </el-select>
              </div>
              <p class="form-hint">
                仅在黑暗主题（<code>data-theme="dark"</code>）下生效，优先级高于上面的通用变量；
                切换主题时由 CSS 自动生效，无需额外处理。
              </p>
            </el-form-item>
          </el-col>
        </el-row>




        <el-divider content-position="left">扩展字段</el-divider>

        <el-form-item label="隐藏导航项 (hideNavItems)">
          <el-select v-model="hideNavItemsForm" multiple placeholder="选择要隐藏的导航项" style="width: 100%">
            <el-option v-for="item in NAV_ITEMS" :key="item" :label="item" :value="item" />
          </el-select>
          <p class="form-hint">勾选的导航项不再显示在前台菜单中（按导航 label 匹配）</p>
        </el-form-item>

        <el-form-item label="黑暗主题头像 (dark_avatar)">
          <el-input v-model="darkAvatarForm" placeholder="https://…">
            <template #append>
              <el-upload :show-file-list="false" accept="image/*" :http-request="handleDarkAvatarUpload"
                :disabled="darkAvatarUploading">
                <el-button :loading="darkAvatarUploading" :disabled="darkAvatarUploading">上传</el-button>
              </el-upload>
            </template>
          </el-input>
          <p class="form-hint">黑暗主题下显示的头像；未设置时回退常规头像</p>
        </el-form-item>

        <el-form-item label="黑暗主题首页背景图 (dark_home_images)">
          <div style="display: flex; gap: 8px; align-items: flex-start; width: 100%">
            <el-input v-model="darkHomeImagesForm" type="textarea" :rows="4"
              placeholder="https://example.com/bg1.jpg" />
            <el-upload :show-file-list="false" accept="image/*" :http-request="handleDarkHomeImageUpload"
              :disabled="darkHomeImageUploading">
              <el-button :loading="darkHomeImageUploading" :disabled="darkHomeImageUploading">上传</el-button>
            </el-upload>
          </div>
          <p class="form-hint">黑暗主题下优先使用的首页背景图（每行一个 URL）；未设置时回退常规背景图</p>
        </el-form-item>

        <el-divider content-position="left">卡片透明度 (card_opacity / card_text_fade)</el-divider>
        <el-row :gutter="12">
          <el-col :xs="24" :md="12">
            <el-form-item label="卡片背景透明度（0-1）">
              <el-slider v-model="cardOpacityForm" :min="0" :max="1" :step="0.05" show-input
                style="padding: 0 8px" />
              <p class="form-hint">控制文章列表卡片（blog-post）与文章详情卡片（blog-article-card）的<b>背景</b>透明度；
                数值越小越透，页面背景/封面可见，文字保持不透明</p>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="文字同步变淡 (card_text_fade)">
              <el-switch v-model="cardTextFadeForm" />
              <p class="form-hint">开启后整卡（含文字）按上方透明度同步变淡；关闭则仅背景透明</p>
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">导航栏透明度 (nav_opacity / nav_text_fade)</el-divider>
        <el-row :gutter="12">
          <el-col :xs="24" :md="12">
            <el-form-item label="导航栏背景透明度（0-1）">
              <el-slider v-model="navOpacityForm" :min="0" :max="1" :step="0.05" show-input
                style="padding: 0 8px" />
              <p class="form-hint">控制顶部导航栏（blog-menu）的<b>背景</b>透明度；数值越小越透，文字保持不透明</p>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="文字同步变淡 (nav_text_fade)">
              <el-switch v-model="navTextFadeForm" />
              <p class="form-hint">开启后整条导航栏（含文字）按上方透明度同步变淡；关闭则仅背景透明</p>
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">文章详情封面背景 (post_cover_enabled)</el-divider>
        <el-form-item label="开启封面背景">
          <el-switch v-model="postCoverEnabledForm" />
          <p class="form-hint">
            开启后，文章详情页若设置了封面图，将封面作为整页固定背景显示
            （background-attachment: fixed 铺满视口，卡片透明时透出）；未开启或无封面时使用默认背景
          </p>
        </el-form-item>

        <el-divider content-position="left">页面背景图 (nav_bg_image / nav_bg_image_dark)</el-divider>
        <el-row :gutter="12">
          <el-col :xs="24" :md="12">
            <el-form-item label="浅色主题（每页一张图）">
              <el-input v-model="navBgImageForm" type="textarea" :rows="5" class="mono-input" />
              <p class="form-hint">JSON 对象，键为导航项 label（首页/文章/分类/标签/归档/作品/友链/关于），
                如 {"首页":"https://...","文章":"https://..."}；作为 blog-shell 整页背景，按当前页面切换</p>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="深色主题（每页一张图）">
              <el-input v-model="navBgImageDarkForm" type="textarea" :rows="5" class="mono-input" />
              <p class="form-hint">深色主题下优先使用；某页面未配置时回退左侧浅色主题的对应图</p>
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">页脚背景色 (footer_inner_bg)</el-divider>
        <el-form-item label="页脚背景色">
          <el-color-picker v-model="footerInnerBgForm" show-alpha />
          <p class="form-hint">blog-footer-inner 的背景色，支持带透明度（拖动 Alpha 滑块）；留空则不设置（保持默认透明）</p>
        </el-form-item>

        <el-divider content-position="left">音乐播放器（APlayer）</el-divider>
        <el-form-item label="启用播放器">
          <el-switch v-model="aplayerForm.enabled" />
          <p class="form-hint">
            开启后前台右下角显示悬浮播放器（形态固定），播放启用状态的音乐；
            需先在「音乐管理」中添加曲目。选项存于站点配置扩展字段 <code>extra.aplayer</code>
          </p>
        </el-form-item>
        <el-row :gutter="12">
          <el-col :xs="24" :md="12">
            <el-form-item label="自动播放">
              <el-switch v-model="aplayerForm.autoplay" />
              <p class="form-hint">浏览器可能拦截带声音的自动播放</p>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="主题色">
              <el-color-picker v-model="aplayerForm.theme" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="播放顺序">
              <el-select v-model="aplayerForm.order" style="width: 100%">
                <el-option label="列表顺序" value="list" />
                <el-option label="随机播放" value="random" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="循环模式">
              <el-select v-model="aplayerForm.loop" style="width: 100%">
                <el-option label="全部循环" value="all" />
                <el-option label="单曲循环" value="one" />
                <el-option label="不循环" value="none" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="默认音量">
              <el-slider v-model="aplayerForm.volume" :min="0" :max="1" :step="0.05" show-input style="padding: 0 8px" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="播放列表默认收起">
              <el-switch v-model="aplayerForm.listFolded" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item label="列表最大高度（px）">
              <el-input-number v-model="aplayerForm.listMaxHeight" :min="120" :max="800" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">其他配置</el-divider>
        <el-form-item label="其他配置（JSON）">
          <el-input v-model="form.extra" type="textarea" :rows="4" class="mono-input" />
          <p class="form-hint">
            以上设置区未覆盖的扩展字段可在此以 JSON 书写（如
            <code>{"custom_key": "value"}</code>）；
            <code>hideNavItems</code> / <code>dark_avatar</code> / <code>dark_home_images</code> /
            <code>aplayer</code> / <code>card_opacity</code> / <code>card_text_fade</code> /
            <code>nav_opacity</code> / <code>nav_text_fade</code> /
            <code>post_cover_enabled</code> / <code>nav_bg_image</code> / <code>nav_bg_image_dark</code> /
            <code>footer_inner_bg</code> 由上方设置区维护，请勿在此手写
          </p>
        </el-form-item>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </el-form>
    </el-card>

    <el-card>
      <template #header>配置版本列表</template>
      <el-table v-loading="loading" :data="result.list">
        <el-table-column label="版本" width="120">
          <template #default="{ row }">v{{ row.version }}</template>
        </el-table-column>
        <el-table-column label="标题" min-width="180">
          <template #default="{ row }">{{ row.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="row.is_active === 1 ? 'success' : 'info'">
              {{ row.is_active === 1 ? '生效中' : '未生效' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="190">
          <template #default="{ row }">{{ formatDate(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="warning" @click="copyConfig(row)">复制</el-button>
            <el-button v-if="row.is_active !== 1" link type="success" @click="activate(row)">激活</el-button>
            <el-button v-if="row.is_active !== 1" link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination v-if="result.total > result.pageSize" style="margin-top: 16px; justify-content: flex-end"
        background layout="prev, pager, next, total" :total="result.total" :page-size="result.pageSize"
        :current-page="result.page" @current-change="changePage" />
    </el-card>
  </div>
</template>
