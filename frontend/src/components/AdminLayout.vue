<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { AUTH_EXPIRED_EVENT, clearTokens, getAccessToken, getProfile } from '@/api'
import type { User } from '@/api'
import { useTheme } from '@/composables/useTheme'
import { markPageReady } from '@/composables/usePageReady'
import { ElMessageBox } from 'element-plus'

const router = useRouter()
const route = useRoute()
const user = ref<User | null>(null)
const isCollapse = ref(false)
const { theme, toggleTheme } = useTheme()

const activeMenu = ref('/admin/dashboard')
const baseNavItems = [
  { path: '/admin/dashboard', title: '仪表盘', icon: 'Odometer' },
  { path: '/admin/posts', title: '文章', icon: 'Document' },
  { path: '/admin/categories', title: '分类', icon: 'CollectionTag' },
  { path: '/admin/tags', title: '标签', icon: 'PriceTag' },
  { path: '/admin/comments', title: '评论', icon: 'ChatDotRound' },
  { path: '/admin/configs', title: '配置', icon: 'Tools' },
  { path: '/admin/friends', title: '友链', icon: 'Link' },
  { path: '/admin/works', title: '作品', icon: 'PictureFilled' },
  { path: '/admin/musics', title: '音乐', icon: 'Headset' },
  { path: '/admin/files', title: '文件', icon: 'FolderOpened' },
]

// 用户管理仅超级管理员（is_master=1）可见
const navItems = computed(() => {
  const items = [...baseNavItems]
  if (user.value?.is_master === 1) {
    items.push({ path: '/admin/users', title: '用户', icon: 'User' })
  }
  items.push({ path: '/admin/profile', title: '个人信息', icon: 'Postcard' })
  return items
})

function handleAuthExpired() {
  user.value = null
  router.push('/admin/login')
}

async function logout() {
  try {
    await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' })
    clearTokens()
    user.value = null
    router.push('/admin/login')
  } catch {
    // 取消退出
  }
}

function go(path: string) {
  router.push(path)
}

onMounted(async () => {
  // 后台各页自带加载指示，布局挂载即视为首屏就绪（结束 App.vue 全局 loading）
  markPageReady()
  activeMenu.value = route.path
  window.addEventListener(AUTH_EXPIRED_EVENT, handleAuthExpired)

  if (getAccessToken()) {
    try {
      user.value = await getProfile()
    } catch {
      clearTokens()
      router.push('/admin/login')
    }
  } else {
    router.push('/admin/login')
  }
})

function toggleSidebar() {
  isCollapse.value = !isCollapse.value
}

onBeforeUnmount(() => {
  window.removeEventListener(AUTH_EXPIRED_EVENT, handleAuthExpired)
})
</script>

<template>
  <el-container class="admin-layout">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="admin-aside">
      <div class="admin-logo">{{ isCollapse ? '博' : '博客管理后台' }}</div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :collapse-transition="false"
        background-color="--color-admin-dark"
        text-color="#cbd5e1"
        active-text-color="#ffffff"
        @select="go"
      >
        <el-menu-item v-for="item in navItems" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container class="admin-body">
      <el-header class="admin-header">
        <div class="admin-header-left">
          <el-button text @click="toggleSidebar">
            <el-icon size="20">
              <Expand v-if="isCollapse" />
              <Fold v-else />
            </el-icon>
          </el-button>
          <div class="admin-header-title">{{ route.meta.title || '博客后台' }}</div>
        </div>
        <div class="admin-header-user">
          <el-button
            text
            :aria-label="theme === 'dark' ? '切换到浅色模式' : '切换到黑暗模式'"
            :title="theme === 'dark' ? '切换到浅色模式' : '切换到黑暗模式'"
            @click="toggleTheme"
          >
            <el-icon size="18">
              <Sunny v-if="theme === 'dark'" />
              <Moon v-else />
            </el-icon>
          </el-button>
          <span v-if="user">{{ user.username }}</span>
          <el-tag v-if="user?.is_master === 1" type="warning" size="small" style="margin: 0 8px">
            超级管理员
          </el-tag>
          <el-button text @click="logout">退出</el-button>
        </div>
      </el-header>

      <el-main class="admin-main">
        <RouterView />
      </el-main>
    </el-container>
  </el-container>
</template>
