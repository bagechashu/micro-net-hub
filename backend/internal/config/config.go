package config

import (
	"fmt"
	"micro-net-hub/internal/global"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

// 系统配置，对应yml
// viper内置了mapstructure, yml文件用"-"区分单词, 转为驼峰方便

// 全局配置变量
var Conf = new(config)

type config struct {
	System    *SystemConfig    `mapstructure:"system" json:"system"`
	Logs      *LogsConfig      `mapstructure:"logs" json:"logs"`
	Database  *Database        `mapstructure:"database" json:"database"`
	Mysql     *MysqlConfig     `mapstructure:"mysql" json:"mysql"`
	Jwt       *JwtConfig       `mapstructure:"jwt" json:"jwt"`
	RateLimit *RateLimitConfig `mapstructure:"rate-limit" json:"rateLimit"`
	Ldap      *LdapConfig      `mapstructure:"ldap" json:"ldap"`
	Radius    *RadiusConfig    `mapstructure:"radius" json:"radius"`
	Email     *EmailConfig     `mapstructure:"email" json:"email"`
	Notice    *NoticeConfig    `mapstructure:"notice" json:"notice"`
	Sync      *SyncConfig      `mapstructure:"sync" json:"sync"`
	DingTalk  *DingTalkConfig  `mapstructure:"dingtalk" json:"dingTalk"`
	WeCom     *WeComConfig     `mapstructure:"wecom" json:"weCom"`
	FeiShu    *FeiShuConfig    `mapstructure:"feishu" json:"feiShu"`
}

// 设置读取配置信息
func InitConfig() {
	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("读取应用目录失败:%s", err))
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/")

	// 在读取配置文件前设置默认值
	viper.SetDefault("sync.ldap-sync-time", "0 */2 * * * *")

	// 读取配置信息
	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("读取配置文件失败:%s", err))
	}

	// 将读取的配置信息保存至全局变量Conf
	if err := viper.Unmarshal(Conf); err != nil {
		panic(fmt.Errorf("初始化配置文件失败:%s", err))
	}
	// 读取rsa key
	Conf.System.RSAPublicBytes = RSAReadKeyFromFile(Conf.System.RSAPublicKey)
	Conf.System.RSAPrivateBytes = RSAReadKeyFromFile(Conf.System.RSAPrivateKey)

	// 热更新配置
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 将读取的配置信息保存至全局变量Conf
		if err := viper.Unmarshal(Conf); err != nil {
			panic(fmt.Errorf("热加载配置文件失败:%s", err))
		}
		// 读取rsa key
		Conf.System.RSAPublicBytes = RSAReadKeyFromFile(Conf.System.RSAPublicKey)
		Conf.System.RSAPrivateBytes = RSAReadKeyFromFile(Conf.System.RSAPrivateKey)
		global.Log.Info("热加载配置文件完成!")
	})
}

// 从文件中读取RSA key
func RSAReadKeyFromFile(filename string) []byte {
	f, err := os.Open(filename)
	var b []byte

	if err != nil {
		return b
	}
	defer f.Close()
	fileInfo, _ := f.Stat()
	b = make([]byte, fileInfo.Size())
	_, err = f.Read(b)
	if err != nil {
		return b
	}
	return b
}

type SystemConfig struct {
	Mode            string        `mapstructure:"mode" json:"mode"`
	UrlPathPrefix   string        `mapstructure:"url-path-prefix" json:"urlPathPrefix"`
	Host            string        `mapstructure:"host" json:"host"`
	Port            int           `mapstructure:"port" json:"port"`
	ReadTimeout     time.Duration `mapstructure:"read-timeout" json:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"write-timeout" json:"writeTimeout"`
	MaxHeaderMBytes int           `mapstructure:"max-header-MBytes" json:"maxHeaderMBytes"`
	InitData        bool          `mapstructure:"init-data" json:"initData"`
	RSAPublicKey    string        `mapstructure:"rsa-public-key" json:"rsaPublicKey"`
	RSAPrivateKey   string        `mapstructure:"rsa-private-key" json:"rsaPrivateKey"`
	RSAPublicBytes  []byte        `mapstructure:"-" json:"-"`
	RSAPrivateBytes []byte        `mapstructure:"-" json:"-"`
}

type LogsConfig struct {
	Level      zapcore.Level `mapstructure:"level" json:"level"`
	Path       string        `mapstructure:"path" json:"path"`
	MaxSize    int           `mapstructure:"max-size" json:"maxSize"`
	MaxBackups int           `mapstructure:"max-backups" json:"maxBackups"`
	MaxAge     int           `mapstructure:"max-age" json:"maxAge"`
	Compress   bool          `mapstructure:"compress" json:"compress"`
}

type Database struct {
	Driver string `mapstructure:"driver" json:"driver"`
	Source string `mapstructure:"source" json:"source"`
}

