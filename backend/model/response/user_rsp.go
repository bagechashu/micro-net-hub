package response

import "micro-net-hub/model"

type UserListRsp struct {
	Total int          `json:"total"`
	Users []model.User `json:"users"`
}
