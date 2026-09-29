/// <reference types="vite/client" />

/** 环境变量类型声明（供 src/api/http.ts 等使用） */
interface ImportMetaEnv {
  /** API 基础路径，默认 `/api`（配合 Vite 代理或反向代理使用） */
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
