package response

import "micro-net-hub/model"

type LogListRsp struct {
	Total int64                `json:"total"`
	Logs  []model.OperationLog `json:"logs"`
}
