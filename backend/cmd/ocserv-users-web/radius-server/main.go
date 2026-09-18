// radius-server.go
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// ConfigUser 代表 JSON 中的用户定义
type ConfigUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var userMap map[string]string

func main() {
	// 命令行参数
	var (
		listenAddr = flag.String("addr", "0.0.0.0:1812", "RADIUS 服务监听地址")
		secret     = flag.String("secret", "default-radius-secret", "RADIUS 共享密钥")
		usersFile  = flag.String("users", "users.json", "用户列表 JSON 文件")
	)
	flag.Parse()

	// 加载用户
	data, err := os.ReadFile(*usersFile)
	if err != nil {
		log.Fatalf("无法读取用户文件 %s: %v", *usersFile, err)
	}

	var configUsers []ConfigUser
	if err := json.Unmarshal(data, &configUsers); err != nil {
		log.Fatalf("无法解析用户文件: %v", err)
	}

	// 建立用户查找表
	userMap = make(map[string]string)
	for _, u := range configUsers {
		userMap[u.Username] = u.Password
	}
	log.Printf("加载 %d 个用户", len(userMap))

	// 使用 PacketServer 提供服务
	server := radius.PacketServer{
		Addr:               *listenAddr,
		SecretSource:       radius.StaticSecretSource([]byte(*secret)),
		Handler:            radius.HandlerFunc(AuthHandler),
		InsecureSkipVerify: false,
	}

	log.Printf("RADIUS 服务器启动在 %s", *listenAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("RADIUS 服务器启动失败: %v", err)
	}
}

// AuthHandler handles RADIUS Access-Request packets and returns Accept/Reject.
func AuthHandler(w radius.ResponseWriter, r *radius.Request) {
	// r.Packet 是 *radius.Packet
	// 使用 rfc2865 取得用户名和密码
	username := rfc2865.UserName_GetString(r.Packet)
	password := rfc2865.UserPassword_GetString(r.Packet)

	log.Printf("收到 Access-Request: user=%s from=%s", username, r.RemoteAddr)

	code := radius.CodeAccessReject
	// 检查用户名/密码是否存在且匹配
	if pw, exists := userMap[username]; exists && pw == password {
		code = radius.CodeAccessAccept
	} else {
		log.Printf("认证拒绝: user=%s", username)
	}
	// 写入响应
	// r.Response(code) 会构造一个符合 RFC rfc2865 的响应
	if err := w.Write(r.Response(code)); err != nil {
		log.Printf("RADIUS write response error: %v", err)
	}
}
