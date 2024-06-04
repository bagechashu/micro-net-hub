package model

import "gorm.io/gorm"

type DnsRecord struct {
	gorm.Model
	Zone   string `gorm:"type:varchar(64)" json:"zone"`
	Type   string `gorm:"type:varchar(64)" json:"type"`
	Host   string `gorm:"type:varchar(64)" json:"host"`
	Value  string `gorm:"type:varchar(64)" json:"value"`
	Ttl    uint   `gorm:"type:bigint(20)" json:"ttl"`
	Status uint   `gorm:"type:uint(1)" json:"status"`
}
