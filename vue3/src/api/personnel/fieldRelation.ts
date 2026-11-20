import request from '@/utils/request';

// 字段动态关系列表（完成）
export function relationList(params: unknown) {
  return request({
    url: '/api/goldap/fieldrelation/list',
    method: 'get',
    params
  });
}
// 添加字段动态关系（完成）
export function relationAdd(data: unknown) {
  return request({
    url: '/api/goldap/fieldrelation/add',
    method: 'post',
    data
  });
}
// 更新字段动态关系 （完成）
export function relationUp(data: unknown) {
  return request({
    url: '/api/goldap/fieldrelation/update',
    method: 'post',
    data
  });
}
// 删除字段动态关系（完成）
export function relationDel(data: unknown) {
  return request({
    url: '/api/goldap/fieldrelation/delete',
    method: 'post',
    data
  });
}

