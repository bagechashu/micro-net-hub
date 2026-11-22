package global

import (
	"micro-net-hub/internal/pkg/ldappool"

	"github.com/casbin/casbin/v3"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"

	ut "github.com/go-playground/universal-translator"
)

// 全局CasbinEnforcer
var CasbinEnforcer *casbin.Enforcer

// 全局数据库对象
var DB *gorm.DB

// 全局日志变量
// var Log *zap.Logger
var Log *zap.SugaredLogger
var BasicLog *zap.Logger

// LogLevel 日志级别（对应 logs.level），修改后立即生效
var LogLevel = zap.NewAtomicLevel()

// 全局Validate数据校验实列
var Validate *validator.Validate

// 全局翻译器
var Trans ut.Translator

// Global LdapPool
var LdapPool ldappool.LdapPool
