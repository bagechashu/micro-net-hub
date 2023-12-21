package response

import "micro-net-hub/model"

type RoleListRsp struct {
	Total int64        `json:"total"`
	Roles []model.Role `json:"roles"`
}
