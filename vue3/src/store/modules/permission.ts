import { defineStore } from 'pinia';
import { constantRoutes } from '@/router';
import { getUserMenuTreeByUserId } from '@/api/system/menu';
import Layout from '@/layout/index.vue';
import type { RouteRecordRaw } from 'vue-router';

interface MenuItem {
  path?: string;
  name?: string;
  component?: string;
  children?: MenuItem[];
  hidden?: number;
  redirect?: string;
  title?: string;
  icon?: string;
  noCache?: number;
  breadcrumb?: number;
  activeMenu?: string;
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

    async generateRoutesAnonymous() {
      this.setRoutes([]);
      return;
    }
  }
});

export default usePermissionStore;

export function buildRoutesFromMenu(menuTree: MenuItem[]): RouteRecordRaw[] {
  return menuTree
    .filter(m => !!m.path)
    .map(menu => {
      const children = menu.children?.length
        ? buildRoutesFromMenu(menu.children)
        : undefined;

      const route: RouteRecordRaw = {
        path: normalizePath(menu.path!),
        name: buildRouteName(menu),
        component: resolveComponent(menu.component),
        redirect: menu.redirect || undefined,
        meta: {
          title: menu.title || menu.name,
          icon: menu.icon,
          hidden: menu.hidden === 1,
          noCache: menu.noCache === 1,
          breadcrumb: menu.breadcrumb !== 0,
          activeMenu: menu.activeMenu
        }
      } as RouteRecordRaw;

      if (children?.length) {
        route.children = children;
      }

      return route;
    });
}

/* -------------------------- Helpers -------------------------- */
// permission.ts (top-level)
const viewModules = import.meta.glob('@/views/**/*.vue');

function resolveComponent(component?: string) {
  if (!component || component === '' || component === 'Layout') {
    return Layout;
  }

  const viewPath = `/src/views${component}.vue`;
  const loader = viewModules[viewPath];

  if (!loader) {
    console.error(
      `[Router] View component not found: ${viewPath}`
    );
    return Layout; // or NotFound component
  }

  return loader;
}

function buildRouteName(menu: MenuItem): string {
  if (menu.name) return menu.name;
  return `route_${menu.path?.replace(/\//g, '_')}`;
}

function normalizePath(path: string): string {
  if (!path.startsWith('/')) {
    return `/${path}`;
  }
  return path;
}
