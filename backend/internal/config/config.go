package config

import (
	"fmt"
	"micro-net-hub/internal/global"
	"os"
	"path/filepath"
	"strings"
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
	System    *System    `mapstructure:"system" json:"system"`
	Logs      *Logs      `mapstructure:"logs" json:"logs"`
	Database  *Database  `mapstructure:"database" json:"database"`
	Mysql     *Mysql     `mapstructure:"mysql" json:"mysql"`
	Jwt       *Jwt       `mapstructure:"jwt" json:"jwt"`
	RateLimit *RateLimit `mapstructure:"rate-limit" json:"rateLimit"`
	Ldap      *Ldap      `mapstructure:"ldap" json:"ldap"`
	Radius    *Radius    `mapstructure:"radius" json:"radius"`
	Dns       *Dns       `mapstructure:"dns" json:"dns"`
	Email     *Email     `mapstructure:"email" json:"email"`
	Notice    *Notice    `mapstructure:"notice" json:"notice"`
	Sync      *Sync      `mapstructure:"sync" json:"sync"`
	DingTalk  *DingTalk  `mapstructure:"dingtalk" json:"dingTalk"`
	WeCom     *WeCom     `mapstructure:"wecom" json:"weCom"`
	FeiShu    *FeiShu    `mapstructure:"feishu" json:"feiShu"`
}

// ConfigDirName 默认配置目录名：config.yml、RSA 密钥、自定义 logo 等放在二进制文件所在目录下的 conf/，
// 方便 Docker 整个目录挂载（rotate-rsa-key 的备份也会落在这里）。可用启动参数 -conf 指定其他目录。
const ConfigDirName = "conf"

// 默认的 RSA 密钥文件名（相对配置文件所在目录）
const (
	DefaultRSAPublicKeyFile  = "auth-pub.pem"
	DefaultRSAPrivateKeyFile = "auth-priv-rsa.pem"
)

// configDir config.yml 所在目录，配置里的相对路径（RSA 密钥、site-logo）以它为准
var configDir string

// InitConfig 按默认位置查找并读取 config.yml，见 InitConfigFrom
func InitConfig() {
	InitConfigFrom("")
}

// InitConfigFrom 读取配置信息。
//
// dir 不为空（启动参数 -conf）时只读取 <dir>/config.yml。为空时依次查找：
//  1. 二进制文件所在目录下的 conf/config.yml（默认位置）
//  2. 二进制文件所在目录下的 config.yml（旧的部署方式）
//  3. 工作目录下的 conf/config.yml、config.yml（go run 时二进制在临时目录，开发时用这两项）
func InitConfigFrom(dir string) {
	viper.SetConfigType("yml")
	if dir != "" {
		viper.SetConfigFile(filepath.Join(dir, "config.yml"))
	} else {
		viper.SetConfigName("config")
		for _, p := range defaultSearchPaths() {
			viper.AddConfigPath(p)
		}
	}

	// 在读取配置文件前设置默认值
	viper.SetDefault("sync.ldap-sync-time", "0 */2 * * * *")
	viper.SetDefault("radius.fail-times-before-block5min", 9)
	viper.SetDefault("email.enable", false)
	viper.SetDefault("dns.listen-addr", "0.0.0.0:53")
	viper.SetDefault("dns.read-timeout-second", 5)
	viper.SetDefault("dns.write-timeout-second", 5)
	viper.SetDefault("dns.max-recursion-depth", 5)
	viper.SetDefault("dns.forward-addr", "1.1.1.1:53")
	viper.SetDefault("logs.audit-get-requests", true)

	// 读取配置信息
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("读取配置文件失败:%s", err))
	}

	// FIXME: 读取环境变量, 覆盖配置文件中的变量. 当前不生效.
	viper.AutomaticEnv()

	// 将读取的配置信息保存至全局变量Conf
	configDir = filepath.Dir(viper.ConfigFileUsed())
	if err := viper.Unmarshal(Conf); err != nil {
		panic(fmt.Errorf("初始化配置文件失败:%s", err))
	}
	resolvePaths()
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
		resolvePaths()
		global.LogLevel.SetLevel(Conf.Logs.Level)
		// 读取rsa key
		Conf.System.RSAPublicBytes = RSAReadKeyFromFile(Conf.System.RSAPublicKey)
		Conf.System.RSAPrivateBytes = RSAReadKeyFromFile(Conf.System.RSAPrivateKey)
		global.Log.Info("热加载配置文件完成!")
	})
}

