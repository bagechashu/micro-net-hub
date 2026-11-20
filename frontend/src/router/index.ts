import { createRouter, createWebHashHistory } from 'vue-router';
import type { RouteRecordRaw, RouteComponent } from 'vue-router';
import i18n from '@/i18n';

const t = i18n.global.t;

/* Layout */
const Layout = (): RouteComponent => import('@/layout/index.vue');

/**
 * constantRoutes
 * 所有人都能访问的基础页面；其余页面根据后端返回的菜单动态添加（store/modules/permission.ts）。
 *
 * 注意：hidden、alwaysShow 属性配置在 meta 中。
 */
export const constantRoutes: RouteRecordRaw[] = [
  {
    path: '/redirect',
    component: Layout,
    meta: { hidden: true },
    children: [
      {
        path: '/redirect/:path(.*)',
        component: () => import('@/views/redirect/index.vue')
      }
    ]
  },
  {
    // 找回密码：未登录也可访问，放在 Layout 里保持顶部导航和页脚
    path: '/changePass',
    component: Layout,
    meta: { hidden: true },
    children: [
      {
        path: '',
        name: 'changePass',
        component: () => import('@/views/changePassword/index.vue'),
        meta: { title: t('loginform.forgetPassword'), icon: 'user', noCache: true, hidden: true }
      }
    ]
  },
  {
    // 错误页：放在 Layout 里，保留顶部导航和页脚
    path: '/404',
    component: Layout,
    meta: { hidden: true },
    children: [
      {
        path: '',
        name: 'Page404',
        component: () => import('@/views/error-page/404.vue'),
        meta: { title: t('errorPage.notFoundTitle'), hidden: true, hidePageHeader: true }
      }
    ]
  },
  {
    path: '/401',
    component: Layout,
    meta: { hidden: true },
    children: [
      {
        path: '',
        name: 'Page401',
        component: () => import('@/views/error-page/401.vue'),
        meta: { title: t('errorPage.forbiddenTitle'), hidden: true, hidePageHeader: true }
      }
    ]
  },
  {
    path: '/',
    component: Layout,
    redirect: '/sitenav',
    children: [
      {
        path: '/sitenav',
        component: () => import('@/views/sitenav/index.vue'),
        name: 'SiteNav',
        meta: { title: t('sidebar.Sitenav'), icon: 'home', affix: true, hidden: false }
      }
    ]
  },
  {
    path: '/profile',
    component: Layout,
    redirect: '/profile/index',
    meta: { hidden: true },
    children: [
      {
        path: 'index',
        component: () => import('@/views/profile/index.vue'),
        name: 'Profile',
        meta: { title: t('profile.profile'), icon: 'user', noCache: true }
      }
    ]
  }
];

/** 未匹配的地址统一跳转到 404，需在动态路由之后添加 */
export const notFoundRoute: RouteRecordRaw = {
  path: '/:pathMatch(.*)*',
  name: 'NotFound',
  redirect: '/404',
  meta: { hidden: true }
};

const router = createRouter({
  // 后端以 /ui/ 托管静态页面，使用 hash 模式即可，无需额外配置服务端回退
  history: createWebHashHistory(import.meta.env.BASE_URL),
  scrollBehavior: () => ({ top: 0 }),
  routes: constantRoutes
});

/** addDynamicRoutes 添加的路由的移除函数，用于退出登录或切换用户时清理 */
let removeDynamicRoutes: Array<() => void> = [];

export function addDynamicRoutes(routes: RouteRecordRaw[]) {
  routes.forEach(route => {
    // 后端菜单里有些条目只是指向前端已有页面的入口（例如“个人中心” /profile/index，组件为 Layout、没有子菜单）。
    // vue-router 4+ 中后添加的同路径路由会覆盖已有路由，导致只渲染一个空布局；这类条目只保留在侧边栏，不再注册路由。
    if (!route.children?.length && router.resolve(route.path).matched.length > 0) {
      return;
    }
    removeDynamicRoutes.push(router.addRoute(route));
  });
  removeDynamicRoutes.push(router.addRoute(notFoundRoute));
}

/** 移除所有动态添加的路由，恢复到 constantRoutes */
export function resetRouter() {
  removeDynamicRoutes.forEach(remove => remove());
  removeDynamicRoutes = [];
}

export default router;
