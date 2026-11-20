import { defineStore } from 'pinia';
import { constantRoutes } from '@/router';
import { getUserMenuTreeByUserId } from '@/api/system/menu';
import i18n from '@/i18n';
import Layout from '@/layout/index.vue';
import type { RouteRecordRaw } from 'vue-router';

/** 后端菜单（backend/internal/module/account/model/menu.go）。布尔类字段：1 表示“是”，2 表示“否” */
interface MenuItem {
  name?: string;
  title?: string;
  icon?: string;
  path?: string;
  redirect?: string;
  component?: string;
  hidden?: number;
  noCache?: number;
  alwaysShow?: number;
  breadcrumb?: number;
  activeMenu?: string;
  children?: MenuItem[];
}

interface IPermissionState {
  routes: RouteRecordRaw[];
  addRoutes: RouteRecordRaw[];
}

const usePermissionStore = defineStore('permission', {
  state: (): IPermissionState => ({
    routes: [],
    addRoutes: []
  }),

  actions: {
    setRoutes(routes: RouteRecordRaw[]) {
      this.addRoutes = routes;
      this.routes = constantRoutes.concat(routes);
    },

    async generateRoutes(userinfo: { ID: string | number }) {
      const { data } = await getUserMenuTreeByUserId({ id: userinfo.ID });
      const accessedRoutes = buildRoutesFromMenu(data || []);
      this.setRoutes(accessedRoutes);
      return accessedRoutes;
    },

    generateRoutesAnonymous() {
      this.setRoutes([]);
    }
  }
});

export default usePermissionStore;

export function buildRoutesFromMenu(menuTree: MenuItem[], isRoot = true): RouteRecordRaw[] {
  return menuTree
    .filter(m => !!m.path)
    .map(menu => {
      const children = menu.children?.length
        ? buildRoutesFromMenu(menu.children, false)
        : undefined;

      const route = {
        // 只有顶层路由需要以 "/" 开头；子路由保持相对路径，否则会丢掉父级前缀
        path: isRoot ? normalizePath(menu.path!) : menu.path!,
        name: buildRouteName(menu),
        component: resolveComponent(menu.component),
        redirect: menu.redirect || undefined,
        meta: {
          name: menu.name,
          title: menuTitle(menu),
          icon: menu.icon,
          hidden: menu.hidden === 1,
          noCache: menu.noCache === 1,
          // 与 Vue2 版本一致：有子菜单的父菜单总是以子菜单形式显示（即使只有一个子项）
          alwaysShow: menu.alwaysShow === 1 || !!children?.length,
          // 后端：1 显示（默认），2 隐藏
          breadcrumb: menu.breadcrumb !== 2,
          activeMenu: menu.activeMenu || undefined
        }
      } as RouteRecordRaw;

      if (children?.length) {
        route.children = children;
      }

      return route;
    });
}

/* -------------------------- Helpers -------------------------- */
const viewModules = import.meta.glob('@/views/**/*.vue');

function resolveComponent(component?: string) {
  if (!component || component === 'Layout') {
    return Layout;
  }

  const viewPath = `/src/views${component}.vue`;
  const loader = viewModules[viewPath];

  if (!loader) {
    console.error(`[Router] View component not found: ${viewPath}`);
    return () => import('@/views/error-page/404.vue');
  }

  return loader;
}

function menuTitle(menu: MenuItem): string {
  const key = `sidebar.${menu.name}`;
  const { t, te } = i18n.global;
  if (menu.name && te(key)) return t(key);
  return menu.title || menu.name || '';
}

function buildRouteName(menu: MenuItem): string {
  if (menu.name) return menu.name;
  return `route_${menu.path?.replace(/\//g, '_')}`;
}

function normalizePath(path: string): string {
  return path.startsWith('/') ? path : `/${path}`;
}
