package model

import (
	"errors"
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"strings"
	"time"

	"gorm.io/gorm"
)

type OperationLog struct {
	gorm.Model
	Username   string `gorm:"type:varchar(20);comment:'用户登录名'" json:"username"`
	Ip         string `gorm:"type:varchar(20);comment:'Ip地址'" json:"ip"`
	IpLocation string `gorm:"type:varchar(20);comment:'Ip所在地'" json:"ipLocation"`
	Method     string `gorm:"type:varchar(20);comment:'请求方式'" json:"method"`
	Path       string `gorm:"type:varchar(100);comment:'访问路径'" json:"path"`
	Remark     string `gorm:"type:varchar(100);comment:'备注'" json:"remark"`
	Status     int    `gorm:"type:int(4);comment:'响应状态码'" json:"status"`
	StartTime  string `gorm:"type:varchar(2048);comment:'发起时间'" json:"startTime"`
	TimeCost   int64  `gorm:"type:int(6);comment:'请求耗时(ms)'" json:"timeCost"`
	UserAgent  string `gorm:"type:varchar(2048);comment:'浏览器标识'" json:"userAgent"`
}

// OperationLogListReq 操作日志请求结构体
type OperationLogListReq struct {
	Username string `json:"username" form:"username"`
	Ip       string `json:"ip" form:"ip"`
	Path     string `json:"path" form:"path"`
	Status   int    `json:"status" form:"status"`
	PageNum  int    `json:"pageNum" form:"pageNum"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

// OperationLogDeleteReq 批量删除操作日志结构体
type OperationLogDeleteReq struct {
	OperationLogIds []uint `json:"operationLogIds" validate:"required"`
}

type LogListRsp struct {
	Total int64          `json:"total"`
	Logs  []OperationLog `json:"logs"`
}

// List 获取数据列表
func List(req *OperationLogListReq) ([]*OperationLog, error) {
	var list []*OperationLog
	db := global.DB.Model(&OperationLog{}).Order("id DESC")

	username := strings.TrimSpace(req.Username)
	if username != "" {
		db = db.Where("username LIKE ?", fmt.Sprintf("%%%s%%", username))
	}
	ip := strings.TrimSpace(req.Ip)
	if ip != "" {
		db = db.Where("ip LIKE ?", fmt.Sprintf("%%%s%%", ip))
	}
	path := strings.TrimSpace(req.Path)
	if path != "" {
		db = db.Where("path LIKE ?", fmt.Sprintf("%%%s%%", path))
	}
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}

	pageReq := tools.NewPageOption(req.PageNum, req.PageSize)
	err := db.Offset(pageReq.PageNum).Limit(pageReq.PageSize).Find(&list).Error

	return list, err
}

// Count 获取数据总数
func Count() (count int64, err error) {
	err = global.DB.Model(&OperationLog{}).Count(&count).Error
	return count, err
}

// 获取单个用户
func Find(filter map[string]interface{}, data *OperationLog) error {
	return global.DB.Where(filter).First(&data).Error
}

// Exist 判断资源是否存在
func Exist(filter map[string]interface{}) bool {
	var dataObj OperationLog
	err := global.DB.Order("created_at DESC").Where(filter).First(&dataObj).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// Delete 删除资源
func Delete(operationLogIds []uint) error {
	return global.DB.Where("id IN (?)", operationLogIds).Unscoped().Delete(&OperationLog{}).Error
}

// var Logs []OperationLog //全局变量多个线程需要加锁，所以每个线程自己维护一个
// 处理OperationLogChan将日志记录到数据库
func SaveOperationLogChannel(olc <-chan *OperationLog) {
	// 只会在线程开启的时候执行一次
	Logs := make([]OperationLog, 0)
	//5s 自动同步一次
	duration := 5 * time.Second
	timer := time.NewTimer(duration)
	defer timer.Stop()
	for {
		select {
		case log := <-olc:
			Logs = append(Logs, *log)
			// 每10条记录到数据库
			if len(Logs) > 5 {
				global.DB.Create(&Logs)
				Logs = make([]OperationLog, 0)
				timer.Reset(duration) // 入库重置定时器
			}
		case <-timer.C: //5s 自动同步一次
			if len(Logs) > 0 {
				global.DB.Create(&Logs)
				Logs = make([]OperationLog, 0)
			}
			timer.Reset(duration) // 入库重置定时器
		}
	}
}
