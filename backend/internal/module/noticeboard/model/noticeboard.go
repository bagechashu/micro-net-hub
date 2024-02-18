package model

import "gorm.io/gorm"

// TODO: add an Notice Board
type NoticeBoard struct {
	gorm.Model
	Level   uint   `gorm:"type:bigint(20);not null;comment:'公告级别'" json:"level" validate:"required"`
	Content string `gorm:"type:varchar(200);not null;comment:'公告内容'" json:"content" validate:"required"`
	Creator string `gorm:"type:varchar(20);comment:'创建人'" json:"creator" form:"creator"`
}

const (
	LevelInfo     uint = iota + 1 // 1
	LevelWarning                  // 2
	LevelCritical                 // 3
)

func (n NoticeBoard) LevelText() string {
	switch n.Level {
	case LevelInfo:
		return "Info"
	case LevelWarning:
		return "Warning"
	case LevelCritical:
		return "Critical:"
	default:
		return "Other"
	}
}
