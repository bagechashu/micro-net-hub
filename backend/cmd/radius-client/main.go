// Package main 是 micro-net-hub 内置 RADIUS 服务的命令行测试客户端.
//
// 相当于 freeradius-utils 提供的 radtest, 但无需额外安装依赖, 并针对本项目的认证约定做了适配:
//
//   - 发送的密码为 "数据库密码 + TOTP 六位验证码", 与 internal/radiussrv.AuthRequest 的解析逻辑一致
//   - 可携带 NAS-Identifier / NAS-IP-Address 等属性, 便于验证人工审批单上的留痕信息
//   - 支持连续发送多次请求, 便于验证 "失败次数锁定 5 分钟" 与 "人工审批" 这两类流程
//
// 常用示例:
//
//	# 基本认证(密码 + TOTP), 服务端默认监听 127.0.0.1:1812
//	go run ./cmd/radius-client -user admin -pass admin_pass -otp 000000
//
//	# 指定服务端与密钥, 并打印请求/响应的全部属性
//	go run ./cmd/radius-client -server 127.0.0.1:1812 -secret default-radius-secret \
//	    -user admin -pass admin_pass -otp 000000 -nas-id ocserv-1 -verbose
//
//	# 连续 3 次错误密码, 观察服务端的 5 分钟锁定行为
//	go run ./cmd/radius-client -user admin -pass wrong_pass -otp 000000 -count 3 -interval 1s
//
// 退出码约定: 0=Access-Accept, 1=Access-Reject, 2=请求失败(无响应/超时/报文非法), 便于脚本判断.
package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	// defaultServer RADIUS 服务端地址, 对应 config.yml 中 radius.listen-addr 的默认值
	defaultServer = "127.0.0.1:1812"
	// defaultPort 未显式指定端口时使用的 RADIUS 认证端口
	defaultPort = "1812"
	// defaultSecret 共享密钥, 对应 config.yml 中 radius.secret 的默认值
	defaultSecret = "default-radius-secret"
	// defaultNASID NAS-Identifier 默认值, 用于在审批单上区分请求来源
	defaultNASID = "micro-net-hub-radius-client"
	// defaultTimeout 单次请求总超时, 需大于 config.yml 中 radius.approval.wait-seconds(默认 6s),
	// 否则开启人工审批时无法观察到审批通过/拒绝的最终结果
	defaultTimeout = 30 * time.Second
	// otpLength TOTP 验证码长度, 服务端固定取密码的最后 6 位作为 TOTP
	otpLength = 6
)

// 退出码约定
const (
	exitAccept    = 0
	exitReject    = 1
	exitTransport = 2
)

// options 命令行参数
type options struct {
	server   string        // RADIUS 服务端地址
	secret   string        // RADIUS 共享密钥
	user     string        // 用户名
	pass     string        // 数据库密码, 不含 TOTP
	otp      string        // TOTP 六位验证码
	nasID    string        // NAS-Identifier 属性
	nasIP    string        // NAS-IP-Address 属性, 留空时自动探测本机出接口地址
	nasPort  uint          // NAS-Port 属性
	timeout  time.Duration // 单次请求总超时
	retry    time.Duration // 报文重传间隔, 0 表示不重传
	count    int           // 请求次数
	interval time.Duration // 多次请求之间的间隔
	verbose  bool          // 是否打印报文属性
}

// password 返回实际发送给服务端的密码: 数据库密码 + TOTP 验证码.
//
// 服务端取最后 6 位作为 TOTP, 其余部分作为数据库密码, 因此这里必须直接拼接而不做分隔.
func (o options) password() string {
	return o.pass + o.otp
}

// validate 校验参数, 返回 nil 表示参数可用
func (o options) validate() error {
	if strings.TrimSpace(o.user) == "" {
		return errors.New("必须指定用户名, 使用 -user 或第一个位置参数")
	}
	if o.pass == "" {
		return errors.New("必须指定密码, 使用 -pass 或第二个位置参数")
	}
	if o.otp != "" && (len(o.otp) != otpLength || strings.Trim(o.otp, "0123456789") != "") {
		return fmt.Errorf("-otp 必须是 %d 位数字, 当前为 %q", otpLength, o.otp)
	}
	if o.count < 1 {
		return errors.New("-count 必须大于 0")
	}
	if o.timeout <= 0 {
		return errors.New("-timeout 必须大于 0")
	}
	if o.interval < 0 {
		return errors.New("-interval 不能为负数")
	}
	if o.retry < 0 {
		return errors.New("-retry 不能为负数")
	}
	return nil
}

