<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { PostListItem } from '@/api'
import { formatDate } from '@/utils/format'

defineProps<{ post: PostListItem }>()
</script>

<template>
  <article class="blog-post">
    <RouterLink :to="`/post/${post.slug}`">
      <h2 class="blog-post-title">
        {{ post.title }}
        <span v-if="post.is_top === 1" style="margin-left: 8px">
          <i class="fa-solid fa-thumbtack"></i>
        </span>
      </h2>
    </RouterLink>

    <div class="blog-post-meta">
      <RouterLink
        v-if="post.category"
        :to="{ path: '/posts', query: { category_id: post.category.id } }"
        class="blog-post-category"
        :title="`查看「${post.category.name}」分类的文章`"
      >
        <i class="fa-solid fa-bookmark fa-fw"></i>
        {{ post.category.name }}
      </RouterLink>
      <span>
        <i class="fa-solid fa-calendar fa-fw"></i>
        {{ formatDate(post.published_at) }}
      </span>
      <span>
        <i class="fa-solid fa-eye fa-fw"></i>
        {{ post.view_count }}
      </span>
      <span>
        <i class="fa-solid fa-comment fa-fw"></i>
        {{ post.comment_count }}
      </span>
    </div>

    <div class="blog-post-excerpt">{{ post.excerpt || '暂无摘要' }}</div>

    <div v-if="post.tags.length" class="blog-post-tags">
      <span class="blog-tag">
        <i class="fa-solid fa-tags fa-fw"></i>
      </span>
      <RouterLink
        v-for="tag in post.tags"
        :key="tag.id"
        :to="{ path: '/posts', query: { tag_id: tag.id } }"
        class="blog-tag"
      >
        {{ tag.name }}
      </RouterLink>
    </div>

    <RouterLink :to="`/post/${post.slug}`" class="blog-go-post">阅读全文</RouterLink>
  </article>
</template>