type MysqlConfig struct {
	Username    string `mapstructure:"username" json:"username"`
	Password    string `mapstructure:"password" json:"password"`
	Database    string `mapstructure:"database" json:"database"`
	Host        string `mapstructure:"host" json:"host"`
	Port        int    `mapstructure:"port" json:"port"`
	Query       string `mapstructure:"query" json:"query"`
	LogMode     bool   `mapstructure:"log-mode" json:"logMode"`
	LogLevel    int    `mapstructure:"log-level" json:"logLevel"`
	TablePrefix string `mapstructure:"table-prefix" json:"tablePrefix"`
	Charset     string `mapstructure:"charset" json:"charset"`
	Collation   string `mapstructure:"collation" json:"collation"`
}

type JwtConfig struct {
	Realm      string `mapstructure:"realm" json:"realm"`
	Key        string `mapstructure:"key" json:"key"`
	Timeout    int    `mapstructure:"timeout" json:"timeout"`
	MaxRefresh int    `mapstructure:"max-refresh" json:"maxRefresh"`
}

type RateLimitConfig struct {
	FillInterval int64 `mapstructure:"fill-interval" json:"fillInterval"`
	Capacity     int64 `mapstructure:"capacity" json:"capacity"`
}

type LdapConfig struct {
	Url                string `mapstructure:"url" json:"url"`
	MaxConn            int    `mapstructure:"max-conn" json:"maxConn"`
	BaseDN             string `mapstructure:"base-dn" json:"baseDN"`
	AdminDN            string `mapstructure:"admin-dn" json:"adminDN"`
	AdminPass          string `mapstructure:"admin-pass" json:"adminPass"`
	UserDN             string `mapstructure:"user-dn" json:"userDN"`
	UserInitPassword   string `mapstructure:"user-init-password" json:"userInitPassword"`
	GroupNameModify    bool   `mapstructure:"group-name-modify" json:"groupNameModify"`
	UserNameModify     bool   `mapstructure:"user-name-modify" json:"userNameModify"`
	DefaultEmailSuffix string `mapstructure:"default-email-suffix" json:"defaultEmailSuffix"`
}

type RadiusConfig struct {
	ListenAddr  string `mapstructure:"listen-addr" json:"listenAddr"`
	Secret      string `mapstructure:"secret" json:"secret"`
	GroupFilter string `mapstructure:"group-filter" json:"groupFilter"`
}

type EmailConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
	User string `mapstructure:"user" json:"user"`
	Pass string `mapstructure:"pass" json:"pass"`
}

type NoticeConfig struct {
	ProjectName              string `mapstructure:"project-name" json:"projectName"`
	ServiceDomain            string `mapstructure:"service-domain" json:"serviceDomain"`
	VPNServer                string `mapstructure:"vpn-server" json:"vpnServer"`
	DefaultNoticeSwitch      bool   `mapstructure:"default-notice-switch" json:"default-notice-switch"`
	DefaultNoticeRoleKeyword string `mapstructure:"default-notice-role-keyword" json:"defaultNoticeRoleKeyword"`
	HeaderHTML               string `mapstructure:"header-html" json:"headerHTML"`
	FooterHTML               string `mapstructure:"footer-html" json:"footerHTML"`
}

type SyncConfig struct {
	EnableSync    bool   `mapstructure:"enable-sync" json:"enableSync"`
	IsUpdateSyncd bool   `mapstructure:"is-update-syncd" json:"isUpdateSyncd"`
	UserSyncTime  string `mapstructure:"user-sync-time" json:"userSyncTime"`
	DeptSyncTime  string `mapstructure:"dept-sync-time" json:"deptSyncTime"`
	LdapSyncTime  string `mapstructure:"ldap-sync-time" json:"ldapSyncTime"`
}

type DingTalkConfig struct {
	AppKey      string   `mapstructure:"app-key" json:"appKey"`
	AppSecret   string   `mapstructure:"app-secret" json:"appSecret"`
	AgentId     string   `mapstructure:"agent-id" json:"agentId"`
	RootOuName  string   `mapstructure:"root-ou-name" json:"rootOuName"`
	Flag        string   `mapstructure:"flag" json:"flag"`
	DeptList    []string `mapstructure:"dept-list" json:"deptList"`
	ULeaveRange uint     `mapstructure:"user-leave-range" json:"userLevelRange"`
}

type WeComConfig struct {
	Flag       string `mapstructure:"flag" json:"flag"`
	CorpID     string `mapstructure:"corp-id" json:"corpId"`
	AgentID    int    `mapstructure:"agent-id" json:"agentId"`
	CorpSecret string `mapstructure:"corp-secret" json:"corpSecret"`
}

type FeiShuConfig struct {
	Flag      string   `mapstructure:"flag" json:"flag"`
	AppID     string   `mapstructure:"app-id" json:"appId"`
	AppSecret string   `mapstructure:"app-secret" json:"appSecret"`
	DeptList  []string `mapstructure:"dept-list" json:"deptList"`
}
