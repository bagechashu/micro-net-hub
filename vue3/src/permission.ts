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

router.afterEach((to) => {
  NProgress.done();
  try {
    sessionStorage.removeItem(chunkReloadKey(to.fullPath));
  } catch {
    // 忽略存储不可用
  }
});

// 页面模块（懒加载的 .vue / chunk）加载失败时，路由导航会被中断，表现为“点了没反应、进度条卡住”。
// 常见原因：发布新版本后旧页面去加载已不存在的旧 chunk；开发时 Vite 依赖预构建过期（504 Outdated Optimize Dep）。
// 处理：对同一目标页自动整页刷新一次；仍然失败则提示错误，避免无限刷新。
const CHUNK_ERROR = /Failed to fetch dynamically imported module|Importing a module script failed|error loading dynamically imported module/i;

function chunkReloadKey(path: string) {
  return `chunk-reload:${path}`;
}

router.onError((error, to) => {
  NProgress.done();
  if (!CHUNK_ERROR.test(String((error as Error)?.message ?? error))) {
    return;
  }
  const key = chunkReloadKey(to.fullPath);
  let reloaded: boolean;
  try {
    reloaded = sessionStorage.getItem(key) === '1';
    sessionStorage.setItem(key, '1');
  } catch {
    // 无法记录时按“已重试过”处理，避免循环刷新
    reloaded = true;
  }
  if (!reloaded) {
    window.location.href = router.resolve(to).href;
    window.location.reload();
  } else {
    ElMessage.error('页面加载失败，请刷新后重试');
  }
});
