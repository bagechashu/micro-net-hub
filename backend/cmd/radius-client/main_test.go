package main

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	testSecret   = "test-radius-secret"
	testUser     = "admin"
	testPass     = "admin_pass"
	testOTP      = "000000"
	testPassword = testPass + testOTP
)

// startFakeServer 启动一个仅用于测试的 RADIUS 服务端, 返回其监听地址.
//
// 复用 layeh.com/radius 的 PacketServer + 随机端口, 使端到端测试不依赖外部服务.
func startFakeServer(t *testing.T, handler radius.Handler) string {
	t.Helper()

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &radius.PacketServer{
		SecretSource: radius.StaticSecretSource([]byte(testSecret)),
		Handler:      handler,
		ErrorLog:     log.New(io.Discard, "", 0),
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = server.Serve(conn)
	}()

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Log("假 RADIUS 服务端未在超时内退出")
		}
	})

	return conn.LocalAddr().String()
}

// checkAuthHandler 仅当用户名密码与测试常量一致时返回 Access-Accept
func checkAuthHandler(t *testing.T) radius.Handler {
	t.Helper()

	return radius.HandlerFunc(func(w radius.ResponseWriter, r *radius.Request) {
		code := radius.CodeAccessReject
		if rfc2865.UserName_GetString(r.Packet) == testUser &&
			rfc2865.UserPassword_GetString(r.Packet) == testPassword {
			code = radius.CodeAccessAccept
		}
		assert.NoError(t, w.Write(r.Response(code)))
	})
}

