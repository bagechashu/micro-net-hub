import request from '@/utils/request';

export function getRoles(params: unknown) {
  return request({
    url: '/api/role/list',
    method: 'get',
    params
  });
}

export function createRole(data: unknown) {
  return request({
    url: '/api/role/add',
    method: 'post',
    data
  });
}

export function updateRoleById(data: unknown) {
  return request({
    url: '/api/role/update',
    method: 'post',
    data
  });
}

export function getRoleMenusById(params: unknown) {
  return request({
    url: '/api/role/getmenulist',
    method: 'get',
    params
  });
}

export function updateRoleMenusById(data: unknown) {
  return request({
    url: '/api/role/updatemenus',
    method: 'post',
    data
  });
}

export function getRoleApisById(params: unknown) {
  return request({
    url: '/api/role/getapilist',
    method: 'get',
    params
  });
}

export function updateRoleApisById(data: unknown) {
  return request({
    url: '/api/role/updateapis',
    method: 'post',
    data
  });
}

export function batchDeleteRoleByIds(data: unknown) {
  return request({
    url: '/api/role/delete',
    method: 'post',
    data
  });
}
