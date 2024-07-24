import request from "@/utils/request";
// 获取首页的数据 (已完成)
export function getNotice() {
  return request({
    url: "/api/notice/getall",
    method: "get"
  });
}

// 创建接口（已完成）
export function createNotice(data) {
  return request({
    url: "/api/notice/mgr/add",
    method: "post",
    data
  });
}

// 更新接口（已完成）
export function updateNotice(data) {
  return request({
    url: "/api/notice/mgr/update",
    method: "post",
    data
  });
}

// 批量删除接口（已完成）
export function batchDeleteNoticeByIds(data) {
  return request({
    url: "/api/notice/mgr/delete",
    method: "post",
    data
  });
}
