import request from "@/utils/request";
// 获取首页的数据 (已完成)
export function getNav() {
  return request({
    url: "/api/sitenav/getnav",
    method: "get"
  });
}

// 创建接口（已完成）
export function createNavGroup(data) {
  return request({
    url: "/api/sitenav/group/add",
    method: "post",
    data
  });
}

// 更新接口（已完成）
export function updateNavGroup(data) {
  return request({
    url: "/api/sitenav/group/update",
    method: "post",
    data
  });
}

// 批量删除接口（已完成）
export function batchDeleteNavGroupByIds(data) {
  return request({
    url: "/api/sitenav/group/delete",
    method: "post",
    data
  });
}

// 创建接口（已完成）
export function createNavSite(data) {
  return request({
    url: "/api/sitenav/site/add",
    method: "post",
    data
  });
}

// 更新接口（已完成）
export function updateNavSite(data) {
  return request({
    url: "/api/sitenav/site/update",
    method: "post",
    data
  });
}

// 批量删除接口（已完成）
export function batchDeleteNavSiteByIds(data) {
  return request({
    url: "/api/sitenav/site/delete",
    method: "post",
    data
  });
}
