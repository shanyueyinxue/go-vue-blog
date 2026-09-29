<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { getPublicMusics } from '@/api'
import type { APlayerOptions, PublicMusic } from '@/api'
import { useSiteStore } from '@/stores/site'

// 播放器就绪/销毁事件：供布局在播放器出现时给页脚预留空间，避免遮挡
const emit = defineEmits<{ ready: []; unready: [] }>()

const siteStore = useSiteStore()
const { config, loaded } = storeToRefs(siteStore)
const container = ref<HTMLDivElement | null>(null)

/**
 * aplayer 库与样式「按需动态引入」：不启用播放器时不加载任何 aplayer 代码/CSS
 * （静态 import 会随主包一并下载，v-if 只能控制显示无法控制加载；
 * 动态 import 让 Vite 将 aplayer 拆为独立 chunk，仅在需要创建播放器时才请求）。
 */
type APlayerConstructor = typeof import('aplayer').default
type APlayerInstance = InstanceType<APlayerConstructor>
let APlayerCtor: APlayerConstructor | null = null
let cssLoaded = false
let player: APlayerInstance | null = null
let fetching = false

async function loadAPlayer(): Promise<APlayerConstructor> {
  if (!APlayerCtor) {
    const mod = await import('aplayer')
    APlayerCtor = mod.default
  }
  if (!cssLoaded) {
    // Vite 动态导入 CSS：独立 chunk，按需加载并注入 <style>
    await import('aplayer/dist/APlayer.min.css')
    cssLoaded = true
  }
  return APlayerCtor
}

/** 读取站点配置 extra.aplayer 选项（播放器形态固定为 fixed，不可配置） */
function readOptions(): APlayerOptions {
  const raw = config.value.extra?.aplayer
  return raw && typeof raw === 'object' ? (raw as APlayerOptions) : {}
}

function destroyPlayer() {
  if (player) {
    player.destroy()
    player = null
    emit('unready')
  }
}

async function initPlayer() {
  const options = readOptions()
  if (!options.enabled) {
    destroyPlayer()
    return
  }
  if (fetching) return
  fetching = true
  let musics: PublicMusic[] = []
  try {
    musics = await getPublicMusics()
  } catch {
    // 拉取失败保持静默：不展示播放器
  } finally {
    fetching = false
  }
  if (!options.enabled) return // 拉取期间播放器被关闭
  if (!container.value) return

  const audio = musics
    .filter((m) => m.url)
    .map((m) => ({
      name: m.name,
      artist: m.artist || '未知歌手',
      url: m.url,
      cover: m.cover || undefined,
      lrc: m.lrc || undefined,
    }))
  if (!audio.length) {
    destroyPlayer()
    return
  }

  // 需要创建播放器时才按需加载 aplayer 库与样式（未启用/无曲目时不加载）
  const APlayer = await loadAPlayer()

  destroyPlayer()
  player = new APlayer({
    container: container.value,
    fixed: true, // 播放器形态固定为右下角悬浮模式，不允许通过 extra.aplayer 修改
    audio,
    autoplay: !!options.autoplay,
    theme: typeof options.theme === 'string' && options.theme ? options.theme : '#42b983',
    order: options.order === 'random' ? 'random' : 'list',
    loop: options.loop === 'one' || options.loop === 'none' ? options.loop : 'all',
    volume: typeof options.volume === 'number' ? Math.min(Math.max(options.volume, 0), 1) : 0.7,
    listFolded: !!options.listFolded,
    listMaxHeight: typeof options.listMaxHeight === 'number' ? options.listMaxHeight : 320,
  })
  emit('ready')
}

// 站点配置加载完成或 extra.aplayer 变化时初始化
watch(
  [loaded, () => config.value.extra?.aplayer],
  () => {
    if (loaded.value) initPlayer()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  destroyPlayer()
})
</script>

<template>
  <div ref="container" class="blog-aplayer" aria-hidden="true"></div>
</template>
