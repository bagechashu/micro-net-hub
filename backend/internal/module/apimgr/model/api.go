package model

import "gorm.io/gorm"

type Api struct {
	gorm.Model
	Method   string `gorm:"type:varchar(20);comment:'请求方式'" json:"method"`
	Path     string `gorm:"type:varchar(100);comment:'访问路径'" json:"path"`
	Category string `gorm:"type:varchar(50);comment:'所属类别'" json:"category"`
	Remark   string `gorm:"type:varchar(100);comment:'备注'" json:"remark"`
	Creator  string `gorm:"type:varchar(20);comment:'创建人'" json:"creator"`
}

// ApiListReq 获取资源列表结构体
type ApiListReq struct {
	Method   string `json:"method" form:"method"`
	Path     string `json:"path" form:"path"`
	Category string `json:"category" form:"category"`
	Creator  string `json:"creator" form:"creator"`
	PageNum  int    `json:"pageNum" form:"pageNum"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

// ApiAddReq 添加资源结构体
type ApiAddReq struct {
	Method   string `json:"method" validate:"required,min=1,max=20"`
	Path     string `json:"path" validate:"required,min=1,max=100"`
	Category string `json:"category" validate:"required,min=1,max=50"`
	Remark   string `json:"remark" validate:"min=0,max=100"`
}

// ApiUpdateReq 更新资源结构体
type ApiUpdateReq struct {
	ID       uint   `json:"id" validate:"required"`
	Method   string `json:"method" validate:"min=1,max=20"`
	Path     string `json:"path" validate:"min=1,max=100"`
	Category string `json:"category" validate:"min=1,max=50"`
	Remark   string `json:"remark" validate:"min=0,max=100"`
}

// ApiDeleteReq 删除资源结构体
type ApiDeleteReq struct {
	ApiIds []uint `json:"apiIds" validate:"required"`
}

// ApiGetTreeReq 获取资源树结构体
type ApiGetTreeReq struct {
}

type ApiTreeRsp struct {
	ID       int    `json:"ID"`
	Remark   string `json:"remark"`
	Category string `json:"category"`
	Children []*Api `json:"children"`
}

type ApiListRsp struct {
	Total int64 `json:"total"`
	Apis  []Api `json:"apis"`
}
