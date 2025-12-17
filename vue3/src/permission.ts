import router from './router';
import userStore from './store/modules/user';
import permissionStore from './store/modules/permission';
import NProgress from 'nprogress'; // progress bar
import 'nprogress/nprogress.css'; // progress bar style
import { getToken } from '@/utils/auth'; // get token from cookie
import getPageTitle from '@/utils/get-page-title';
import { ElMessage } from 'element-plus';

NProgress.configure({ showSpinner: false }); // NProgress Configuration

router.beforeEach(async (to, from, next) => {
  // console.log('router.beforeEach', to.path, from.path);
  // start progress bar
  NProgress.start();

  // set page title
  document.title = getPageTitle(to.meta.title);

  // determine whether the user has logged in
  const hasToken = getToken();

  if (hasToken) {
    if (to.path === '/login') {
      // if is logged in, redirect to the home page
      NProgress.done(); // hack: https://github.com/PanJiaChen/vue-element-admin/pull/2939
      next({ path: '/' });
    } else {
      // determine whether the user has obtained his permission roles through getInfo
      const hasRoles = userStore().roles && userStore().roles.length > 0;
      if (hasRoles) {
        next();
      } else {
        try {
          // get user info
          // note: roles must be a object array! such as: ['admin'] or ,['developer','editor']
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const infoRes = await userStore().getInfo() as any;
          // console.log('current user:', infoRes);

          // generate accessible routes map based on backend menu for the user
          const accessRoutes = await permissionStore().generateRoutes(infoRes);
          console.log('accessRoutes=', accessRoutes);

          // dynamically add accessible routes
          // router.addRoutes(accessRoutes);
          accessRoutes.forEach(item => {
            router.addRoute(item);
          });
          // console.log('next=', accessRoutes);

          // hack method to ensure that addRoutes is complete
          // set the replace: true, so the navigation will not leave a history record
          next({ ...to, replace: true });
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } catch (error: any) {
          // 确保控制台能够打印错误信息
          console.error('Permission check error:', error);
          // remove token and go to login page to re-login
          await userStore().resetToken();
          ElMessage.error(error.message || 'Has Error');
          NProgress.done();
          next('/');
          // next(`/login?redirect=${to.path}`);
        }
      }
    }
  } else {
    await permissionStore().generateRoutesAnonymous();
    const whiteList = ['/', '/auth-redirect', '/sitenav', '/changePass']; // no redirect whitelist
    /* has no token*/
    if (whiteList.includes(to.path)) {
      // in the free login whitelist, go directly
      next();
    } else {
      // other pages that do not have permission to access are redirected to the login page.
      NProgress.done();
      next('/');
      // next(`/login?redirect=${to.path}`);
    }
  }
});

router.afterEach(() => {
  // finish progress bar
  NProgress.done();
});
