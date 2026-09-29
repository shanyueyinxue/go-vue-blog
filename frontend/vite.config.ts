import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import Sitemap from 'vite-plugin-sitemap'


// 开发环境代理目标：后端默认端口 8090（见 backend/config-example.yaml server.port），
// 可用环境变量 VITE_API_PROXY_TARGET 覆盖，如 http://127.0.0.1:8080
const apiTarget = process.env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8090'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
    // Element Plus 按需导入：只打包模板中实际使用的组件/指令，
    // 样式仍由 main.ts 的 element-plus/dist/index.css 全量提供（importStyle: false）。
    // dts: false —— 不生成类型声明，保持 el-* 组件与改动前一致的宽松类型行为。
    Components({
      dts: false,
      resolvers: [
        ElementPlusResolver({ importStyle: false }),
      ],
    }),
    Sitemap({
      hostname: 'https://learn.com.cn',  // 必填
      // 如果你的路由是动态生成的（如博客文章），在这里手动列出
      dynamicRoutes: [
        '/',
        '/about',
        // 如果有动态路由，可以在这里用数组生成
      ],
      // 排除不需要收录的页面
      exclude: ['/admin', '/posts'],
      robots: [
        {
          userAgent: '*',
          allow: '/',
          disallow: '/admin',
        }
      ]
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    // 使用 Rolldown 默认分包策略。注意：不要在此处配置自定义 advancedChunks 分组 ——
    // 曾将 vue/axios/element-plus 等拆为独立 vendor 块，导致 CJS 互操作助手
    // （__exportAll/__commonJSMin/__toESM）与消费方跨块分布，形成 api ↔ vendor
    // 循环依赖，生产包运行时抛 "e is not a function"（模块初始化顺序被破坏）。
    // 默认分包同样会按异步路由自动拆出共享块，具备基本的缓存复用能力。
    chunkSizeWarningLimit: 1100,
  },
  server: {
    // 开发环境代理：将 /api 请求转发到后端服务（见 backend/config-example.yaml server.port）
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true,
      },
      '/uploads': {
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
})
