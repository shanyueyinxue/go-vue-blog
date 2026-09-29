import { createRouter, createWebHistory } from 'vue-router'
import SiteLayout from '@/components/SiteLayout.vue'
import AdminLayout from '@/components/AdminLayout.vue'
import { getAccessToken, getProfile } from '@/api'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  // 路由切换时平滑滚动到顶部（点击导航栏项时有动画效果）；
  // 同一路由仅 query 变化（分页/筛选）不强制滚动；浏览器前进/后退保留原位置
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    if (to.path !== from.path) {
      return { top: 0, behavior: 'smooth' }
    }
    return false
  },
  routes: [
    {
      path: '/',
      component: SiteLayout,
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('@/views/HomeView.vue'),
          // 首页标题直接使用站点名（meta.title 留空），描述可单独定制
          meta: { title: '', description: '技术笔记、生活随想与经验分享。' },
        },
        {
          path: 'archive',
          name: 'archive',
          component: () => import('@/views/ArchiveView.vue'),
          meta: { title: '归档', description: '按时间归档浏览本站全部文章。' },
        },
        {
          path: 'posts',
          name: 'posts',
          component: () => import('@/views/PostsView.vue'),
          meta: { title: '文章', description: '浏览本站全部文章，可按分类或标签筛选。' },
        },
        {
          path: 'categories',
          name: 'categories',
          component: () => import('@/views/CategoriesView.vue'),
          meta: { title: '分类', description: '按分类浏览本站文章。' },
        },
        {
          path: 'tags',
          name: 'tags',
          component: () => import('@/views/TagsView.vue'),
          meta: { title: '标签', description: '按标签浏览本站文章。' },
        },
        {
          path: 'friends',
          name: 'friends',
          component: () => import('@/views/FriendsView.vue'),
          meta: { title: '友情链接', description: '与本站互链的朋友们。' },
        },
        {
          path: 'works',
          name: 'works',
          component: () => import('@/views/WorksView.vue'),
          meta: { title: '作品', description: '本站作者的作品展示。' },
        },
        {
          path: 'about',
          name: 'about',
          component: () => import('@/views/AboutView.vue'),
          meta: { title: '关于', description: '关于本站与作者。' },
        },
        {
          path: 'search',
          name: 'search',
          component: () => import('@/views/SearchView.vue'),
          meta: { title: '搜索', description: '站内搜索文章。' },
        },
        {
          path: 'post/:slug',
          name: 'post-detail',
          component: () => import('@/views/PostDetailView.vue'),
          // 具体标题/描述/关键词由页面组件根据后端返回的文章数据动态覆盖（见 PostDetailView.vue）
          meta: { title: '文章详情', description: '阅读文章全文、发表评论。' },
        },
        {
          // 兜底 404：未匹配的任意路径（含 /admin/xxx 之外的未知地址）
          path: ':pathMatch(.*)*',
          name: 'not-found',
          component: () => import('@/views/NotFoundView.vue'),
          meta: { title: '页面不存在', description: '您访问的页面不存在或已被移除。' },
        },
      ],
    },
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('@/views/admin/LoginView.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/admin',
      component: AdminLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/admin/dashboard',
        },
        {
          path: 'dashboard',
          name: 'admin-dashboard',
          component: () => import('@/views/admin/DashboardView.vue'),
        },
        {
          path: 'profile',
          name: 'admin-profile',
          component: () => import('@/views/admin/ProfileView.vue'),
        },
        {
          path: 'users',
          name: 'admin-users',
          component: () => import('@/views/admin/UserView.vue'),
          meta: { requiresMaster: true },
        },
        {
          path: 'posts',
          name: 'admin-posts',
          component: () => import('@/views/admin/PostListView.vue'),
        },
        {
          path: 'posts/new',
          name: 'admin-post-new',
          component: () => import('@/views/admin/PostEditView.vue'),
        },
        {
          path: 'posts/:id/edit',
          name: 'admin-post-edit',
          component: () => import('@/views/admin/PostEditView.vue'),
        },
        {
          path: 'categories',
          name: 'admin-categories',
          component: () => import('@/views/admin/CategoryView.vue'),
        },
        {
          path: 'tags',
          name: 'admin-tags',
          component: () => import('@/views/admin/TagView.vue'),
        },
        {
          path: 'comments',
          name: 'admin-comments',
          component: () => import('@/views/admin/CommentView.vue'),
        },
        {
          path: 'configs',
          name: 'admin-configs',
          component: () => import('@/views/admin/ConfigView.vue'),
        },
        {
          path: 'friends',
          name: 'admin-friends',
          component: () => import('@/views/admin/FriendView.vue'),
        },
        {
          path: 'works',
          name: 'admin-works',
          component: () => import('@/views/admin/WorkView.vue'),
        },
        {
          path: 'musics',
          name: 'admin-musics',
          component: () => import('@/views/admin/MusicView.vue'),
        },
        {
          path: 'files',
          name: 'admin-files',
          component: () => import('@/views/admin/FileView.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.requiresAuth && !getAccessToken()) {
    return { name: 'admin-login' }
  }
  if (to.name === 'admin-login' && getAccessToken()) {
    return { name: 'admin-dashboard' }
  }
  // 用户管理等仅超级管理员（is_master=1）可访问：后端 403 为最终保障，
  // 前端守卫提前拦截并跳回仪表盘
  if (to.meta.requiresMaster) {
    try {
      const profile = await getProfile()
      if (profile.is_master !== 1) {
        return { name: 'admin-dashboard' }
      }
    } catch {
      return { name: 'admin-login' }
    }
  }
  return true
})

export default router