// TestNormalizeServerAddr 校验地址补全与非法地址识别
func TestNormalizeServerAddr(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		want    string
		wantErr bool
	}{
		{name: "host only", addr: "127.0.0.1", want: "127.0.0.1:1812"},
		{name: "host with port", addr: "127.0.0.1:11812", want: "127.0.0.1:11812"},
		{name: "domain name", addr: "radius.example.com", want: "radius.example.com:1812"},
		{name: "ipv6", addr: "::1", want: "[::1]:1812"},
		{name: "trim spaces", addr: " 127.0.0.1 ", want: "127.0.0.1:1812"},
		{name: "empty", addr: "", wantErr: true},
		{name: "too many colons", addr: "127.0.0.1:1812:extra", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeServerAddr(tt.addr)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestOptionsPassword 密码必须是 "数据库密码 + TOTP" 的直接拼接
func TestOptionsPassword(t *testing.T) {
	assert.Equal(t, testPassword, options{pass: testPass, otp: testOTP}.password())
	assert.Equal(t, testPass, options{pass: testPass}.password())
}

// TestOptionsValidate 校验参数检查规则
func TestOptionsValidate(t *testing.T) {
	valid := options{user: testUser, pass: testPass, otp: testOTP, count: 1, timeout: time.Second}

	tests := []struct {
		name    string
		modify  func(*options)
		wantErr bool
	}{
		{name: "valid", modify: func(*options) {}},
		{name: "valid without otp", modify: func(o *options) { o.otp = "" }},
		{name: "missing user", modify: func(o *options) { o.user = "  " }, wantErr: true},
		{name: "missing pass", modify: func(o *options) { o.pass = "" }, wantErr: true},
		{name: "otp not digit", modify: func(o *options) { o.otp = "00000a" }, wantErr: true},
		{name: "otp wrong length", modify: func(o *options) { o.otp = "12345" }, wantErr: true},
		{name: "count zero", modify: func(o *options) { o.count = 0 }, wantErr: true},
		{name: "timeout zero", modify: func(o *options) { o.timeout = 0 }, wantErr: true},
		{name: "negative interval", modify: func(o *options) { o.interval = -time.Second }, wantErr: true},
		{name: "negative retry", modify: func(o *options) { o.retry = -time.Second }, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := valid
			tt.modify(&opts)

			err := opts.validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestBuildAccessRequest 校验报文的属性组装结果
func TestBuildAccessRequest(t *testing.T) {
	opts := options{
		secret:  testSecret,
		user:    testUser,
		pass:    testPass,
		otp:     testOTP,
		nasID:   "ocserv-1",
		nasIP:   "10.0.0.1",
		nasPort: 7,
	}

	packet, err := buildAccessRequest(opts, "127.0.0.1:1812")
	require.NoError(t, err)

	assert.Equal(t, radius.CodeAccessRequest, packet.Code)
	assert.Equal(t, []byte(testSecret), packet.Secret)
	assert.Equal(t, testUser, rfc2865.UserName_GetString(packet))
	// UserPassword_GetString 会用报文自身的 Authenticator 解密, 因此这里应还原出明文
	assert.Equal(t, testPassword, rfc2865.UserPassword_GetString(packet))
	assert.Equal(t, "ocserv-1", rfc2865.NASIdentifier_GetString(packet))
	assert.Equal(t, "10.0.0.1", rfc2865.NASIPAddress_Get(packet).String())
	assert.Equal(t, rfc2865.NASPort(7), rfc2865.NASPort_Get(packet))
	// 固定携带的 NAS 属性, 与 ocserv 真实请求保持一致
	assert.Equal(t, rfc2865.NASPortType_Value_Virtual, rfc2865.NASPortType_Get(packet))
	assert.Equal(t, rfc2865.ServiceType_Value_LoginUser, rfc2865.ServiceType_Get(packet))
	assert.Equal(t, rfc2865.FramedProtocol_Value_PPP, rfc2865.FramedProtocol_Get(packet))
}

// TestBuildAccessRequestOptionalAttrs 可选属性留空时不应出现在报文中
func TestBuildAccessRequestOptionalAttrs(t *testing.T) {
	opts := options{secret: testSecret, user: testUser, pass: testPass, otp: testOTP, nasIP: "127.0.0.1"}

	packet, err := buildAccessRequest(opts, "127.0.0.1:1812")
	require.NoError(t, err)

	value, err := rfc2865.NASIdentifier_Lookup(packet)
	assert.Equal(t, radius.ErrNoAttribute, err)
	assert.Empty(t, value)

	_, hasNasPort := packet.Lookup(rfc2865.NASPort_Type)
	assert.False(t, hasNasPort)
}

// TestBuildAccessRequestInvalidNasIP -nas-ip 非法时必须显式报错, 不能静默忽略
func TestBuildAccessRequestInvalidNasIP(t *testing.T) {
	opts := options{secret: testSecret, user: testUser, pass: testPass, nasIP: "not-an-ip"}

	packet, err := buildAccessRequest(opts, "127.0.0.1:1812")
	require.Error(t, err)
	assert.Nil(t, packet)
	assert.Contains(t, err.Error(), "-nas-ip")
}

// TestExchangeAccept 端到端验证请求被假服务端接受
func TestExchangeAccept(t *testing.T) {
	addr := startFakeServer(t, checkAuthHandler(t))

	opts := options{secret: testSecret, user: testUser, pass: testPass, otp: testOTP, timeout: 3 * time.Second}
	result := exchange(context.Background(), opts, addr)

	require.NoError(t, result.err)
	assert.True(t, result.accepted())
	assert.Equal(t, exitAccept, exitCode([]attemptResult{result}))
}

// TestExchangeReject 密码错误时收到 Access-Reject
func TestExchangeReject(t *testing.T) {
	addr := startFakeServer(t, checkAuthHandler(t))

	opts := options{secret: testSecret, user: testUser, pass: "wrong_pass", otp: testOTP, timeout: 3 * time.Second}
	result := exchange(context.Background(), opts, addr)

	require.NoError(t, result.err)
	assert.Equal(t, radius.CodeAccessReject, result.code)
	assert.False(t, result.accepted())
	assert.Equal(t, exitReject, exitCode([]attemptResult{result}))
}

// TestExchangeTimeout 服务端不响应时应返回超时错误, 并归类为传输层失败
func TestExchangeTimeout(t *testing.T) {
	addr := startFakeServer(t, radius.HandlerFunc(func(radius.ResponseWriter, *radius.Request) {}))

	opts := options{secret: testSecret, user: testUser, pass: testPass, otp: testOTP, timeout: 300 * time.Millisecond}
	result := exchange(context.Background(), opts, addr)

	require.Error(t, result.err)
	assert.ErrorIs(t, result.err, context.DeadlineExceeded)
	assert.Equal(t, exitTransport, exitCode([]attemptResult{result}))
}

// TestExitCode 汇总规则: 传输层错误优先, 其次要求全部 Accept
func TestExitCode(t *testing.T) {
	accepted := attemptResult{code: radius.CodeAccessAccept}
	rejected := attemptResult{code: radius.CodeAccessReject}
	failed := attemptResult{err: context.DeadlineExceeded}

	tests := []struct {
		name    string
		results []attemptResult
		want    int
	}{
		{name: "empty", results: nil, want: exitReject},
		{name: "all accepted", results: []attemptResult{accepted, accepted}, want: exitAccept},
		{name: "rejected", results: []attemptResult{accepted, rejected}, want: exitReject},
		{name: "transport error wins", results: []attemptResult{accepted, failed}, want: exitTransport},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, exitCode(tt.results))
		})
	}
}

// TestAttributeDisplay 校验属性名与属性值的展示(密码必须脱敏)
func TestAttributeDisplay(t *testing.T) {
	assert.Equal(t, "User-Name", attributeName(rfc2865.UserName_Type))
	assert.Equal(t, "NAS-Identifier", attributeName(rfc2865.NASIdentifier_Type))
	assert.Equal(t, "Type-200", attributeName(radius.Type(200)))

	assert.Equal(t, `"**********"`, attributeValue(&radius.AVP{
		Type:      rfc2865.UserPassword_Type,
		Attribute: radius.Attribute("1234567890"),
	}))
	assert.Equal(t, `"admin"`, attributeValue(&radius.AVP{
		Type:      rfc2865.UserName_Type,
		Attribute: radius.Attribute("admin"),
	}))
	assert.Equal(t, "10.0.0.1", attributeValue(&radius.AVP{
		Type:      rfc2865.NASIPAddress_Type,
		Attribute: radius.Attribute(net.ParseIP("10.0.0.1").To4()),
	}))
	assert.Equal(t, "Login-User", attributeValue(&radius.AVP{
		Type:      rfc2865.ServiceType_Type,
		Attribute: radius.NewInteger(uint32(rfc2865.ServiceType_Value_LoginUser)),
	}))
	assert.Equal(t, "7", attributeValue(&radius.AVP{
		Type:      rfc2865.NASPort_Type,
		Attribute: radius.NewInteger(7),
	}))
	assert.Equal(t, "0x0102", attributeValue(&radius.AVP{
		Type:      radius.Type(200),
		Attribute: radius.Attribute{0x01, 0x02},
	}))
}

// TestMaskPassword 脱敏只保留长度信息
func TestMaskPassword(t *testing.T) {
	assert.Equal(t, "", maskPassword(""))
	assert.Equal(t, "*****", maskPassword("12345"))
}
