/**
 * 文章详情封面背景 URL（模块级共享状态）
 *
 * PostDetailView 加载到文章后写入 cover_image，卸载/切换文章时清空；
 * SiteLayout 据此把封面图设置为 .blog-shell 的背景
 * （extra.post_cover_enabled 开启时生效，background-attachment: fixed 固定铺满视口）。
 */
import { ref } from 'vue'

const postCover = ref('')

export function usePostCover() {
  return postCover
}
