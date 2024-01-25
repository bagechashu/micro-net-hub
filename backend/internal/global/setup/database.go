package setup

import (
	"fmt"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	apiMgrModel "micro-net-hub/internal/module/apimgr/model"
	fieldRelationModel "micro-net-hub/internal/module/goldap/field_relation/model"
	opLogModel "micro-net-hub/internal/module/operation_log/model"
	siteNavModel "micro-net-hub/internal/module/sitenav/model"
	totpModel "micro-net-hub/internal/module/totp/model"
	userModel "micro-net-hub/internal/module/user/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

// 初始化数据库
func InitDB() {
	switch config.Conf.Database.Driver {
	case "mysql":
		global.DB = ConnMysql()
	case "sqlite3":
		global.DB = ConnSqlite()
	}
	dbAutoMigrate()
}

// 自动迁移表结构
func dbAutoMigrate() {
	_ = global.DB.AutoMigrate(
		&userModel.User{},
		&userModel.Role{},
		&userModel.Group{},
		&userModel.Menu{},
		&totpModel.Totp{},
		&apiMgrModel.Api{},
		&opLogModel.OperationLog{},
		&fieldRelationModel.FieldRelation{},
		&siteNavModel.NavGroup{},
		&siteNavModel.NavSite{},
	)
}

func ConnSqlite() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(config.Conf.Database.Source), &gorm.Config{
		// 禁用外键(指定外键时不会在mysql创建真实的外键约束)
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		global.Log.Panicf("failed to connect sqlite3: %v", err)
	}
	dbObj, err := db.DB()
	if err != nil {
		global.Log.Panicf("failed to get sqlite3 obj: %v", err)
	}
	// 参见： https://github.com/glebarez/sqlite/issues/52
	dbObj.SetMaxOpenConns(1)
	return db
}

func ConnMysql() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&%s",
		config.Conf.Mysql.Username,
		config.Conf.Mysql.Password,
		config.Conf.Mysql.Host,
		config.Conf.Mysql.Port,
		config.Conf.Mysql.Database,
		config.Conf.Mysql.Charset,
		config.Conf.Mysql.Collation,
		config.Conf.Mysql.Query,
	)
	// 隐藏密码
	showDsn := fmt.Sprintf(
		"%s:******@tcp(%s:%d)/%s?charset=%s&collation=%s&%s",
		config.Conf.Mysql.Username,
		config.Conf.Mysql.Host,
		config.Conf.Mysql.Port,
		config.Conf.Mysql.Database,
		config.Conf.Mysql.Charset,
		config.Conf.Mysql.Collation,
		config.Conf.Mysql.Query,
	)

	gormConf := &gorm.Config{
		// 禁用外键(指定外键时不会在mysql创建真实的外键约束)
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 开启mysql日志
	if config.Conf.Mysql.LogMode {
		logger := zapgorm2.New(global.BasicLog)
		logger.SetAsDefault()

		var gormLogLevel gormlogger.LogLevel
		if config.Conf.Mysql.LogLevel < 1 || config.Conf.Mysql.LogLevel > 4 {
			gormLogLevel = gormlogger.LogLevel(3)
		} else {
			gormLogLevel = gormlogger.LogLevel(config.Conf.Mysql.LogLevel)
		}

		global.Log.Debugf("mysql log level config: %d, effect: %d", config.Conf.Mysql.LogLevel, gormLogLevel)

		gormConf.Logger = logger.LogMode(gormLogLevel) // 4: info; 3: warn; 2: error; 1: silent
	}

	global.Log.Debugf("gorm Config: %+v", gormConf)

	db, err := gorm.Open(mysql.Open(dsn), gormConf)
	if err != nil {
		global.Log.Panicf("初始化mysql数据库异常: %v", err)
	}
	global.Log.Infof("初始化mysql数据库完成! dsn: %s", showDsn)
	return db
}
