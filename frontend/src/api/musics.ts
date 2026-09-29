/**
 * 音乐相关接口（docs/api.md 第 13 章）
 *
 * - `GET    /api/musics`           公开音乐列表（仅启用，供前台 APlayer 播放）
 * - `GET    /api/admin/musics`     管理端音乐列表（分页，含禁用）
 * - `POST   /api/admin/musics`     创建音乐（音频文件先经 /admin/files 上传）
 * - `GET    /api/admin/musics/:id` 音乐详情
 * - `PUT    /api/admin/musics/:id` 更新音乐（更换音频时后端同步清理旧文件）
 * - `DELETE /api/admin/musics/:id` 删除音乐（同步删除关联音频文件记录与存储对象）
 */

import { request } from './http'
import type { Music, MusicQueryParams, MusicUpsertParams, PageResult, PublicMusic } from './types'

/** 公开音乐列表（docs/api.md 13.1，供 APlayer 播放） */
export function getPublicMusics() {
  return request<PublicMusic[]>({
    url: '/musics',
    method: 'get',
  })
}

/** 管理端音乐列表（docs/api.md 13.2，分页，可按歌名/歌手搜索） */
export function listMusics(params: MusicQueryParams) {
  return request<PageResult<Music>>({
    url: '/admin/musics',
    method: 'get',
    params,
  })
}

/** 创建音乐（docs/api.md 13.2） */
export function createMusic(data: MusicUpsertParams) {
  return request<Music>({
    url: '/admin/musics',
    method: 'post',
    data,
  })
}

/** 更新音乐（docs/api.md 13.2） */
export function updateMusic(id: number, data: MusicUpsertParams) {
  return request<Music>({
    url: `/admin/musics/${id}`,
    method: 'put',
    data,
  })
}

/** 删除音乐（docs/api.md 13.2，同步删除关联音频文件） */
export function deleteMusic(id: number) {
  return request<{ message: string }>({
    url: `/admin/musics/${id}`,
    method: 'delete',
  })
}
