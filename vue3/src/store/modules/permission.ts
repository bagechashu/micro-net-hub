import { defineStore } from 'pinia';
import { asyncRoutes, constantRoutes } from '@/router';
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
  routes: Array<RouteRecordRaw>;
  addRoutes: Array<RouteRecordRaw>;
}

/**
 * Use meta.role to determine if the current user has permission
 * @param roles
 * @param route
 */
function hasPermission(roles:string[], route:RouteRecordRaw):boolean {
  if (route.meta && route.meta.roles) {
    const rolesArr = route.meta.roles as string[];
    return roles.some(role => rolesArr.includes(role));
  } else {
    return true;
  }
}

/**
 * Filter asynchronous routing tables by recursion
 * @param routes asyncRoutes
 * @param roles
 */
export function filterAsyncRoutes(routes:RouteRecordRaw[], roles: string[]): Array<RouteRecordRaw> {
  const res:Array<RouteRecordRaw> = [];

  routes.forEach(route => {
    const tmp = { ...route };
    if (hasPermission(roles, tmp)) {
      if (tmp.children) {
        tmp.children = filterAsyncRoutes(tmp.children, roles);
      }
      res.push(tmp);
    }
  });

  return res;
}

export default defineStore({
  id: 'permission',
  state: ():IPermissionState => ({
    routes: [],
    addRoutes: []
  }),
  getters: {},
  actions: {
    setRoutes(routes: RouteRecordRaw[]) {
      this.addRoutes = routes;
      this.routes = constantRoutes.concat(routes);
    },
    // Convert backend menu tree to vue-router RouteRecordRaw
    async generateRoutesFromMenu(userinfo: { id: string | number }) {
      try {
        const { data } = await getUserMenuTreeByUserId({ id: userinfo.id });
        const menuTree = data;
        const accessedRoutes = getRoutesFromMenuTree(menuTree);
        this.setRoutes(accessedRoutes);
        return accessedRoutes;
      } catch (err) {
        return Promise.reject(err);
      }
    },
    generateRoutes(roles: string[]) {
      let accessedRoutes;
      if (roles.includes('admin')) {
        accessedRoutes = asyncRoutes || [];
      } else {
        accessedRoutes = filterAsyncRoutes(asyncRoutes, roles);
      }
      this.setRoutes(accessedRoutes);
      return accessedRoutes;
    }
  }
});

/**
 * Helpers to transform backend menu data to RouteRecordRaw
 */
export function getRoutesFromMenuTree(menuTree: MenuItem[]): RouteRecordRaw[] {
  const routes: RouteRecordRaw[] = [];
  menuTree.forEach(menu => {
    let childrenRoutes: RouteRecordRaw[] | undefined;
    if (menu.children && menu.children.length > 0) {
      childrenRoutes = getRoutesFromMenuTree(menu.children);
    }
    const route: RouteRecordRaw = {
      path: menu.path as string,
      name: menu.name as string,
      component: loadComponent(menu.component as string),
      hidden: menu.hidden === 1,
      redirect: menu.redirect || undefined,
      meta: {
        name: menu.name,
        title: menu.title || menu.name,
        icon: menu.icon,
        noCache: menu.noCache === 1,
        breadcrumb: menu.breadcrumb === 1,
        activeMenu: menu.activeMenu
      }
    } as RouteRecordRaw;
    if (childrenRoutes && childrenRoutes.length > 0) {
      route.children = childrenRoutes;
    }
    routes.push(route);
  });
  return routes;
}

export function loadComponent(component:string) {
  if (!component || component === '' || component === 'Layout') {
    return Layout;
  }
  // component is expected like '/system/user/index'
  return () => import(`@/views${component}.vue`);
}
