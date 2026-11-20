import request from '@/utils/request';

export function getInfo() {
  return request({
    url: '/api/user/info',
    method: 'get'
  });
}

export function getUsers(params: unknown) {
  return request({
    url: '/api/user/list',
    method: 'get',
    params
  });
}

export function changePwd(data: unknown) {
  return request({
    url: '/api/user/changePwd',
    method: 'post',
    data
  });
}

export function sendCode(data: unknown) {
  return request({
    url: '/api/base/sendcode',
    method: 'post',
    data
  });
}

export function emailPass(data: unknown) {
  return request({
    url: '/api/base/changePwd',
    method: 'post',
    data
  });
}

export function createUser(data: unknown) {
  return request({
    url: '/api/user/add',
    method: 'post',
    data
  });
}

export function updateUserById(data: unknown) {
  return request({
    url: '/api/user/update',
    method: 'post',
    data
  });
}

export function batchDeleteUserByIds(data: unknown) {
  return request({
    url: '/api/user/delete',
    method: 'post',
    data
  });
}
