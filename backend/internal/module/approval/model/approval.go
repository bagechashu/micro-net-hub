// Package model 定义人工审批的申请单与放行凭证, 以及对应的数据访问方法.
package model

import (
	"time"

	"micro-net-hub/internal/global"

	"gorm.io/gorm"
)

// 审批申请单状态
const (
	RequestStatusPending  uint8 = 1 // 待审批
	RequestStatusApproved uint8 = 2 // 审批已通过
	RequestStatusRejected uint8 = 3 // 审批已拒绝
	RequestStatusExpired  uint8 = 4 // 已过期(超时未处理)
	RequestStatusCanceled uint8 = 5 // 已作废
)

// 审批渠道标识
const (
	ChannelBotTelegram = "bot:telegram" // Telegram 渠道
	ChannelSystem      = "system"       // 系统(过期/清理等自动处理)
)

// ApprovalRequest 一次登录认证的人工审批申请单.
//
// 申请单是审批的最小单元, 同一用户名的重复认证请求会复用同一条待审批申请单,
// 避免客户端重试导致审批人被重复打扰.
type ApprovalRequest struct {
	gorm.Model
	Username string `gorm:"type:varchar(50);not null;index;comment:'申请人用户名'" json:"username"`
	Nickname string `gorm:"type:varchar(50);comment:'申请人昵称'" json:"nickname"`
	// Meta 认证上下文(JSON 文本), 字段随认证场景不同而不同, 仅作留痕审计
	Meta   string `gorm:"type:text;comment:'认证上下文(JSON)'" json:"meta"`
	Status uint8  `gorm:"type:tinyint(1);not null;default:1;index;comment:'状态:1待审批,2已通过,3已拒绝,4已过期,5已作废'" json:"status"`
	// PendingKey 待审批唯一占位键: 申请单处于待审批状态时非空(值为 pending:<username>),
	// 通过唯一索引在数据库层兜底保证同一用户同时只有一条待审批申请单(多实例部署下).
	// 申请单进入终态(通过/拒绝/过期/作废)时置空, 释放占位.
	PendingKey     *string    `gorm:"type:varchar(60);uniqueIndex;comment:'待审批唯一占位键'" json:"-"`
	Reason         string     `gorm:"type:varchar(255);comment:'审批人填写的原因'" json:"reason"`
	Decider        string     `gorm:"type:varchar(128);comment:'审批人标识'" json:"decider"`
	DeciderChannel string     `gorm:"type:varchar(50);comment:'审批渠道, 如 bot:telegram'" json:"deciderChannel"`
	DecidedAt      *time.Time `json:"decidedAt"`
	ExpireAt       time.Time  `gorm:"index;comment:'申请单过期时间'" json:"expireAt"`
	NoticeCount    int        `gorm:"default:0;comment:'已通知审批人的次数'" json:"noticeCount"`
	LastNoticeAt   *time.Time `json:"lastNoticeAt"`
}

// ApprovalGrant 审批通过后签发的放行凭证.
//
// 认证是同步的一次性交互, 审批人未必能在等待窗口内响应, 因此审批通过后签发凭证:
// 申请人在凭证有效期内重新连接即可直接放行, 无需再次审批.
type ApprovalGrant struct {
	gorm.Model
	Username  string    `gorm:"type:varchar(50);not null;index;comment:'放行用户名'" json:"username"`
	RequestID uint      `gorm:"comment:'来源申请单ID'" json:"requestId"`
	ExpireAt  time.Time `gorm:"index;comment:'凭证过期时间'" json:"expireAt"`
	MaxUses   int       `gorm:"default:0;comment:'有效期内最大使用次数, 0 表示不限次'" json:"maxUses"`
	UsedCount int       `gorm:"default:0;comment:'已使用次数'" json:"usedCount"`
	Revoked   bool      `gorm:"default:false;comment:'是否已作废'" json:"revoked"`
	GrantedBy string    `gorm:"type:varchar(128);comment:'签发人标识'" json:"grantedBy"`
	Channel   string    `gorm:"type:varchar(50);comment:'签发渠道, 如 bot:telegram'" json:"channel"`
}

// StatusText 返回申请单状态的展示文案
func (r *ApprovalRequest) StatusText() string {
	switch r.Status {
	case RequestStatusPending:
		return "待审批"
	case RequestStatusApproved:
		return "已通过"
	case RequestStatusRejected:
		return "已拒绝"
	case RequestStatusExpired:
		return "已过期"
	case RequestStatusCanceled:
		return "已作废"
	default:
		return "未知状态"
	}
}

// PendingKeyFor 生成某用户的待审批唯一占位键.
//
// 仅申请单处于待审批状态时写入该字段, 数据库唯一索引据此保证同一用户同时
// 只有一条待审批申请单; 进入终态时置空即可释放占位.
func PendingKeyFor(username string) *string {
	key := "pending:" + username
	return &key
}

// CreateApprovalRequest 创建一条审批申请单
func CreateApprovalRequest(req *ApprovalRequest) error {
	return global.DB.Create(req).Error
}

