import request from "@/utils/request";
// 获取首页的数据 (已完成)
export function getSiteNav() {
  return request({
    url: "/api/sitenav/getall",
    method: "get"
  });
}

