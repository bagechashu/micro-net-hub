<template>
  <nav class="app-nav">
    <div class="app-container-xl">
      <!-- 菜单项变化（登录/登出后）时重新挂载：el-menu 只在挂载和尺寸变化时计算折叠，否则登录后可能只剩“•••” -->
      <el-menu
        :key="menuKey"
        mode="horizontal"
        :default-active="activeMenu"
        :ellipsis="true"
        router
        class="app-nav__menu">
        <template v-for="item in navItems" :key="item.path">
          <el-sub-menu v-if="item.children" :index="item.path" popper-class="app-nav__popper">
            <template #title>
              <nav-icon :name="item.icon" />
              <span>{{ item.title }}</span>
            </template>
            <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path">
              <nav-icon :name="child.icon" />
              <span>{{ child.title }}</span>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="item.path">
            <nav-icon :name="item.icon" />
            <span>{{ item.title }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
import usePermissionStore from '@/store/modules/permission';
import { constantRoutes } from '@/router';
import NavIcon from './NavIcon.vue';

defineOptions({ name: 'AppNav' });

interface NavItem {
  path: string;
  title: string;
  icon?: string;
  children?: NavItem[];
}

const route = useRoute();
const permission = usePermissionStore();

function joinPath(base: string, path: string): string {
  if (path.startsWith('/')) return path;
  return `${base.replace(/\/$/, '')}/${path}`;
}

// 隐藏的基础路由（如 /profile），后端菜单里指向它们的入口不在导航栏重复显示（个人中心在右上角用户菜单里）
const hiddenPrefixes = constantRoutes
  .filter(r => r.meta?.hidden)
  .map(r => r.path)
  .filter(p => p !== '/');

function isHidden(r: RouteRecordRaw, fullPath: string) {
  return !!r.meta?.hidden || hiddenPrefixes.some(p => fullPath === p || fullPath.startsWith(`${p}/`));
}

function toItem(r: RouteRecordRaw, base = ''): NavItem | null {
  const fullPath = joinPath(base, r.path);
  if (isHidden(r, fullPath)) return null;

  const children = (r.children || [])
    .map(c => toItem(c, fullPath))
    .filter((c): c is NavItem => !!c);

  if (r.children?.length) {
    if (!children.length) return null;
    // 没有标题的外层布局（如首页 "/"）只有一个子页面时，直接显示该子页面
    if (!r.meta?.title && children.length === 1) return children[0];
    return { path: fullPath, title: r.meta?.title as string, icon: r.meta?.icon as string, children };
  }
  if (!r.meta?.title) return null;
  return { path: fullPath, title: r.meta.title as string, icon: r.meta.icon as string };
}

const navItems = computed(() => permission.routes
  .map(r => toItem(r))
  .filter((i): i is NavItem => !!i));

const menuKey = computed(() => navItems.value.map(i => `${i.path}:${i.title}`).join('|'));

const activeMenu = computed(() => (route.meta.activeMenu as string) || route.path);

</script>

<style lang="scss" scoped>
.app-nav {
  background: var(--app-header-bg);
  border-bottom: 1px solid var(--app-header-border);
}

.app-nav__menu {
  --el-menu-bg-color: transparent;
  --el-menu-hover-bg-color: transparent;
  --el-menu-text-color: var(--app-nav-text);
  --el-menu-hover-text-color: var(--el-text-color-primary);
  --el-menu-active-color: var(--el-color-primary);
  --el-menu-horizontal-height: 52px;
  --el-menu-item-font-size: 14px;
  border-bottom: none;
  margin-left: -12px;

  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    padding: 0 12px;
    font-weight: 500;
  }

  :deep(.el-sub-menu__title) {
    padding-right: 30px;
  }

  :deep(.el-sub-menu .el-sub-menu__icon-arrow) {
    right: 10px;
    margin-top: -5px;
  }

  :deep(.el-menu-item.is-active),
  :deep(.el-sub-menu.is-active .el-sub-menu__title) {
    border-bottom-width: 2px;
  }
}

</style>

<style lang="scss">
// 下拉子菜单（挂在 body 上，不能用 scoped）
.app-nav__popper {
  .el-menu {
    padding: 4px;
    border-radius: 8px;
  }

  .el-menu-item {
    height: 36px;
    line-height: 36px;
    border-radius: 6px;
  }

}
</style>
