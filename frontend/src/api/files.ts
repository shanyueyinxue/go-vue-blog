/**
 * 文件上传相关接口（docs/api.md 第 10 章，均需 JWT）
 *
 * - `POST   /api/admin/files`        上传文件（multipart/form-data）
 * - `GET    /api/admin/files`        文件列表（分页）
 * - `DELETE /api/admin/files/:id`    删除文件（清理本地/OSS 存储）
 */

import { request } from './http'
import type { FileItem, FileListParams, PageResult, UploadResult } from './types'

/**
 * 上传文件（docs/api.md 10.1，`multipart/form-data`）。
 *
 * @param file 待上传文件（必填，单文件）
 * @param dir  存储目录（可选，如 `posts`）
 */
export function uploadFile(file: File, dir?: string) {
  const formData = new FormData()
  formData.append('file', file)
  if (dir) {
    formData.append('dir', dir)
  }
  return request<UploadResult>({
    url: '/admin/files',
    method: 'post',
    data: formData,
    // 由浏览器自动生成 boundary，无需手动指定 Content-Type
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 文件列表（docs/api.md 10.2，分页，可按 MIME 类型筛选） */
export function listFiles(params: FileListParams) {
  return request<PageResult<FileItem>>({
    url: '/admin/files',
    method: 'get',
    params,
  })
}

/** 删除文件（docs/api.md 10.2，物理删除记录并清理存储） */
export function deleteFile(id: number) {
  return request<{ message: string }>({
    url: `/admin/files/${id}`,
    method: 'delete',
  })
}
