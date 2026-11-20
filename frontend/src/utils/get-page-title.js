import useAppStore from '@/store/modules/app';

// 浏览器标题：「页面标题 - 站点名称」，站点名称来自后端配置（见 store/modules/app.ts）
export default function getPageTitle(pageTitle) {
  const title = useAppStore().siteName;
  if (pageTitle) {
    return `${pageTitle} - ${title}`;
  }
  return `${title}`;
}
