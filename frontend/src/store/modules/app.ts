import { defineStore } from 'pinia';
import Cookies from 'js-cookie';
import defaultSettings from '@/settings';
import { getSiteInfo } from '@/api/system/base';

interface IAppState {
  size: 'large' | 'default' | 'small';
  /** 站点名称：来自后端 system.site-name，缓存在 localStorage 避免刷新时先闪一下默认名 */
  siteName: string;
  /** 自定义 logo 地址：来自后端 system.site-logo，空字符串表示使用内置 logo（只用于顶部品牌，浏览器图标保持默认） */
  siteLogo: string;
  /** 登录弹框（顶部栏）是否打开：放在全局，登录过期提示等地方也能打开 */
  loginDialogVisible: boolean;
  /** 登录已过期提示是否显示（utils/request.ts 收到 token 过期时设置） */
  sessionExpired: boolean;
}

export const SITE_NAME_KEY = 'site-name';
export const SITE_LOGO_KEY = 'site-logo';

function cachedSiteLogo(): string {
  try {
    return localStorage.getItem(SITE_LOGO_KEY) || '';
  } catch {
    return '';
  }
}

function cachedSiteName(): string {
  try {
    return localStorage.getItem(SITE_NAME_KEY) || defaultSettings.title;
  } catch {
    return defaultSettings.title;
  }
}

/** 全局组件尺寸（App.vue 中的 el-config-provider 使用）与站点名称 */
const useAppStore = defineStore('app', {
  state: (): IAppState => ({
    size: (Cookies.get('size') as IAppState['size']) || 'default',
    siteName: cachedSiteName(),
    siteLogo: cachedSiteLogo(),
    loginDialogVisible: false,
    sessionExpired: false
  }),
  actions: {
    setSize(size: IAppState['size']) {
      this.size = size;
      Cookies.set('size', size);
    },
    /** 自定义 logo 加载失败时回退内置 logo */
    resetSiteLogo() {
      this.siteLogo = '';
    },
    async loadSiteName() {
      try {
        const res: any = await getSiteInfo();
        const name = String(res?.data?.name || '').trim();
        // 只接受后端返回的同源相对地址
        const logo = String(res?.data?.logo || '');
        this.siteLogo = logo.startsWith('/') && !logo.startsWith('//') ? logo : '';
        if (name) this.siteName = name;
        try {
          if (name) localStorage.setItem(SITE_NAME_KEY, name);
          if (this.siteLogo) localStorage.setItem(SITE_LOGO_KEY, this.siteLogo);
          else localStorage.removeItem(SITE_LOGO_KEY);
        } catch {
          // 忽略存储不可用
        }
      } catch {
        // 后端不可用时沿用缓存或默认名称
      }
    }
  }
});

export default useAppStore;
