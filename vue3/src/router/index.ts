// import { markRaw } from 'vue';
import { createRouter, createWebHashHistory } from 'vue-router'; // createWebHashHistory, createWebHistory
import type { Router, RouteRecordRaw, RouteComponent } from 'vue-router';

/* Layout */
const Layout = ():RouteComponent => import('@/layout/index.vue');

/**
 * constantRoutes
 * a base page that does not have permission requirements
 * all roles can be accessed
 *
 * 注意：hidden、alwaysShow 属性配置移动到了meta中！！！
 */
export const constantRoutes:RouteRecordRaw[] = [
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
  // {
  //   path: '/login',
  //   component: () => import('@/views/login/index.vue'),
  //   meta: { hidden: true }
  // },
  {
    path: '/auth-redirect',
    component: () => import('@/views/login/auth-redirect.vue'),
    meta: { hidden: true }
  },
  {
    path: '/404',
    component: () => import('@/views/error-page/404.vue'),
    meta: { hidden: true }
  },
  {
    path: '/401',
    component: () => import('@/views/error-page/401.vue'),
    meta: { hidden: true }
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
        meta: { title: '网站导航', icon: 'tree', affix: true, hidden: false }
      }
    ]
  },

  {
    path: '/changePass',
    name: 'changePass',
    component: () => import('@/views/changePassword/index.vue'),
    meta: { title: '忘记密码', icon: 'user', noCache: true, hidden: true }
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
        meta: { title: '个人中心', icon: 'user', noCache: true }
      }
    ]
  }
];

console.log('BASE_URL=', import.meta.env);

const createTheRouter = ():Router => createRouter({
  // history: createWebHashHistory(import.meta.env.BASE_URL),
  // 注意，如果要配置 HTML5 模式，则需要修改nginx配置，参考资料：
  // https://router.vuejs.org/zh/guide/essentials/history-mode.html
  history: createWebHashHistory(import.meta.env.BASE_URL),
  scrollBehavior: () => ({ top: 0 }),
  routes: constantRoutes
});

interface RouterPro extends Router {
  matcher: unknown;
}

const router = createTheRouter() as RouterPro;

// Detail see: https://github.com/vuejs/vue-router/issues/1234#issuecomment-357941465
export function resetRouter() {
  // router.clearRoutes(); RangeError: Maximum call stack size exceeded
  const newRouter = createTheRouter() as RouterPro;
  router.matcher = newRouter.matcher; // reset router
}

export default router;