// defaultSearchPaths 未指定配置目录时查找 config.yml 的目录（按优先级，去重）
func defaultSearchPaths() []string {
	var bases []string
	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		bases = append(bases, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		bases = append(bases, wd)
	}
	var paths []string
	seen := map[string]bool{}
	for _, b := range bases {
		for _, p := range []string{filepath.Join(b, ConfigDirName), b} {
			if !seen[p] {
				seen[p] = true
				paths = append(paths, p)
			}
		}
	}
	return paths
}

// resolvePaths 把配置里的文件路径（RSA 密钥、site-logo）解析成以 config.yml 所在目录为基准的路径。
// 旧部署的 config.yml 就在工作目录，解析结果和原来相同。
func resolvePaths() {
	sys := Conf.System
	if sys == nil {
		return
	}
	if strings.TrimSpace(sys.RSAPublicKey) == "" {
		sys.RSAPublicKey = DefaultRSAPublicKeyFile
	}
	if strings.TrimSpace(sys.RSAPrivateKey) == "" {
		sys.RSAPrivateKey = DefaultRSAPrivateKeyFile
	}
	sys.RSAPublicKey = ResolvePath(sys.RSAPublicKey)
	sys.RSAPrivateKey = ResolvePath(sys.RSAPrivateKey)
	sys.SiteLogo = ResolvePath(sys.SiteLogo)
}

// ResolvePath 相对路径以 config.yml 所在目录为基准；空串和绝对路径原样返回
func ResolvePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || filepath.IsAbs(p) || configDir == "" {
		return p
	}
	return filepath.Join(configDir, p)
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