// FindApprovalRequestByID 按 ID 获取申请单
func FindApprovalRequestByID(id uint) (*ApprovalRequest, error) {
	req := new(ApprovalRequest)
	err := global.DB.Where("id = ?", id).First(req).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return req, nil
}

// FindPendingRequestByUsername 获取某用户当前待审批的申请单, 不存在时返回 nil
func FindPendingRequestByUsername(username string) (*ApprovalRequest, error) {
	req := new(ApprovalRequest)
	err := global.DB.Where("username = ? AND status = ?", username, RequestStatusPending).
		Order("id DESC").First(req).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return req, nil
}

// CountPendingRequests 统计当前待审批的申请单数量
func CountPendingRequests() (int64, error) {
	var count int64
	err := global.DB.Model(&ApprovalRequest{}).Where("status = ?", RequestStatusPending).Count(&count).Error
	return count, err
}

// ListPendingRequests 获取待审批申请单列表, 按创建顺序排列(先来先审)
func ListPendingRequests(limit int) ([]*ApprovalRequest, error) {
	reqs := make([]*ApprovalRequest, 0, limit)
	err := global.DB.Where("status = ?", RequestStatusPending).
		Order("id ASC").Limit(limit).Find(&reqs).Error
	return reqs, err
}

// UpdateNoticeInfo 更新申请单的通知次数与最近通知时间
func UpdateNoticeInfo(id uint, noticeCount int, noticeAt time.Time) error {
	return global.DB.Model(&ApprovalRequest{}).Where("id = ?", id).
		Updates(map[string]interface{}{"notice_count": noticeCount, "last_notice_at": noticeAt}).Error
}

// DecideRequest 原子地把待审批申请单置为终态.
//
// 条件更新保证同一申请单只会被处理一次: 返回 true 表示本次调用是决定该申请单的赢家,
// false 表示已被其他审批人(或其它进程)抢先处理, 调用方据此给出幂等回执.
func DecideRequest(id uint, status uint8, decider, channel, reason string, decidedAt time.Time) (bool, error) {
	res := global.DB.Model(&ApprovalRequest{}).
		Where("id = ? AND status = ?", id, RequestStatusPending).
		Updates(map[string]interface{}{
			"status":          status,
			"decider":         decider,
			"decider_channel": channel,
			"reason":          reason,
			"decided_at":      decidedAt,
			"pending_key":     nil,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

// ExpirePendingRequestsBefore 把已超时且仍在待审批的申请单置为已过期, 返回处理条数
func ExpirePendingRequestsBefore(now time.Time) (int64, error) {
	res := global.DB.Model(&ApprovalRequest{}).
		Where("status = ? AND expire_at <= ?", RequestStatusPending, now).
		Updates(map[string]interface{}{
			"status":          RequestStatusExpired,
			"decider":         "system",
			"decider_channel": ChannelSystem,
			"decided_at":      now,
			"pending_key":     nil,
		})
	return res.RowsAffected, res.Error
}

// CleanupRequestsBefore 删除指定时间之前创建的申请单, 返回删除条数
func CleanupRequestsBefore(before time.Time) (int64, error) {
	res := global.DB.Unscoped().Where("created_at < ?", before).Delete(&ApprovalRequest{})
	return res.RowsAffected, res.Error
}

// CreateApprovalGrant 创建放行凭证
func CreateApprovalGrant(grant *ApprovalGrant) error {
	return global.DB.Create(grant).Error
}

// FindActiveGrant 获取某用户当前有效的放行凭证, 不存在时返回 nil
func FindActiveGrant(username string, now time.Time) (*ApprovalGrant, error) {
	grant := new(ApprovalGrant)
	err := global.DB.Where("username = ? AND revoked = ? AND expire_at > ?", username, false, now).
		Order("id DESC").First(grant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	if grant.MaxUses > 0 && grant.UsedCount >= grant.MaxUses {
		return nil, nil
	}
	return grant, nil
}

// ConsumeGrant 原子地消耗一次凭证使用次数.
//
// 在数据库层同时校验凭证仍有效(未作废、未过期、未超出最大使用次数), 仅当满足条件时
// 才自增 used_count; 返回 true 表示本次消耗成功. max_uses 为 0 表示不限次.
func ConsumeGrant(id uint, now time.Time) (bool, error) {
	res := global.DB.Model(&ApprovalGrant{}).
		Where("id = ? AND revoked = ? AND expire_at > ?", id, false, now).
		Where("max_uses = 0 OR used_count < max_uses").
		UpdateColumn("used_count", gorm.Expr("used_count + 1"))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

// RevokeGrantByUsername 作废某用户当前的有效凭证
func RevokeGrantByUsername(username string) error {
	return global.DB.Model(&ApprovalGrant{}).
		Where("username = ? AND revoked = ?", username, false).
		Update("revoked", true).Error
}

// CleanupGrantsBefore 删除指定时间之前创建的凭证, 返回删除条数
func CleanupGrantsBefore(before time.Time) (int64, error) {
	res := global.DB.Unscoped().Where("created_at < ?", before).Delete(&ApprovalGrant{})
	return res.RowsAffected, res.Error
}
