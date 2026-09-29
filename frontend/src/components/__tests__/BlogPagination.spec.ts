import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import BlogPagination from '@/components/BlogPagination.vue'
import { buildPages } from '@/utils/pagination'

describe('buildPages 分页按钮序列', () => {
  it('总页数 <= 0 时返回空数组', () => {
    expect(buildPages(0, 1)).toEqual([])
    expect(buildPages(-1, 1)).toEqual([])
  })

  it('页码不超过 7 时全部展示，不出现省略号', () => {
    expect(buildPages(7, 4)).toEqual([1, 2, 3, 4, 5, 6, 7])
  })

  it('中间页：两侧各展示 1 个页码，空隙用省略号', () => {
    expect(buildPages(10, 5)).toEqual([1, 'ellipsis', 4, 5, 6, 'ellipsis', 10])
  })

  it('首页附近：不出现多余省略号', () => {
    expect(buildPages(10, 1)).toEqual([1, 2, 'ellipsis', 10])
    expect(buildPages(10, 2)).toEqual([1, 2, 3, 'ellipsis', 10])
  })

  it('末页附近：不出现多余省略号', () => {
    expect(buildPages(10, 9)).toEqual([1, 'ellipsis', 8, 9, 10])
    expect(buildPages(10, 10)).toEqual([1, 'ellipsis', 9, 10])
  })

  it('当前页越界时夹紧到有效范围', () => {
    expect(buildPages(10, 0)).toEqual([1, 2, 'ellipsis', 10])
    expect(buildPages(10, 99)).toEqual([1, 'ellipsis', 9, 10])
  })

  it('siblingCount = 2 时展示更多相邻页码', () => {
    expect(buildPages(20, 10, 2)).toEqual([
      1,
      'ellipsis',
      8,
      9,
      10,
      11,
      12,
      'ellipsis',
      20,
    ])
  })

  it('非法 siblingCount 按 0 处理', () => {
    expect(buildPages(20, 10, -1)).toEqual([1, 'ellipsis', 10, 'ellipsis', 20])
  })
})

describe('BlogPagination 组件', () => {
  const mountPagination = (props: { total: number; page: number; pageSize?: number }) =>
    mount(BlogPagination, { props })

  it('渲染上一页/下一页与页码按钮，当前页高亮', () => {
    const wrapper = mountPagination({ total: 100, page: 5 })
    const texts = wrapper.findAll('.blog-page-btn').map((btn) => btn.text())
    expect(texts).toContain('上一页')
    expect(texts).toContain('下一页')
    // 页码按钮：1 … 4 5 6 … 10
    expect(texts).toContain('1')
    expect(texts).toContain('4')
    expect(texts).toContain('5')
    expect(texts).toContain('6')
    expect(texts).toContain('10')

    const current = wrapper.find('.blog-page-current')
    expect(current.text()).toBe('5')
    expect(current.attributes('aria-current')).toBe('page')
    expect(wrapper.findAll('.blog-page-ellipsis')).toHaveLength(2)
  })

  it('第一页时上一页禁用，最后一页时下一页禁用', () => {
    const first = mountPagination({ total: 100, page: 1 })
    const prev = first.findAll('.blog-page-btn').find((b) => b.text().includes('上一页'))
    expect(prev?.attributes('disabled')).toBeDefined()

    const last = mountPagination({ total: 100, page: 10 })
    const next = last.findAll('.blog-page-btn').find((b) => b.text().includes('下一页'))
    expect(next?.attributes('disabled')).toBeDefined()
  })

  it('点击页码按钮触发 change 事件', async () => {
    const wrapper = mountPagination({ total: 100, page: 5 })
    const page6 = wrapper.findAll('.blog-page-btn').find((b) => b.text() === '6')
    await page6?.trigger('click')
    expect(wrapper.emitted('change')?.[0]).toEqual([6])
  })

  it('点击上一页/下一页触发 change 事件', async () => {
    const wrapper = mountPagination({ total: 100, page: 5 })
    const prev = wrapper.findAll('.blog-page-btn').find((b) => b.text().includes('上一页'))
    await prev?.trigger('click')
    expect(wrapper.emitted('change')?.[0]).toEqual([4])

    const next = wrapper.findAll('.blog-page-btn').find((b) => b.text().includes('下一页'))
    await next?.trigger('click')
    expect(wrapper.emitted('change')?.[1]).toEqual([6])
  })

  it('点击当前页不触发 change', async () => {
    const wrapper = mountPagination({ total: 100, page: 5 })
    await wrapper.find('.blog-page-current').trigger('click')
    expect(wrapper.emitted('change')).toBeUndefined()
  })
})
