package noticeboard

import (
	"fmt"
	"micro-net-hub/internal/module/account/auth"
	"micro-net-hub/internal/module/noticeboard/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

// GetGroups returns a list of groups.
func GetAll(c *gin.Context) {
	ns := new(model.NoticeBoards)

	err := ns.Find(map[string]interface{}{})
	if err != nil {
		helper.ErrV2(c, helper.ReloadErr(err))
		return
	}

	helper.Success(c, ns)
}

type NoticeBoardAddReq struct {
	Level   uint   `json:"level" validate:"required"`
	Content string `json:"content" validate:"required"`
}

func AddNoticeBoard(c *gin.Context) {
	var req NoticeBoardAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	n := model.NoticeBoard{
		Level:   req.Level,
		Content: req.Content,
		Creator: ctxUser.Username,
	}
	err = n.Add()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("创建 NoticeBoard 失败: %s", err.Error())))
		return
	}
	// model.CacheNoticeBoardClear()
	helper.Success(c, nil)
}

type NoticeBoardUpdateReq struct {
	Id      string `json:"id" validate:"required"`
	Level   uint   `json:"level" validate:"required"`
	Content string `json:"content" validate:"required"`
}

func UpdateNoticeBoard(c *gin.Context) {
	var req NoticeBoardUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	n := &model.NoticeBoard{}
	err = n.Find(map[string]interface{}{"ID": req.Id})
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取 NoticeBoard 失败: %s", err.Error())))
		return
	}

	n.Level = req.Level
	n.Content = req.Content
	n.Creator = ctxUser.Username
	err = n.Update()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新 NoticeBoard 失败: %s", err.Error())))
		return
	}

	// model.CacheNoticeBoardClear()
	helper.Success(c, nil)
}

type NoticeBoardDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

func DeleteNoticeBoard(c *gin.Context) {
	var req NoticeBoardDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}
	n := &model.NoticeBoard{}
	for _, id := range req.Ids {
		err := n.Find(map[string]interface{}{"ID": id})
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("查询 NoticeBoard 失败: %s", err.Error())))
			return
		}
		err = n.Delete()
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除 NoticeBoard 失败: %s", err.Error())))
			return
		}
	}

	// model.CacheNoticeBoardClear()
	helper.Success(c, nil)
}
