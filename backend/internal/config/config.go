package config

import (
	"fmt"
	"os"
	"time"

	"micro-net-hub/internal/global"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

// 系统配置，对应yml
// viper内置了mapstructure, yml文件用"-"区分单词, 转为驼峰方便

// 全局配置变量
var Conf = new(config)

type config struct {
	System     *System     `mapstructure:"system" json:"system"`
	Logs       *Logs       `mapstructure:"logs" json:"logs"`
	Database   *Database   `mapstructure:"database" json:"database"`
	Mysql      *Mysql      `mapstructure:"mysql" json:"mysql"`
	Jwt        *Jwt        `mapstructure:"jwt" json:"jwt"`
	RateLimit  *RateLimit  `mapstructure:"rate-limit" json:"rateLimit"`
	Ldap       *Ldap       `mapstructure:"ldap" json:"ldap"`
	LdapServer *LdapServer `mapstructure:"ldap-server" json:"ldapServer"`
	Radius     *Radius     `mapstructure:"radius" json:"radius"`
	Dns        *Dns        `mapstructure:"dns" json:"dns"`
	Email      *Email      `mapstructure:"email" json:"email"`
	Notice     *Notice     `mapstructure:"notice" json:"notice"`
	Sync       *Sync       `mapstructure:"sync" json:"sync"`
	DingTalk   *DingTalk   `mapstructure:"dingtalk" json:"dingTalk"`
	WeCom      *WeCom      `mapstructure:"wecom" json:"weCom"`
	FeiShu     *FeiShu     `mapstructure:"feishu" json:"feiShu"`
	Bot        *Bot        `mapstructure:"bot" json:"bot"`
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
	setViperDefaults()

	// 读取配置信息
	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("读取配置文件失败:%s", err))
	}

	// FIXME: 读取环境变量, 覆盖配置文件中的变量. 当前不生效.
	viper.AutomaticEnv()

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

// setViperDefaults 注册配置项的默认值, 供 InitConfig 在读取配置文件前调用.
//
// 独立成函数: 默认值键必须与 mapstructure 标签一一对应, 抽出来后可被测试直接校验,
// 避免"键名写错导致默认值静默失效".
func setViperDefaults() {
	viper.SetDefault("sync.ldap-sync-time", "0 */2 * * * *")
	viper.SetDefault("radius.fail-times-before-block5min", 9)
	viper.SetDefault("bot.enable", false)
	viper.SetDefault("radius.approval.enable", false)
	viper.SetDefault("radius.approval.timezone", "Asia/Shanghai")
	viper.SetDefault("radius.approval.wait-seconds", 6)
	viper.SetDefault("radius.approval.pending-ttl-seconds", 120)
	viper.SetDefault("radius.approval.notice-interval-seconds", 15)
	viper.SetDefault("radius.approval.reject-cooldown-seconds", 60)
	viper.SetDefault("radius.approval.grant-ttl-minutes", 30)
	viper.SetDefault("radius.approval.grant-max-uses", 0)
	viper.SetDefault("radius.approval.max-pending-global", 200)
	viper.SetDefault("radius.approval.cleanup-cron", "0 30 4 * * *")
	viper.SetDefault("radius.approval.bot.enable", true)
	viper.SetDefault("radius.approval.bot.private-only", true)
	viper.SetDefault("radius.approval.bot.applicant-notify", true)
	viper.SetDefault("radius.approval.bot.max-commands-per10s", 5)
	viper.SetDefault("email.enable", false)
	viper.SetDefault("ldap-server.enable-manage", false)
	viper.SetDefault("ldap-server.listen-addr", "0.0.0.0:1389")
	viper.SetDefault("ldap-server.totp-enable", true)
	viper.SetDefault("dns.listen-addr", "0.0.0.0:53")
	viper.SetDefault("dns.read-timeout-second", 5)
	viper.SetDefault("dns.write-timeout-second", 5)
	viper.SetDefault("dns.max-recursion-depth", 5)
	viper.SetDefault("dns.forward-addr", "1.1.1.1:53")
	viper.SetDefault("logs.audit-get-requests", true)
	viper.SetDefault("ldap-server.binddn-role-keyword", "binddn")
}

// 从文件中读取RSA key
func RSAReadKeyFromFile(filename string) []byte {
	f, err := os.Open(filename)
	var b []byte

	if err != nil {
		return b
	}
	defer func() { _ = f.Close() }()
	fileInfo, _ := f.Stat()
	b = make([]byte, fileInfo.Size())
	_, err = f.Read(b)
	if err != nil {
		return b
	}
	return b
}

