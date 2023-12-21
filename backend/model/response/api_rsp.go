package response

import "micro-net-hub/model"

type ApiTreeRsp struct {
	ID       int          `json:"ID"`
	Remark   string       `json:"remark"`
	Category string       `json:"category"`
	Children []*model.Api `json:"children"`
}

type ApiListRsp struct {
	Total int64       `json:"total"`
	Apis  []model.Api `json:"apis"`
}
