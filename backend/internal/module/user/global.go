package user

import jsoniter "github.com/json-iterator/go"

// TODO: user's model & logic change to better organization
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
