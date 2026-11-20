<template>
  <div class="app-layout">
    <AppHeader />
    <AppNav />
    <AppMain />
  </div>
</template>

<script setup lang="ts">
import AppHeader from './components/AppHeader.vue';
import AppNav from './components/AppNav.vue';
import AppMain from './components/AppMain.vue';

defineOptions({ name: 'LayoutIndex' });
</script>

<style lang="scss">
.app-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--el-bg-color-page);
}

// 居中限宽容器（Tabler 的 container-xl）
.app-container-xl {
  width: 100%;
  max-width: var(--app-container-width);
  margin: 0 auto;
  padding-left: 24px;
  padding-right: 24px;
  box-sizing: border-box;

  @media (max-width: 768px) {
    padding-left: 12px;
    padding-right: 12px;
  }
}
</style>
