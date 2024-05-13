package dashboard

import (
	"fmt"
	accountModel "micro-net-hub/internal/module/account/model"
	apiMgrModel "micro-net-hub/internal/module/apimgr/model"
	opLogModel "micro-net-hub/internal/module/operationlog/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type DashboardListRsp struct {
	DataType  string `json:"dataType"`
	DataName  string `json:"dataName"`
	DataCount int64  `json:"dataCount"`
	Icon      string `json:"icon"`
	Path      string `json:"path"`
}

// Dashboard 系统首页展示数据
func Dashboard(c *gin.Context) {
	userCount, err := accountModel.UserCount()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取用户总数失败")))
		return
	}

	groupCount, err := accountModel.GroupCount()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取分组总数失败")))
		return
	}
	roleCount, err := accountModel.RoleCount()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取角色总数失败")))
		return
	}
	menuCount, err := accountModel.MenuCount()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取菜单总数失败")))
		return
	}
	apiCount, err := apiMgrModel.Count()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取接口总数失败")))
		return
	}
	logCount, err := opLogModel.Count()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取日志总数失败")))
		return
	}

	rst := make([]*DashboardListRsp, 0)

	rst = append(rst,
		&DashboardListRsp{
			DataType:  "user",
			DataName:  "用户",
			DataCount: userCount,
			Icon:      "people",
			Path:      "#/personnel/user",
		},
		&DashboardListRsp{
			DataType:  "group",
			DataName:  "分组",
			DataCount: groupCount,
			Icon:      "peoples",
			Path:      "#/personnel/group",
		},
		&DashboardListRsp{
			DataType:  "role",
			DataName:  "角色",
			DataCount: roleCount,
			Icon:      "eye-open",
			Path:      "#/system/role",
		},
		&DashboardListRsp{
			DataType:  "menu",
			DataName:  "菜单",
			DataCount: menuCount,
			Icon:      "tree-table",
			Path:      "#/system/menu",
		},
		&DashboardListRsp{
			DataType:  "api",
			DataName:  "接口",
			DataCount: apiCount,
			Icon:      "tree",
			Path:      "#/system/api",
		},
		&DashboardListRsp{
			DataType:  "log",
			DataName:  "日志",
			DataCount: logCount,
			Icon:      "documentation",
			Path:      "#/log/operation-log",
		},
	)

	helper.Success(c, rst)
}
