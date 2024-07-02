import router from "./router";
import store from "./store";
import { Message } from "element-ui";
import NProgress from "nprogress"; // progress bar
import "nprogress/nprogress.css"; // progress bar style
import { getToken } from "@/utils/auth"; // get token from cookie
import getPageTitle from "@/utils/get-page-title";

NProgress.configure({ showSpinner: false }); // NProgress Configuration

// 路由守卫
router.beforeEach(async(to, from, next) => {
  NProgress.start();
  document.title = getPageTitle(to.meta.title);

  const handleLoggedIn = async() => {
    if (to.path === "/login") {
      next("/");
      NProgress.done();
    } else {
      const hasRoles = store.getters.roles?.length > 0;
      if (hasRoles) {
        next();
      } else {
        try {
          const { ID, roles } = await store.dispatch("user/getInfo");
          const userinfo = { id: ID, roles };
          const accessRoutes = await store.dispatch("permission/generateRoutes", userinfo);
          accessRoutes.push({ path: "*", redirect: "/404", hidden: true });
          router.addRoutes(accessRoutes);
          next({ ...to, replace: true });
        } catch (error) {
          await store.dispatch("user/resetToken");
          Message.error(`获取用户信息失败: ${error || "未知错误"}`);
          next(`/`);
          NProgress.done();
        }
      }
    }
  };

  const handleLoggedOut = async() => {
    await store.dispatch("permission/generateRoutesAnonymous");
    const whiteList = ["/login", "/auth-redirect", "/sitenav", "/changePass"]; // no redirect whitelist 没有重定向白名单
    const isWhiteListed = whiteList.includes(to.path);
    if (isWhiteListed) {
      next();
    } else {
      next("/");
    }
    NProgress.done();
  };

  const hasToken = getToken();
  if (hasToken) {
    await handleLoggedIn();
  } else {
    await handleLoggedOut();
  }
});

router.afterEach(() => {
  // finish progress bar
  NProgress.done();
});
