import router, { addDynamicRoutes } from './router';
import userStore from './store/modules/user';
import permissionStore from './store/modules/permission';
import NProgress from 'nprogress'; // progress bar
import 'nprogress/nprogress.css'; // progress bar style
import { getToken } from '@/utils/auth'; // get token from cookie
import getPageTitle from '@/utils/get-page-title';
import { ElMessage } from 'element-plus';

NProgress.configure({ showSpinner: false }); // NProgress Configuration

// 未登录也能访问的页面
const whiteList = ['/', '/sitenav', '/changePass', '/404', '/401'];

router.beforeEach(async (to) => {
  NProgress.start();
  document.title = getPageTitle(to.meta.title as string | undefined);

  if (!getToken()) {
    permissionStore().generateRoutesAnonymous();
    if (whiteList.includes(to.path)) {
      return true;
    }
    NProgress.done();
    return '/';
  }

  const user = userStore();
  if (user.roles && user.roles.length > 0) {
    return true;
  }

  try {
    // 首次进入或刷新页面：拉取用户信息和菜单，生成动态路由
    const info = await user.getInfo();
    const accessRoutes = await permissionStore().generateRoutes(info);
    addDynamicRoutes(accessRoutes);
    // 动态路由刚添加，需要重新解析一次目标地址；replace 避免留下多余的历史记录
    return { ...to, replace: true };
  } catch (error) {
    console.error('Permission check error:', error);
    // 令牌无效等情况：清掉登录状态，回到首页重新登录
    user.resetToken();
    ElMessage.error(`获取用户信息失败: ${(error as Error)?.message || error || '未知错误'}`);
    NProgress.done();
    return '/';
  }
});

router.afterEach(() => {
  NProgress.done();
});
