<template>
  <component :is="tablerIcon" v-if="tablerIcon" class="nav-icon" />
  <svg-icon v-else-if="name && !name.startsWith('el-icon')" :icon-class="name" class-name="nav-icon" />
</template>

<script setup lang="ts">
import { computed, h } from 'vue';
import { menuIcon } from '@/icons/tabler';

defineOptions({ name: 'NavIcon' });

const props = defineProps<{ name?: string }>();

// 后端菜单存的是旧图标名，优先映射为 Tabler 图标；没有映射的仍用 svg 精灵图标
const tablerIcon = computed(() => {
  const icon = menuIcon(props.name);
  return icon ? () => h(icon as any, { size: 18, stroke: 1.75 }) : undefined;
});
</script>

<style scoped>
.nav-icon {
  flex: none;
  width: 18px;
  height: 18px;
  margin-right: 8px;
  color: var(--app-muted-text);
}

:global(.el-menu-item.is-active .nav-icon),
:global(.el-sub-menu.is-active > .el-sub-menu__title .nav-icon) {
  color: var(--el-color-primary);
}
</style>
