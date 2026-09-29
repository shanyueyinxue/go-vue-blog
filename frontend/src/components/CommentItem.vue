<script setup lang="ts">
import { computed } from 'vue'
import type { Comment } from '@/api'
import { formatDate } from '@/utils/format'

const props = withDefaults(
  defineProps<{
    comment: Comment
    /** 当前层级：1 为顶层评论，回复每深一层 +1 */
    depth?: number
  }>(),
  {
    depth: 1,
  },
)

const emit = defineEmits<{
  reply: [comment: Comment]
}>()

// 最多嵌套 3 层（顶层 1 + 回复 2 层缩进）；超过 3 层的回复平铺展示，不再缩进
const isNested = computed(() => props.depth > 1 && props.depth <= 3)
const isFlat = computed(() => props.depth > 3)
</script>

<template>
  <div
    class="blog-comment"
    :class="[
      isNested ? 'reply' : '',
      props.depth === 3 ? 'depth-3' : '',
      isFlat ? 'depth-flat' : '',
    ]"
  >
    <div class="blog-comment-head">
      <span class="blog-comment-nick">{{ comment.nickname }}</span>
      <span v-if="comment.is_author === 1" class="badge">博主</span>
      <span class="blog-comment-time">{{ formatDate(comment.created_at) }}</span>
      <button type="button" class="btn btn-sm" @click="emit('reply', comment)">回复</button>
    </div>
    <p class="blog-comment-content">{{ comment.content }}</p>

    <!-- 递归渲染楼中楼子评论；超过 3 层的回复平铺，不再继续缩进 -->
    <div v-if="comment.replies?.length" class="blog-comment-replies">
      <CommentItem
        v-for="child in comment.replies"
        :key="child.id"
        :comment="child"
        :depth="props.depth + 1"
        @reply="emit('reply', $event)"
      />
    </div>
  </div>
</template>