// attemptResult 单次认证请求的结果
type attemptResult struct {
	request  *radius.Packet // 发出的 Access-Request
	response *radius.Packet // 收到的响应, 请求失败时为 nil
	code     radius.Code    // 响应码
	elapsed  time.Duration  // 耗时
	err      error          // 传输层错误(无响应/超时/报文非法)
}

// accepted 是否收到 Access-Accept
func (r attemptResult) accepted() bool {
	return r.err == nil && r.code == radius.CodeAccessAccept
}

func main() {
	opts := parseFlags()
	applyPositionalArgs(&opts)

	if err := opts.validate(); err != nil {
		fmt.Fprintf(os.Stderr, "参数错误: %v\n\n", err)
		flag.Usage()
		os.Exit(exitTransport)
	}

	server, err := normalizeServerAddr(opts.server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "参数错误: %v\n\n", err)
		flag.Usage()
		os.Exit(exitTransport)
	}

	// 支持 Ctrl+C 中断, 避免连续请求场景下必须等完全部次数
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("目标: %s  用户: %s  密码: %s(含 TOTP)  NAS-Identifier: %s  请求次数: %d\n",
		server, opts.user, maskPassword(opts.password()), opts.nasID, opts.count)

	results := make([]attemptResult, 0, opts.count)
	for i := 1; i <= opts.count; i++ {
		if i > 1 && !sleep(ctx, opts.interval) {
			break
		}
		if ctx.Err() != nil {
			fmt.Println("已中断, 停止后续请求")
			break
		}

		result := exchange(ctx, opts, server)
		results = append(results, result)
		reportAttempt(opts, i, result)
	}

	reportSummary(results)
	os.Exit(exitCode(results))
}

// parseFlags 注册并解析命令行参数
func parseFlags() options {
	var opts options

	flag.StringVar(&opts.server, "server", defaultServer, "RADIUS 服务端地址, 支持 host 或 host:port")
	flag.StringVar(&opts.secret, "secret", defaultSecret, "RADIUS 共享密钥, 需与 config.yml 中 radius.secret 一致")
	flag.StringVar(&opts.user, "user", "", "用户名, 也可作为第一个位置参数传入")
	flag.StringVar(&opts.pass, "pass", "", "数据库密码(不含 TOTP), 也可作为第二个位置参数传入")
	flag.StringVar(&opts.otp, "otp", "", "TOTP 六位验证码, 会直接拼接到 -pass 之后")
	flag.StringVar(&opts.nasID, "nas-id", defaultNASID, "NAS-Identifier 属性, 留空表示不发送")
	flag.StringVar(&opts.nasIP, "nas-ip", "", "NAS-IP-Address 属性, 留空时自动探测本机出接口地址")
	flag.UintVar(&opts.nasPort, "nas-port", 0, "NAS-Port 属性, 0 表示不发送")
	flag.DurationVar(&opts.timeout, "timeout", defaultTimeout, "单次请求总超时, 需大于 radius.approval.wait-seconds")
	flag.DurationVar(&opts.retry, "retry", 0, "报文重传间隔, 0 表示不重传; 设为 1s 可模拟 ocserv/radcli 的重试行为")
	flag.IntVar(&opts.count, "count", 1, "发送请求的次数, 用于验证失败次数锁定或人工审批流程")
	flag.DurationVar(&opts.interval, "interval", time.Second, "多次请求之间的间隔")
	flag.BoolVar(&opts.verbose, "verbose", false, "打印请求与响应的全部属性")

	flag.Usage = usage
	flag.Parse()

	return opts
}

