import { defineStore } from 'pinia';
import { login as apiLogin, logout as apiLogout, refreshToken as apiRefreshToken } from '@/api/system/base';
import { getInfo as apiGetInfo } from '@/api/system/user';
import { getToken, setToken, removeToken } from '@/utils/auth';
import { resetRouter, addDynamicRoutes } from '@/router';
import usePermissionStore from './permission';

export interface IUserState {
  token: string;
  userId: number | string;
  name: string;
  mail: string;
  avatar: string;
  introduction: string;
  roles: string[];
  /** /api/user/info 原始返回，个人中心展示昵称、手机号、职位等 */
  info: IUserInfo | null;
}

/** /api/user/info 返回的用户信息（backend/internal/module/account/model/user.go） */
export interface IUserInfo {
  ID: number;
  username: string;
  mail?: string;
  avatar?: string;
  introduction?: string;
  roles?: Array<{ name: string; keyword?: string }>;
  [key: string]: unknown;
}

const useUserStore = defineStore('user', {
  state: (): IUserState => ({
    token: getToken() || '',
    userId: '',
    name: '',
    mail: '',
    avatar: '',
    introduction: '',
    roles: [],
    info: null
  }),
  actions: {
    async login(userInfo: { username: string; password: string }) {
      const { username, password } = userInfo;
      const { data } = await apiLogin({ username: username.trim(), password });
      this.token = data.token;
      setToken(data.token);
    },

    async getInfo(): Promise<IUserInfo> {
      const response = await apiGetInfo();
      const data = response?.data as IUserInfo | null;
      if (!data) {
        throw new Error('Verification failed, please Login again.');
      }

      const { ID, roles, username, mail, avatar, introduction } = data;
      // roles must be a non-empty array
      if (!roles || roles.length <= 0) {
        throw new Error('getInfo: roles must be a non-null array!');
      }

      this.userId = ID;
      this.roles = roles.map(r => r.name);
      this.name = username;
      this.mail = mail || '';
      this.avatar = avatar || '';
      this.introduction = introduction || '';
      this.info = data;
      return data;
    },

    async logout() {
      try {
        await apiLogout();
      } finally {
        // 即使后端登出失败，也要清掉本地状态
        this.resetToken();
        resetRouter();
        usePermissionStore().generateRoutesAnonymous();
      }
    },

    async refreshToken() {
      const { data } = await apiRefreshToken();
      this.token = data.token;
      setToken(data.token);
    },

    // remove token
    resetToken() {
      this.token = '';
      this.userId = '';
      this.name = '';
      this.mail = '';
      this.avatar = '';
      this.introduction = '';
      this.roles = [];
      this.info = null;
      removeToken();
    },

    // 重新拉取用户信息和菜单，刷新动态路由
    async resetAccessRouter() {
      const info = await this.getInfo();
      resetRouter();
      const accessRoutes = await usePermissionStore().generateRoutes(info);
      addDynamicRoutes(accessRoutes);
    }
  }
});
export default useUserStore;
