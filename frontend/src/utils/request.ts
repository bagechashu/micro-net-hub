import axios from 'axios';
import type { AxiosError, AxiosResponse } from 'axios';
import { ElMessage, ElMessageBox } from 'element-plus';
import router from '@/router';
import userStore from '@/store/modules/user';
import { getToken } from '@/utils/auth';

/**
 * 后端统一响应格式（backend/internal/server/helper/response.go）：
 * - 成功：{ code: 200, msg, data }；登录接口返回 { code: 200, message, data }
 * - 业务失败：HTTP 200 + { errcode, msg } 或 { code: 非 200, msg }
 * - JWT 失败：HTTP 401 + { code: 401, message }
 */
export interface ApiResponse<T = any> {
  code?: number;
  errcode?: number;
  msg?: string;
  message?: string;
  data: T;
}

const service = axios.create({
  // 接口地址本身已带 /api 前缀；开发时由 vite 代理到后端
  baseURL: '/',
  timeout: 30000
});

service.interceptors.request.use(
  config => {
    const token = getToken();
    if (token) {
      config.headers.Authorization = 'Bearer ' + token;
    }
    return config;
  },
  error => Promise.reject(error)
);

function showError(message: string) {
  ElMessage({ showClose: true, message, type: 'error', duration: 5 * 1000 });
}

let expiredDialogOpen = false;

function handleTokenExpired() {
  if (expiredDialogOpen) return;
  expiredDialogOpen = true;
  ElMessageBox.alert('登录状态过期, 请重新登录.', '登录状态过期', {
    confirmButtonText: '确定',
    type: 'warning'
  }).finally(async () => {
    expiredDialogOpen = false;
    await userStore().logout().catch(() => undefined);
    router.push('/');
  });
}

service.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const res = response.data;
    if (res && (res.code === 200 || res.code === 0)) {
      return res as any;
    }
    const message = res?.msg || res?.message || 'Error';
    showError(message);
    return Promise.reject(new Error(message));
  },
  (error: AxiosError<ApiResponse>) => {
    const response = error.response;
    // 断网、超时、跨域等情况没有 response
    if (!response) {
      showError(error.message || '网络错误');
      return Promise.reject(error);
    }

    const message = response.data?.message || response.data?.msg || error.message;
    if (response.status === 401 && message?.includes('Token is expired')) {
      handleTokenExpired();
    } else if (response.status === 403) {
      router.push({ path: '/401' });
    } else {
      showError(message);
    }
    return Promise.reject(error);
  }
);

/** 返回值已被拦截器解包为 ApiResponse */
export default service as unknown as <T = any>(config: Parameters<typeof service.request>[0]) => Promise<ApiResponse<T>>;
