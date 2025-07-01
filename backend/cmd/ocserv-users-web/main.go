// env GOOS=linux GOARCH=amd64 go build
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strconv"
)

// Session 表示每个用户 VPN 会话
type Session struct {
	ID           int    `json:"ID"`
	Username     string `json:"Username"`
	Groupname    string `json:"Groupname"`
	State        string `json:"State"`
	RemoteIP     string `json:"Remote IP"`
	UserAgent    string `json:"User-Agent"`
	RX           string `json:"RX"` // raw bytes as string
	TX           string `json:"TX"`
	AverageRX    string `json:"Average RX"`
	AverageTX    string `json:"Average TX"`
	ConnectedAt  string `json:"Connected at"`  // string time
	ConnectedFor string `json:"_Connected at"` // duration string

	RXHuman string
	TXHuman string
}

func toHumanSize(bytesStr string) string {
	n, err := strconv.ParseInt(bytesStr, 10, 64)
	if err != nil {
		return "?"
	}
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case n > GB:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(GB))
	case n > MB:
		return fmt.Sprintf("%.2f MB", float64(n)/float64(MB))
	case n > KB:
		return fmt.Sprintf("%.2f KB", float64(n)/float64(KB))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func getSessions() ([]Session, error) {
	cmd := exec.Command("occtl", "-j", "show", "users")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var sessions []Session
	if err := json.Unmarshal(output, &sessions); err != nil {
		return nil, err
	}

	for i := range sessions {
		sessions[i].RXHuman = toHumanSize(sessions[i].RX)
		sessions[i].TXHuman = toHumanSize(sessions[i].TX)
	}

	return sessions, nil
}

//go:embed static/index.html
var Static embed.FS

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(Static, "static/index.html"))

	sessions, err := getSessions()
	if err != nil {
		http.Error(w, "无法获取用户数据: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, sessions)
}

func main() {
	http.HandleFunc("/", indexHandler)

	log.Println("服务运行中： http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("启动失败:", err)
	}
}