type System struct {
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

type LdapServer struct {
	ListenAddr               string `mapstructure:"listen-addr" json:"listenAddr"`
	ListenAddrWithTotpVerify string `mapstructure:"listen-addr-with-totp-verify" json:"listenAddrWithTotpVerify"`
	// TlsEnable  bool   `mapstructure:"tls-enable" json:"tlsEnable"`
	BaseDN            string `mapstructure:"base-dn" json:"baseDN"`
	BindDNRoleKeyword string `mapstructure:"binddn-role-keyword" json:"bindDNRoleKeyword"`
	TotpEnable        bool   `mapstructure:"totp-enable" json:"totpEnable"`
}

type Radius struct {
	FailTimesBeforeBlock5min int             `mapstructure:"fail-times-before-block5min" json:"failTimesBeforeBlock5min"`
	ListenAddr               string          `mapstructure:"listen-addr" json:"listenAddr"`
	Secret                   string          `mapstructure:"secret" json:"secret"`
	GroupFilter              string          `mapstructure:"group-filter" json:"groupFilter"`
	Approval                 *RadiusApproval `mapstructure:"approval" json:"approval"`
}

// Bot 通用 Bot 接入配置.
//
// 每个实例独立启停, 由 bot.Manager 统一托管, 任一实例启动失败不影响其它实例.
type Bot struct {
	Enable    bool          `mapstructure:"enable" json:"enable"`
	Instances []BotInstance `mapstructure:"instances" json:"instances"`
}

// BotInstance 一个 Bot 实例的配置.
//
// Config 为类型专属配置项(如 telegram 的 token), 具体必填项由该类型的 TypeSpec 约束.
type BotInstance struct {
	ID     string            `mapstructure:"id" json:"id"`
	Type   string            `mapstructure:"type" json:"type"`
	Name   string            `mapstructure:"name" json:"name"`
	Enable bool              `mapstructure:"enable" json:"enable"`
	Config map[string]string `mapstructure:"config" json:"config"`
}

// RadiusApproval RADIUS 认证人工审批配置.
//
// 命中时间窗口且在适用范围内的认证请求, 需人工审批通过后方可放行; 审批可通过 Bot 完成.
type RadiusApproval struct {
	// 是否开启人工审批
	Enable bool `mapstructure:"enable" json:"enable"`
	// 时间窗口判定使用的时区
	Timezone string `mapstructure:"timezone" json:"timezone"`
	// 需要审批的时间窗口列表, 命中任意一个即需要审批; 留空表示全天都需要审批
	TimeWindows []ApprovalTimeWindow `mapstructure:"time-windows" json:"timeWindows"`
	// 审批适用的用户范围
	Scope ApprovalScope `mapstructure:"scope" json:"scope"`
	// 审批的 Bot 渠道配置
	Bot ApprovalBot `mapstructure:"bot" json:"bot"`
	// RADIUS 侧同步等待审批结果的秒数, 超时后本次认证返回拒绝(审批通过后用户重连即可放行)
	WaitSeconds int `mapstructure:"wait-seconds" json:"waitSeconds"`
	// 申请单有效期(秒), 超时后申请单置为已过期
	PendingTTLSeconds int `mapstructure:"pending-ttl-seconds" json:"pendingTTLSeconds"`
	// 同一用户名重复申请时, 通知审批人的最小间隔(秒)
	NoticeIntervalSeconds int `mapstructure:"notice-interval-seconds" json:"noticeIntervalSeconds"`
	// 被拒绝后该用户的申请冷却时间(秒), 冷却期内不重复通知审批人
	RejectCooldownSeconds int `mapstructure:"reject-cooldown-seconds" json:"rejectCooldownSeconds"`
	// 审批通过后放行凭证的有效期(分钟)
	GrantTTLMinutes int `mapstructure:"grant-ttl-minutes" json:"grantTTLMinutes"`
	// 放行凭证在有效期内的最大使用次数, 0 表示不限次
	GrantMaxUses int `mapstructure:"grant-max-uses" json:"grantMaxUses"`
	// 全局待审批申请数量上限, 超出后新的申请直接拒绝, 避免申请风暴
	MaxPendingGlobal int `mapstructure:"max-pending-global" json:"maxPendingGlobal"`
	// 过期申请与凭证的清理任务执行时间(cron 秒级)
	CleanupCron string `mapstructure:"cleanup-cron" json:"cleanupCron"`
}

// ApprovalTimeWindow 一个审批时间窗口.
//
// Days 取值 mon/tue/wed/thu/fri/sat/sun, 留空表示每天; Start 与 End 为 HH:MM,
// 当 End 早于 Start 时视为跨天窗口(如 22:00 - 06:00).
type ApprovalTimeWindow struct {
	Days  []string `mapstructure:"days" json:"days"`
	Start string   `mapstructure:"start" json:"start"`
	End   string   `mapstructure:"end" json:"end"`
}

// ApprovalScope 审批适用的用户范围, 各字段为空时表示不限制该维度.
type ApprovalScope struct {
	Roles        []string `mapstructure:"roles" json:"roles"`
	Groups       []string `mapstructure:"groups" json:"groups"`
	Users        []string `mapstructure:"users" json:"users"`
	ExcludeUsers []string `mapstructure:"exclude-users" json:"excludeUsers"`
}

// ApprovalBot 审批的 Bot 渠道配置.
//
// 审批通知以私聊方式发送给 ApproverChatIDs 中的审批人, 审批人需先与机器人建立会话,
// 且仅这些会话的指令会被执行.
type ApprovalBot struct {
	// 是否开启 Bot 审批
	Enable bool `mapstructure:"enable" json:"enable"`
	// 参与审批的 Bot 实例 ID 列表, 对应 bot.instances 中的 id
	InstanceIDs []string `mapstructure:"instance-ids" json:"instanceIds"`
	// 审批人会话 ID 列表(telegram 的 chat id), 即审批人白名单
	ApproverChatIDs []string `mapstructure:"approver-chat-ids" json:"approverChatIds"`
	// 是否仅允许私聊会话审批(安全边界, 建议保持开启)
	PrivateOnly bool `mapstructure:"private-only" json:"privateOnly"`
	// 是否向申请人发送"申请已提交"与审批结果通知
	ApplicantNotify bool `mapstructure:"applicant-notify" json:"applicantNotify"`
	// 申请人 Bot 会话映射: 平台用户名 -> Bot chat id, 用于向申请人回执
	ApplicantMap map[string]string `mapstructure:"applicant-map" json:"applicantMap"`
	// 单个审批人 10 秒内可执行的指令条数上限
	MaxCommandsPer10s int `mapstructure:"max-commands-per10s" json:"maxCommandsPer10s"`
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
