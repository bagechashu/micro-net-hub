package model

// BaseDashboardReq  系统首页展示数据结构体
type BaseDashboardReq struct {
}

type DashboardListRsp struct {
	DataType  string `json:"dataType"`
	DataName  string `json:"dataName"`
	DataCount int64  `json:"dataCount"`
	Icon      string `json:"icon"`
	Path      string `json:"path"`
}
