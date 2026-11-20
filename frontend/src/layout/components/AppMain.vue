<template>
  <main class="page-wrapper">
    <!-- 页面自带大标题的（如 404）通过 meta.hidePageHeader 隐藏这里的标题区 -->
    <div v-if="!route.meta.hidePageHeader" class="app-container-xl page-header">
      <div class="page-header__titles">
        <div v-if="pretitle" class="page-pretitle">{{ pretitle }}</div>
        <h2 class="page-title">{{ title }}</h2>
      </div>
      <!-- 页面右上角操作区：页面通过 <Teleport to="#page-header-actions" defer> 放入内容 -->
      <div id="page-header-actions" class="page-header__actions" />
    </div>

    <div class="app-container-xl page-body">
      <router-view v-slot="{ Component, route: r }">
        <transition name="fade-transform" mode="out-in">
          <component :is="Component" :key="r.fullPath" />
        </transition>
      </router-view>
    </div>

    <footer class="app-container-xl page-footer">
      <span>Since 2024</span>
      <span class="page-footer__dot">·</span>
      <span>Powered by Open Source</span>
      <span class="page-footer__dot">·</span>
      <span>LICENSE</span>
      <a href="https://github.com/bagechashu/micro-net-hub/blob/main/LICENSE" target="_blank" rel="noopener">GPLv3</a>
    </footer>
  </main>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

defineOptions({ name: 'AppMain' });

const route = useRoute();
const { t } = useI18n();

const titled = computed(() => route.matched.filter(r => r.meta?.title));

// 页面大标题：当前页面；小标题：上级菜单（没有上级时显示“首页”）
const title = computed(() => (route.meta.title as string) || '');
const pretitle = computed(() => {
  const parents = titled.value.slice(0, -1);
  return parents.length ? (parents[parents.length - 1].meta.title as string) : t('common.index');
});
</script>

<style lang="scss" scoped>
.page-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.page-header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px 24px;
  padding-top: 24px;
  padding-bottom: 16px;
}

.page-header__actions {
  display: flex;
  align-items: center;
  gap: 8px;

  &:empty {
    display: none;
  }
}

.page-body {
  flex: 1;
  padding-bottom: 24px;
}

.page-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding-top: 16px;
  padding-bottom: 16px;
  font-size: 12px;
  color: var(--app-muted-text);
  border-top: 1px solid var(--app-header-border);
}

.page-footer__dot {
  opacity: 0.6;
}
</style>