// applyPositionalArgs 支持 radtest 风格的位置参数: radius-client [用户名] [密码]
func applyPositionalArgs(opts *options) {
	args := flag.Args()
	if opts.user == "" && len(args) > 0 {
		opts.user = args[0]
	}
	if opts.pass == "" && len(args) > 1 {
		opts.pass = args[1]
	}
}

// usage 打印帮助信息
func usage() {
	name := filepath.Base(os.Args[0])
	out := flag.CommandLine.Output()

	// 帮助信息的输出失败无需中断流程, 因此统一忽略写错误
	_, _ = fmt.Fprintf(out, "radius-client: micro-net-hub 内置 RADIUS 服务的测试客户端\n\n")
	_, _ = fmt.Fprintf(out, "用法: %s [选项] [用户名] [密码]\n\n", name)
	_, _ = fmt.Fprintln(out, "选项:")
	flag.PrintDefaults()
	_, _ = fmt.Fprintf(out, "\n示例:\n")
	_, _ = fmt.Fprintf(out, "  %s -user admin -pass admin_pass -otp 000000\n", name)
	_, _ = fmt.Fprintf(out, "  %s admin admin_pass000000 -verbose\n", name)
	_, _ = fmt.Fprintf(out, "\n退出码: 0=Access-Accept, 1=Access-Reject, 2=请求失败(无响应/超时/报文非法)\n")
}

// normalizeServerAddr 补全端口并规范化地址: 支持 "host" 与 "host:port" 两种写法
func normalizeServerAddr(addr string) (string, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", errors.New("RADIUS 服务端地址不能为空")
	}

	// 裸 IPv6 地址(如 ::1)会被 SplitHostPort 报为 "too many colons", 先单独处理
	if ip := net.ParseIP(strings.Trim(addr, "[]")); ip != nil {
		return net.JoinHostPort(ip.String(), defaultPort), nil
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if !strings.Contains(err.Error(), "missing port") {
			return "", fmt.Errorf("RADIUS 服务端地址 %q 非法: %w", addr, err)
		}
		// 未带端口: 补默认端口
		return net.JoinHostPort(addr, defaultPort), nil
	}
	if port == "" {
		port = defaultPort
	}
	return net.JoinHostPort(host, port), nil
}

// detectNasIP 探测本机到 RADIUS 服务端的出接口地址, 作为 NAS-IP-Address 的默认值.
//
// 探测失败返回 nil, 此时不发送 NAS-IP-Address 属性(服务端允许该属性缺失).
func detectNasIP(server string) net.IP {
	conn, err := net.Dial("udp", server)
	if err != nil {
		return nil
	}
	defer func() { _ = conn.Close() }()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil
	}
	return addr.IP.To4()
}

// buildAccessRequest 组装 Access-Request 报文, 属性组合参照 ocserv + radcli 的真实请求
func buildAccessRequest(opts options, server string) (*radius.Packet, error) {
	packet := radius.New(radius.CodeAccessRequest, []byte(opts.secret))

	if err := rfc2865.UserName_SetString(packet, opts.user); err != nil {
		return nil, fmt.Errorf("设置 User-Name 失败: %w", err)
	}
	if err := rfc2865.UserPassword_SetString(packet, opts.password()); err != nil {
		return nil, fmt.Errorf("设置 User-Password 失败: %w", err)
	}

	if opts.nasID != "" {
		if err := rfc2865.NASIdentifier_SetString(packet, opts.nasID); err != nil {
			return nil, fmt.Errorf("设置 NAS-Identifier 失败: %w", err)
		}
	}

	if err := addNASIPAddress(packet, opts, server); err != nil {
		return nil, err
	}

	if opts.nasPort > 0 {
		if err := rfc2865.NASPort_Add(packet, rfc2865.NASPort(opts.nasPort)); err != nil {
			return nil, fmt.Errorf("设置 NAS-Port 失败: %w", err)
		}
	}

	// 固定携带 ocserv 场景下的典型 NAS 属性, 使请求与真实客户端保持一致
	if err := rfc2865.NASPortType_Add(packet, rfc2865.NASPortType_Value_Virtual); err != nil {
		return nil, fmt.Errorf("设置 NAS-Port-Type 失败: %w", err)
	}
	if err := rfc2865.ServiceType_Add(packet, rfc2865.ServiceType_Value_LoginUser); err != nil {
		return nil, fmt.Errorf("设置 Service-Type 失败: %w", err)
	}
	if err := rfc2865.FramedProtocol_Add(packet, rfc2865.FramedProtocol_Value_PPP); err != nil {
		return nil, fmt.Errorf("设置 Framed-Protocol 失败: %w", err)
	}

	return packet, nil
}

