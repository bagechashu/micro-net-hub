package user

import jsoniter "github.com/json-iterator/go"

var (
	UserLogicIns   = &UserLogic{}
	GroupLogicIns  = &GroupLogic{}
	RoleLogicIns   = &RoleLogic{}
	MenuLogicIns   = &MenuLogic{}
	SqlLogicIns    = &SqlLogic{}
	PasswdLogicIns = &PasswdLogic{}

	// OpenLdapLogicIns = &OpenLdapLogic{}
	// DingTalkLogicIns = &DingTalkLogic{}
	// FeiShuLogicIns   = &FeiShuLogic{}
	// WeComLogicIns    = &WeComLogic{}

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)
