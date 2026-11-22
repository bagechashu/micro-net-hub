import request from '@/utils/request';

export function login(data: unknown) {
  return request({
    url: '/api/base/login',
    method: 'post',
    data
  });
}

export function refreshToken() {
  return request({
    url: '/api/base/refreshToken',
    method: 'get'
  });
}

export function logout() {
  return request({
    url: '/api/base/logout',
    method: 'post'
  });
}

// 站点名称等公开信息（无需登录），名称在后端 config.yml 的 system.site-name 配置
export function getSiteInfo() {
  return request({
    url: '/api/base/siteinfo',
    method: 'get'
  });
}

// 当前 RSA 公钥（无需登录），用于加密提交的口令，见 utils/rsa.ts
export function getPublicKey() {
  return request({
    url: '/api/base/pubkey',
    method: 'get'
  });
}