// addNASIPAddress 写入 NAS-IP-Address: 未显式指定时自动探测本机出接口地址, 探测失败则跳过该属性
func addNASIPAddress(packet *radius.Packet, opts options, server string) error {
	if opts.nasIP == "" {
		detected := detectNasIP(server)
		if detected == nil {
			return nil
		}
		if err := rfc2865.NASIPAddress_Add(packet, detected); err != nil {
			return fmt.Errorf("设置 NAS-IP-Address 失败: %w", err)
		}
		return nil
	}

	ip := net.ParseIP(opts.nasIP)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("-nas-ip %q 不是合法的 IPv4 地址", opts.nasIP)
	}
	if err := rfc2865.NASIPAddress_Add(packet, ip.To4()); err != nil {
		return fmt.Errorf("设置 NAS-IP-Address 失败: %w", err)
	}
	return nil
}

// exchange 发送一次认证请求并等待响应
func exchange(ctx context.Context, opts options, server string) attemptResult {
	packet, err := buildAccessRequest(opts, server)
	if err != nil {
		return attemptResult{err: err}
	}
	result := attemptResult{request: packet}

	client := &radius.Client{
		Retry:           opts.retry,
		MaxPacketErrors: 10,
	}

	// 每次请求独立超时: 连续请求时单次超时不应吃掉后续请求的时间预算
	reqCtx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	start := time.Now()
	response, err := client.Exchange(reqCtx, packet, server)
	result.elapsed = time.Since(start)
	if err != nil {
		result.err = err
		return result
	}

	result.code = response.Code
	result.response = response
	return result
}

// sleep 等待 d 时长, 被中断时返回 false
func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// reportAttempt 打印单次请求的结果
func reportAttempt(opts options, index int, result attemptResult) {
	prefix := fmt.Sprintf("[%d/%d]", index, opts.count)
	elapsed := result.elapsed.Round(time.Millisecond)

	if opts.verbose && result.request != nil {
		fmt.Printf("%s 请求属性:\n", prefix)
		dumpAttributes(result.request)
	}

	switch {
	case result.err != nil:
		fmt.Printf("%s 请求失败(%s): %v\n", prefix, elapsed, result.err)
		printFailureHint(result.err)
	case result.accepted():
		fmt.Printf("%s Access-Accept (%s)\n", prefix, elapsed)
	default:
		fmt.Printf("%s %s (%s)\n", prefix, result.code, elapsed)
		printRejectHint()
	}

	if result.response == nil {
		return
	}
	if msg := rfc2865.ReplyMessage_GetString(result.response); msg != "" {
		fmt.Printf("%s Reply-Message: %s\n", prefix, msg)
	}
	if opts.verbose {
		fmt.Printf("%s 响应属性:\n", prefix)
		dumpAttributes(result.response)
	}
}

// printRejectHint 认证被拒时给出排查方向.
//
// 服务端只返回 Access-Reject, 具体原因记录在 micro-net-hub 日志中.
func printRejectHint() {
	fmt.Println("      提示: 服务端不返回具体拒绝原因, 请在 micro-net-hub 日志中搜索" +
		" \"Error while performing auth for user\"; 常见原因: 密码错误 / TOTP 错误或未绑定 / " +
		"失败次数过多被锁定 5 分钟 / BindDN 角色禁止登录 / 人工审批未通过或等待超时")
}

