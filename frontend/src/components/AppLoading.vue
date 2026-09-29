<script setup lang="ts">
/**
 * 全局 loading（由独立 Vue 实例渲染，见 main.ts）
 *
 * 生命周期：入口 chunk 执行后立即挂载 → 覆盖重资源动态加载、站点配置预取、
 * 主应用首屏数据加载全过程 → main.ts 在 pageReady + window load（或 10s 兜底）后 unmount。
 *
 * 样式经 CSS 变量取色：theme-vars.css 随入口静态加载，基础配色立即可用；
 * 站点配置（主题、自定义 CSS 变量）由 main.ts 应用到文档后，本组件自动跟随
 * （background-color 带 transition，配置到达时平滑变色一次，不闪烁）。
 * 不依赖 FontAwesome / element-plus（它们属于动态加载的重资源）。
 */
</script>

<template>
  <div class="blog-app-loading">
    <div class="blog-app-loading-circle">
      <h2>LOADING</h2>
      <p>加载过慢请开启缓存</p>
      <div class="blog-app-loading-spinner"></div>
    </div>
  </div>
</template>

<style scoped>
.blog-app-loading {
  align-items: center;
  background: var(--blog-bg, #f6f8fa);
  display: flex;
  flex-direction: column;
  height: 100vh;
  justify-content: center;
  left: 0;
  position: fixed;
  top: 0;
  transition: background-color 0.3s;
  width: 100%;
  z-index: 2147483647;
}

.blog-app-loading-circle {
  align-items: center;
  border: 10px solid var(--blog-accent, #a3ddfb);
  border-radius: 50%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  height: 50vmin;
  justify-content: center;
  padding: 40px;
  text-align: center;
  width: 50vmin;
}

.blog-app-loading-circle h2 {
  color: var(--blog-text, #1e3e3f);
  margin: 10px 0;
}

.blog-app-loading-circle p {
  color: var(--blog-muted, #5c6b72);
  margin: 10px 0;
}

.blog-app-loading-spinner {
  animation: blog-app-loading-spin 1s linear infinite;
  border: 4px solid color-mix(in srgb, var(--blog-primary, #66afef) 30%, transparent);
  border-radius: 50%;
  border-top-color: var(--blog-primary, #66afef);
  box-sizing: border-box;
  height: 40px;
  width: 40px;
}

@keyframes blog-app-loading-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
