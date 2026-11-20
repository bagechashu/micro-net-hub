<template>
  <!-- 通用错误页（Tabler error page 风格）：图标 + 状态码 + 标题 + 说明 + 操作 -->
  <div class="error-page">
    <span class="error-page__icon" :class="`is-${tone}`">
      <component :is="icon" :size="40" stroke="1.5" />
    </span>
    <div class="error-page__code">{{ code }}</div>
    <h1 class="error-page__title">{{ title }}</h1>
    <p class="error-page__desc">{{ desc }}</p>
    <div class="error-page__actions">
      <el-button :icon="ArrowLeft" @click="goBack">{{ $t('errorPage.back') }}</el-button>
      <el-button type="primary" :icon="Home" @click="router.push('/')">{{ $t('errorPage.home') }}</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue';
import { useRouter } from 'vue-router';
import { IconArrowLeft, IconHome } from '@tabler/icons-vue';

defineOptions({ name: 'ErrorPage' });

withDefaults(
  defineProps<{
    code: string;
    title: string;
    desc: string;
    icon: Component;
    tone?: 'primary' | 'warning';
  }>(),
  { tone: 'primary' }
);

const router = useRouter();
const ArrowLeft = IconArrowLeft;
const Home = IconHome;

// 没有上一页（直接打开或从外部进入）时回首页
function goBack() {
  if (window.history.state?.back) router.back();
  else router.push('/');
}
</script>

<style lang="scss" scoped>
.error-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 56px 16px 72px;
  text-align: center;
}

.error-page__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  margin-bottom: 20px;
  border-radius: 50%;

  &.is-primary {
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
  }

  &.is-warning {
    color: var(--el-color-warning);
    background: var(--el-color-warning-light-9);
  }
}

.error-page__code {
  font-size: 72px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0.04em;
  color: var(--el-text-color-placeholder);
}

.error-page__title {
  margin: 16px 0 8px;
  font-size: 22px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.error-page__desc {
  max-width: 460px;
  margin: 0 0 28px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.error-page__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
}
</style>