type System struct {
	// 页面上显示的站点名称（顶部品牌、浏览器标题），为空时使用 DefaultSiteName
	SiteName string `mapstructure:"site-name" json:"siteName"`
	// 自定义 logo 图片文件（svg/png/jpg/webp/ico），相对路径以 config.yml 所在目录为准；为空用内置 logo
	SiteLogo        string        `mapstructure:"site-logo" json:"siteLogo"`
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

type Logs struct {
	Level            zapcore.Level `mapstructure:"level" json:"level"`
	Path             string        `mapstructure:"path" json:"path"`
	MaxSize          int           `mapstructure:"max-size" json:"maxSize"`
	MaxBackups       int           `mapstructure:"max-backups" json:"maxBackups"`
	MaxAge           int           `mapstructure:"max-age" json:"maxAge"`
	Compress         bool          `mapstructure:"compress" json:"compress"`
	AuditGetRequests bool          `mapstructure:"audit-get-requests" json:"auditGetRequests"`
}

type Database struct {
	Driver string `mapstructure:"driver" json:"driver"`
	Source string `mapstructure:"source" json:"source"`
}

type Mysql struct {
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

type Jwt struct {
	Realm         string `mapstructure:"realm" json:"realm"`
	Key           string `mapstructure:"key" json:"key"`
	TimeoutMin    int    `mapstructure:"timeout-min" json:"timeoutMin"`
	MaxRefreshMin int    `mapstructure:"max-refresh-min" json:"maxRefreshMin"`
}

type RateLimit struct {
	FillInterval int64 `mapstructure:"fill-interval" json:"fillInterval"`
	Capacity     int64 `mapstructure:"capacity" json:"capacity"`
}

type Ldap struct {
	EnableManage       bool   `mapstructure:"enable-manage" json:"enableManage"`
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

type Radius struct {
	FailTimesBeforeBlock5min int    `mapstructure:"fail-times-before-block5min" json:"failTimesBeforeBlock5min"`
	ListenAddr               string `mapstructure:"listen-addr" json:"listenAddr"`
	Secret                   string `mapstructure:"secret" json:"secret"`
	GroupFilter              string `mapstructure:"group-filter" json:"groupFilter"`
}

type Dns struct {
	ListenAddr         string `mapstructure:"listen-addr" json:"listenAddr"`
	ReadTimeoutSecond  int64  `mapstructure:"read-timeout-second" json:"readTimeoutSecond"`
	WriteTimeoutSecond int64  `mapstructure:"write-timeout-second" json:"writeTimeoutSecond"`
	MaxRecursionDepth  int    `mapstructure:"max-recursion-depth" json:"maxRecursionDepth"`
	ForwardAddr        string `mapstructure:"forward-addr" json:"forwardAddr"`
}

type Email struct {
	Enable bool   `mapstructure:"enable" json:"enable"`
	Host   string `mapstructure:"host" json:"host"`
	Port   int    `mapstructure:"port" json:"port"`
	User   string `mapstructure:"user" json:"user"`
	Pass   string `mapstructure:"pass" json:"pass"`
}

type Notice struct {
	ProjectName              string `mapstructure:"project-name" json:"projectName"`
	ServiceDomain            string `mapstructure:"service-domain" json:"serviceDomain"`
	VpnInfoSendSwitch        bool   `mapstructure:"vpn-info-send-switch" json:"vpnInfoSendSwitch"`
	VPNServer                string `mapstructure:"vpn-server" json:"vpnServer"`
	DefaultNoticeSwitch      bool   `mapstructure:"default-notice-switch" json:"default-notice-switch"`
	DefaultNoticeRoleKeyword string `mapstructure:"default-notice-role-keyword" json:"defaultNoticeRoleKeyword"`
	AccountCreatedNoticeDir  string `mapstructure:"account-created-notice-dir" json:"accountCreatedNoticeDir"`
	AccountCreatedNoticeSave bool   `mapstructure:"account-created-notice-save" json:"accountCreatedNoticeSave"`
	HeaderHTML               string `mapstructure:"header-html" json:"headerHTML"`
	FooterHTML               string `mapstructure:"footer-html" json:"footerHTML"`
}

type Sync struct {
	EnableSync    bool   `mapstructure:"enable-sync" json:"enableSync"`
	IsUpdateSyncd bool   `mapstructure:"is-update-syncd" json:"isUpdateSyncd"`
	UserSyncTime  string `mapstructure:"user-sync-time" json:"userSyncTime"`
	DeptSyncTime  string `mapstructure:"dept-sync-time" json:"deptSyncTime"`
	LdapSyncTime  string `mapstructure:"ldap-sync-time" json:"ldapSyncTime"`
}

type DingTalk struct {
	AppKey      string   `mapstructure:"app-key" json:"appKey"`
	AppSecret   string   `mapstructure:"app-secret" json:"appSecret"`
	AgentId     string   `mapstructure:"agent-id" json:"agentId"`
	RootOuName  string   `mapstructure:"root-ou-name" json:"rootOuName"`
	Flag        string   `mapstructure:"flag" json:"flag"`
	DeptList    []string `mapstructure:"dept-list" json:"deptList"`
	ULeaveRange uint     `mapstructure:"user-leave-range" json:"userLevelRange"`
}

type WeCom struct {
	Flag       string `mapstructure:"flag" json:"flag"`
	CorpID     string `mapstructure:"corp-id" json:"corpId"`
	AgentID    int    `mapstructure:"agent-id" json:"agentId"`
	CorpSecret string `mapstructure:"corp-secret" json:"corpSecret"`
}

type FeiShu struct {
	Flag      string   `mapstructure:"flag" json:"flag"`
	AppID     string   `mapstructure:"app-id" json:"appId"`
	AppSecret string   `mapstructure:"app-secret" json:"appSecret"`
	DeptList  []string `mapstructure:"dept-list" json:"deptList"`
}

// DefaultSiteName 未配置 system.site-name 时使用的站点名称
const DefaultSiteName = "Micro Net Hub"

// SiteName 返回页面上显示的站点名称
func SiteName() string {
	if Conf != nil && Conf.System != nil {
		if n := strings.TrimSpace(Conf.System.SiteName); n != "" {
			return n
		}
	}
	return DefaultSiteName
}

// siteLogoTypes 允许作为 logo 的图片类型
var siteLogoTypes = map[string]string{
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".ico":  "image/x-icon",
}

// siteLogoMaxBytes logo 文件大小上限
const siteLogoMaxBytes = 2 << 20

// SiteLogoFile 返回配置的自定义 logo 文件；未配置、文件不存在、类型或大小不符合时 ok 为 false
func SiteLogoFile() (path string, contentType string, modTime time.Time, ok bool) {
	if Conf == nil || Conf.System == nil {
		return "", "", time.Time{}, false
	}
	p := strings.TrimSpace(Conf.System.SiteLogo)
	if p == "" {
		return "", "", time.Time{}, false
	}
	p = filepath.Clean(p)
	ct, allowed := siteLogoTypes[strings.ToLower(filepath.Ext(p))]
	if !allowed {
		return "", "", time.Time{}, false
	}
	fi, err := os.Stat(p)
	if err != nil || !fi.Mode().IsRegular() || fi.Size() > siteLogoMaxBytes {
		return "", "", time.Time{}, false
	}
	return p, ct, fi.ModTime(), true
}
