package dns

import (
	"fmt"
	"micro-net-hub/internal/module/account/current"
	"micro-net-hub/internal/module/dns/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

// GetGroups returns a list of groups.
func GetAll(c *gin.Context) {
	dzs := new(model.DnsZones)

	err := dzs.FindWithRecords()
	if err != nil {
		helper.ErrV2(c, helper.ReloadErr(err))
		return
	}

	helper.Success(c, dzs)
}

type ZoneAddReq struct {
	Name string `json:"name" validate:"required"`
}

func AddZone(c *gin.Context) {
	var req ZoneAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := current.GetCurrentLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	dz := model.DnsZone{
		Name:    req.Name,
		Creator: ctxUser.Username,
	}
	err = dz.Add()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("创建 Zone 失败: %s", err.Error())))
		return
	}
	model.CacheDnsZonesClear()
	helper.Success(c, nil)
}

type ZoneUpdateReq struct {
	Id   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

func UpdateZone(c *gin.Context) {
	var req ZoneUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := current.GetCurrentLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	dz := &model.DnsZone{}
	err = dz.Find(map[string]interface{}{"ID": req.Id})
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取 Zone 失败: %s", err.Error())))
		return
	}

	dz.Name = req.Name
	dz.Creator = ctxUser.Username
	err = dz.Update()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新 Zone 失败: %s", err.Error())))
		return
	}

	model.CacheDnsZonesClear()
	helper.Success(c, nil)
}

type ZoneDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

func DeleteZone(c *gin.Context) {
	var req ZoneDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}
	dz := &model.DnsZone{}
	for _, id := range req.Ids {
		err := dz.Find(map[string]interface{}{"ID": id})
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("查询 Zone 失败: %s", err.Error())))
			return
		}
		err = dz.DeleteWithRecords()
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除 Zone 失败: %s", err.Error())))
			return
		}
	}

	model.CacheDnsZonesClear()
	helper.Success(c, nil)
}

type RecordAddReq struct {
	ZoneID uint   `json:"zone_id" validate:"required"`
	Host   string `json:"host" validate:"required"`
	Type   string `json:"type" validate:"required"`
	Value  string `json:"value" validate:"required"`
	Ttl    uint32 `json:"ttl" validate:"required"`
}

func AddRecord(c *gin.Context) {
	var req RecordAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := current.GetCurrentLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	dr := model.DnsRecord{
		ZoneID:  req.ZoneID,
		Host:    req.Host,
		Type:    req.Type,
		Value:   req.Value,
		Ttl:     req.Ttl,
		Creator: ctxUser.Username,
	}

	err = dr.Add()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("创建 Record 失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

type RecordUpdateReq struct {
	ID     uint   `json:"id" validate:"required"`
	ZoneID uint   `json:"zone_id" validate:"required"`
	Host   string `json:"host" validate:"required"`
	Type   string `json:"type" validate:"required"`
	Value  string `json:"value" validate:"required"`
	Ttl    uint32 `json:"ttl" validate:"required"`
}

func UpdateRecord(c *gin.Context) {
	var req RecordUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := current.GetCurrentLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	dr := &model.DnsRecord{}
	err = dr.Find(map[string]interface{}{"ID": req.ID})
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取 Record 失败: %s", err.Error())))
		return
	}

	dr.ZoneID = req.ZoneID
	dr.Host = req.Host
	dr.Type = req.Type
	dr.Value = req.Value
	dr.Ttl = req.Ttl
	dr.Creator = ctxUser.Username

	err = dr.Update()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新 Record 失败: %s", err.Error())))
		return
	}

	model.CacheDnsRecordDel(dr.ZoneID, dr.Host)
	helper.Success(c, nil)
}

type RecordDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

func DeleteRecord(c *gin.Context) {
	var req RecordDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	for _, id := range req.Ids {
		dr := &model.DnsRecord{}
		err := dr.Find(map[string]interface{}{"ID": id})
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取 Record 失败: %s", err.Error())))
			return
		}
		err = dr.Delete()
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除 Record 失败: %s", err.Error())))
			return
		}

		model.CacheDnsRecordDel(dr.ZoneID, dr.Host)
	}

	helper.Success(c, nil)
}
