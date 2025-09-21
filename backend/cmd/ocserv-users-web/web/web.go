package web

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"ocserv-users/internal"
)

// -------------------- Web handler --------------------

//go:embed index.html
var Static embed.FS

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(Static, "static/index.html"))
	sessions, err := internal.GetSessions()
	if err != nil {
		http.Error(w, "无法获取用户数据: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = tmpl.Execute(w, sessions)
}

func NftablesHandler(w http.ResponseWriter, r *http.Request) {
	// 只接受 POST 请求
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	// 执行一次 nftables 规则更新
	err := internal.UpdateNftablesRulesWithSessions(internal.GlobalUserRules)
	if err != nil {
		http.Error(w, fmt.Sprintf("更新规则失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	// w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(map[string]string{
	// 	"status":  "success",
	// 	"message": "nftables 规则更新完成",
	// })
}
func RunWebServer(addr string) {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/nftables", NftablesHandler)
	go func() {
		log.Printf("[web] 服务运行中: http://%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()
}
