import request from '@/utils/request';

export function getMenuTree() {
  return request({
    url: '/api/menu/tree',
    method: 'get'
  });
}

export function getMenus() {
  return request({
    url: '/api/menu/list',
    method: 'get'
  });
}

export function createMenu(data: unknown) {
  return request({
    url: '/api/menu/add',
    method: 'post',
    data
  });
}

export function updateMenuById(data: unknown) {
  return request({
    url: '/api/menu/update',
    method: 'post',
    data
  });
}

export function batchDeleteMenuByIds(data: unknown) {
  return request({
    url: '/api/menu/delete',
    method: 'post',
    data
  });
}

export function getUserMenusByUserId(Id: string | number) {
  return request({
    url: '/api/menu/access/list/' + Id,
    method: 'get'
  });
}

export function getUserMenuTreeByUserId(params: unknown) {
  return request({
    url: '/api/menu/access/tree',
    method: 'get',
    params
  });
}
