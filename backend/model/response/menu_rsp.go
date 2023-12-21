package response

import "micro-net-hub/model"

type MenuListRsp struct {
	Total int64        `json:"total"`
	Menus []model.Menu `json:"menus"`
}
