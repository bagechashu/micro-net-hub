import request from '@/utils/request';

export function getApis(params: unknown) {
  return request({
    url: '/api/api/list',
    method: 'get',
    params
  });
}

export function getApiTree(params: unknown) {
  return request({
    url: '/api/api/tree',
    method: 'get',
    params
  });
}

export function createApi(data: unknown) {
  return request({
    url: '/api/api/add',
    method: 'post',
    data
  });
}

export function updateApiById(data: unknown) {
  return request({
    url: '/api/api/update',
    method: 'post',
    data
  });
}

export function batchDeleteApiByIds(data: unknown) {
  return request({
    url: '/api/api/delete',
    method: 'post',
    data
  });
}
