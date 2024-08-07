const getters = {
  sidebar: state => state.app.sidebar,
  size: state => state.app.size,
  locale: state => state.app.locale,
  device: state => state.app.device,
  visitedViews: state => state.tagsView.visitedViews,
  cachedViews: state => state.tagsView.cachedViews,
  token: state => state.user.token,
  avatar: state => state.user.avatar,
  name: state => state.user.name,
  mail: state => state.user.mail,
  introduction: state => state.user.introduction,
  roles: state => state.user.roles,
  permission_routes: state => state.permission.routes,
  errorLogs: state => state.errorLog.logs,
  routes: state => state.permission.routes
};
export default getters;
