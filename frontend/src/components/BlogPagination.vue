<script setup lang="ts">
import { computed } from 'vue'
import { buildPages } from '@/utils/pagination'

const props = withDefaults(
  defineProps<{
    total: number
    page: number
    pageSize?: number
    /** 当前页两侧展示的页码个数，默认 1 */
    siblingCount?: number
  }>(),
  {
    pageSize: 10,
    siblingCount: 1,
  },
)

const emit = defineEmits<{
  change: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const items = computed(() => buildPages(totalPages.value, props.page, props.siblingCount))

function goTo(page: number) {
  const next = Math.min(Math.max(page, 1), totalPages.value)
  if (next !== props.page) {
    emit('change', next)
  }
}
</script>

<template>
  <nav class="blog-pagination" aria-label="分页导航">
    <button
      type="button"
      class="blog-page-btn"
      :disabled="page <= 1"
      aria-label="上一页"
      @click="goTo(page - 1)"
    >
      <i class="fa-solid fa-angle-left" aria-hidden="true"></i>
      <span>上一页</span>
    </button>

    <template v-for="(item, index) in items" :key="index">
      <span v-if="item === 'ellipsis'" class="blog-page-ellipsis" aria-hidden="true">…</span>
      <button
        v-else
        type="button"
        class="blog-page-btn"
        :class="{ 'blog-page-current': item === page }"
        :aria-current="item === page ? 'page' : undefined"
        @click="goTo(item)"
      >
        {{ item }}
      </button>
    </template>

    <button
      type="button"
      class="blog-page-btn"
      :disabled="page >= totalPages"
      aria-label="下一页"
      @click="goTo(page + 1)"
    >
      <span>下一页</span>
      <i class="fa-solid fa-angle-right" aria-hidden="true"></i>
    </button>

    <span class="blog-page-total">共 {{ totalPages }} 页</span>
  </nav>
</template>
