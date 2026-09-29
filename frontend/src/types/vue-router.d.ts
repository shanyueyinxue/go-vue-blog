import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    /** 页面标题（不含站点名，拼接逻辑见 App.vue） */
    title?: string
    /** 页面描述，缺省时回退站点级 description */
    description?: string
  }
}
