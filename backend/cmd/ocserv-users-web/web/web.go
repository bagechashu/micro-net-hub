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
	tmpl := template.Must(template.ParseFS(Static, "index.html"))
	sessions, err := internal.GetSessions()
	if err != nil {
		http.Error(w, fmt.Sprintf("无法获取用户数据: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	_ = tmpl.Execute(w, sessions)
}

func nftablesHandler(w http.ResponseWriter, r *http.Request) {
	// 只接受 POST 请求
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 执行一次 nftables 规则更新
	err := internal.UpdateNftablesRulesWithSessions(internal.GlobalUserRules)
	if err != nil {
		http.Error(w, fmt.Sprintf("nft updated failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("nft updated successfully\n"))

	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(map[string]string{
	// 	"status":  "success",
	// 	"message": "nft updated successfully",
	// })
}

func RunWebServer(addr string) {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/nftables", nftablesHandler)
	go func() {
		log.Printf("[web] 服务运行中: http://%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()
}
