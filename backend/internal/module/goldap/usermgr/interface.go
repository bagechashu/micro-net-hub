package usermgr

import (
	"sync"
)

type UserMgrPlat interface {
	// // 获取所有部门
	// GetAllDepts() (ret []map[string]interface{}, err error)
	// // 获取所有员工信息
	// GetAllUsers() (ret []map[string]interface{}, err error)
	// // 获取离职人员ID列表
	// GetLeaveUserIds() ([]string, error)
	// // 新接口根据时间范围获取离职人员ID列表
	// GetLeaveUserIdsDateRange(pushDays uint) ([]string, error)
	// // 修改密码
	// GetUserDeptIds(udn string) (ret []string, err error)
	SyncDepts() error
	SyncUsers() error
}

var once sync.Once