// printFailureHint 请求失败时给出排查方向
func printFailureHint(err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Println("      提示: 等待响应超时, 请确认服务端已启动并监听该地址, 且共享密钥与 config.yml 中" +
			" radius.secret 一致; 若已开启人工审批, -timeout 需大于 radius.approval.wait-seconds")
	case errors.Is(err, context.Canceled):
		fmt.Println("      提示: 请求已被中断")
	default:
		fmt.Println("      提示: 请确认服务端地址可达、UDP 1812 端口未被防火墙拦截")
	}
}

// reportSummary 打印汇总结果
func reportSummary(results []attemptResult) {
	var accepted, rejected, failed int
	for _, result := range results {
		switch {
		case result.err != nil:
			failed++
		case result.accepted():
			accepted++
		default:
			rejected++
		}
	}

	fmt.Printf("汇总: 共 %d 次请求, Access-Accept %d 次, Access-Reject %d 次, 请求失败 %d 次\n",
		len(results), accepted, rejected, failed)
}

// exitCode 汇总多次请求的结果作为进程退出码:
// 出现传输层错误优先返回 2; 全部 Access-Accept 返回 0; 其余返回 1.
func exitCode(results []attemptResult) int {
	allAccepted := len(results) > 0
	for _, result := range results {
		if result.err != nil {
			return exitTransport
		}
		if !result.accepted() {
			allAccepted = false
		}
	}
	if allAccepted {
		return exitAccept
	}
	return exitReject
}

// maskPassword 对密码做脱敏, 避免明文出现在终端输出与 CI 日志中
func maskPassword(password string) string {
	return strings.Repeat("*", len(password))
}

// dumpAttributes 按 "属性名 = 值" 的格式打印报文属性; User-Password 只打印掩码
func dumpAttributes(p *radius.Packet) {
	if len(p.Attributes) == 0 {
		fmt.Println("    (无属性)")
		return
	}
	for _, avp := range p.Attributes {
		fmt.Printf("    %-18s = %s\n", attributeName(avp.Type), attributeValue(avp))
	}
}

// attributeName 返回属性的可读名称, 未收录的属性退化为类型编号
func attributeName(attrType radius.Type) string {
	switch attrType {
	case rfc2865.UserName_Type:
		return "User-Name"
	case rfc2865.UserPassword_Type:
		return "User-Password"
	case rfc2865.NASIPAddress_Type:
		return "NAS-IP-Address"
	case rfc2865.NASPort_Type:
		return "NAS-Port"
	case rfc2865.ServiceType_Type:
		return "Service-Type"
	case rfc2865.FramedProtocol_Type:
		return "Framed-Protocol"
	case rfc2865.ReplyMessage_Type:
		return "Reply-Message"
	case rfc2865.CalledStationID_Type:
		return "Called-Station-Id"
	case rfc2865.NASIdentifier_Type:
		return "NAS-Identifier"
	case rfc2865.NASPortType_Type:
		return "NAS-Port-Type"
	default:
		return "Type-" + strconv.Itoa(int(attrType))
	}
}

// attributeValue 返回属性的可读值, 未知或非整型属性退化为十六进制
func attributeValue(avp *radius.AVP) string {
	switch avp.Type {
	case rfc2865.UserPassword_Type:
		// 已加密的密码无排错价值, 仅打印掩码
		return fmt.Sprintf("%q", maskPassword(string(avp.Attribute)))
	case rfc2865.UserName_Type, rfc2865.NASIdentifier_Type,
		rfc2865.CalledStationID_Type, rfc2865.ReplyMessage_Type:
		return fmt.Sprintf("%q", string(avp.Attribute))
	case rfc2865.NASIPAddress_Type, rfc2865.FramedIPAddress_Type:
		return net.IP(avp.Attribute).String()
	}

	value, err := radius.Integer(avp.Attribute)
	if err != nil {
		return "0x" + hex.EncodeToString(avp.Attribute)
	}

	switch avp.Type {
	case rfc2865.ServiceType_Type:
		return rfc2865.ServiceType(value).String()
	case rfc2865.FramedProtocol_Type:
		return rfc2865.FramedProtocol(value).String()
	case rfc2865.NASPortType_Type:
		return rfc2865.NASPortType(value).String()
	default:
		return strconv.FormatUint(uint64(value), 10)
	}
}
