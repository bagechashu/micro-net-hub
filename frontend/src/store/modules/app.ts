import { defineStore } from 'pinia';
import Cookies from 'js-cookie';
import defaultSettings from '@/settings';
import { getSiteInfo } from '@/api/system/base';

interface IAppState {
  size: 'large' | 'default' | 'small';
  /** 站点名称：来自后端 system.site-name，缓存在 localStorage 避免刷新时先闪一下默认名 */
  siteName: string;
  /** 自定义 logo 地址：来自后端 system.site-logo，空字符串表示使用内置 logo */
  siteLogo: string;
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

/** 浏览器标签页图标：有自定义 logo 时用它，否则恢复内置 favicon */
export function applyFavicon(logo: string) {
  const links = document.querySelectorAll<HTMLLinkElement>('link[rel~="icon"]');
  links.forEach(link => {
    if (link.dataset.default === undefined) {
      link.dataset.default = link.getAttribute('href') || '';
      link.dataset.defaultType = link.getAttribute('type') || '';
    }
    link.href = logo || link.dataset.default;
    // 自定义 logo 可能是 png 等格式，不能沿用默认的 type="image/svg+xml"
    if (logo || !link.dataset.defaultType) link.removeAttribute('type');
    else link.setAttribute('type', link.dataset.defaultType);
  });
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
    siteLogo: cachedSiteLogo()
  }),
  actions: {
    setSize(size: IAppState['size']) {
      this.size = size;
      Cookies.set('size', size);
    },
    /** 自定义 logo 加载失败时回退内置 logo */
    resetSiteLogo() {
      this.siteLogo = '';
      applyFavicon('');
    },
    async loadSiteName() {
      try {
        const res: any = await getSiteInfo();
        const name = String(res?.data?.name || '').trim();
        // 只接受后端返回的同源相对地址
        const logo = String(res?.data?.logo || '');
        this.siteLogo = logo.startsWith('/') && !logo.startsWith('//') ? logo : '';
        applyFavicon(this.siteLogo);
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
