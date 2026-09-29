/**
 * 分页按钮序列工具
 *
 * 生成按钮式分页的页码序列：始终包含首页与末页，当前页前后各展示
 * `siblingCount` 个页码，页码之间的空隙用省略号占位。
 *
 * 例如 totalPages = 10、current = 5、siblingCount = 1 时：
 * `[1, 'ellipsis', 4, 5, 6, 'ellipsis', 10]`，渲染为 `1 … 4 5 6 … 10`。
 */

/** 分页按钮项：页码或省略号占位 */
export type PageItem = number | 'ellipsis'

/** 页码不超过该值时直接展示全部页码，不出现省略号 */
const SHOW_ALL_THRESHOLD = 7

/**
 * 生成分页按钮序列
 *
 * @param totalPages 总页数（>= 1）
 * @param current 当前页（越界时会被夹紧到 [1, totalPages]）
 * @param siblingCount 当前页两侧展示的页码个数，默认 1
 */
export function buildPages(totalPages: number, current: number, siblingCount = 1): PageItem[] {
  if (!Number.isInteger(totalPages) || totalPages <= 0) return []
  if (!Number.isInteger(siblingCount) || siblingCount < 0) siblingCount = 0

  const page = Math.min(Math.max(Math.trunc(current), 1), totalPages)

  // 页码不多时全部展示，避免出现 1 … 3 4 5 … 7 这种反而更占空间的省略
  if (totalPages <= SHOW_ALL_THRESHOLD) {
    return Array.from({ length: totalPages }, (_, i) => i + 1)
  }

  // 首页、末页 + 当前页附近页码（去重）
  const pages = new Set<number>([1, totalPages])
  const start = Math.max(2, page - siblingCount)
  const end = Math.min(totalPages - 1, page + siblingCount)
  for (let p = start; p <= end; p++) pages.add(p)

  const sorted = [...pages].sort((a, b) => a - b)
  const items: PageItem[] = []
  let prev = 0
  for (const p of sorted) {
    if (p - prev > 1) items.push('ellipsis')
    items.push(p)
    prev = p
  }
  return items
}
