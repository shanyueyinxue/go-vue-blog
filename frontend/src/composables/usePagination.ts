/**
 * 列表分页组合式函数
 *
 * 统一管理后台列表的分页状态（PageResult）、加载中标记、翻页与刷新，
 * 消除各管理页重复的 `result`/`loading`/`load`/`changePage` 样板代码。
 *
 * @example
 * ```ts
 * const { result, loading, load, changePage, reload } = usePagination<Tag>((p) => listAdminTags(p), 20)
 * // 带筛选条件时把参数包在 fetcher 闭包里：
 * const { result, loading, reload } = usePagination<Post>((p) =>
 *   listAdminPosts({ ...p, status: status.value }),
 * )
 * ```
 */
import { ref } from 'vue'
import type { PageResult } from '@/api'

export function usePagination<T>(
  fetcher: (params: { page: number; pageSize: number }) => Promise<PageResult<T>>,
  pageSize = 10,
) {
  /** 分页结果（page/pageSize/total 均由后端返回，翻页时更新 page） */
  const result = ref<PageResult<T>>({ list: [], total: 0, page: 1, pageSize })
  /** 加载中标记（用于 v-loading） */
  const loading = ref(false)

  /** 按当前页码加载数据 */
  async function load() {
    loading.value = true
    try {
      result.value = await fetcher({ page: result.value.page, pageSize: result.value.pageSize })
    } finally {
      loading.value = false
    }
  }

  /** 翻页：更新页码并重新加载（供 el-pagination @current-change 使用） */
  function changePage(page: number) {
    if (page < 1) return
    result.value.page = page
    load()
  }

  /** 回到第 1 页并重新加载（搜索条件变化、增删改后调用） */
  function reload() {
    result.value.page = 1
    load()
  }

  return { result, loading, load, changePage, reload }
}
