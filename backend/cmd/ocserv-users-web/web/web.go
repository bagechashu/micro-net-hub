package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"ocserv-users/internal"
)

// -------------------- Web handler --------------------

//go:embed index.html
var Static embed.FS

//go:embed config.html
var ConfigHTML embed.FS

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(Static, "index.html"))
	_ = tmpl.Execute(w, nil)
}

func nftablesHandler(w http.ResponseWriter, r *http.Request) {
	// 只接受 POST 请求
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 执行一次 nftables 规则更新
	err := internal.UpdateNftablesRulesWithSessions(internal.Global_UsersRules)
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

func vpnAccessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 执行一次 EnforceVpnAccess 规则检查
	err := internal.EnforceVpnAccess(internal.Global_VpnAccessRules)
	if err != nil {
		http.Error(w, fmt.Sprintf("vpn access enforcement failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("vpn access enforced\n"))
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(ConfigHTML, "config.html"))
	_ = tmpl.Execute(w, nil)
}

func usersAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessions, err := internal.GetSessions()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("无法获取用户数据: %s", err.Error()),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(sessions)
}

func RunWebServer(addr string) {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/config.html", configHandler)
	http.HandleFunc("/api/users", usersAPIHandler)
	http.HandleFunc("/nftables", nftablesHandler)
	http.HandleFunc("/vpnaccess", vpnAccessHandler)
	
	// Config Management APIs
	http.HandleFunc("/api/rule-groups", GetRuleGroupsHandler)
	http.HandleFunc("/api/rule-groups/create", CreateRuleGroupHandler)
	http.HandleFunc("/api/rule-groups/update", UpdateRuleGroupHandler)
	http.HandleFunc("/api/rule-groups/delete", DeleteRuleGroupHandler)
	
	http.HandleFunc("/api/rule-mappings", GetRuleMappingsHandler)
	http.HandleFunc("/api/rule-mappings/create", CreateRuleMappingHandler)
	http.HandleFunc("/api/rule-mappings/update", UpdateRuleMappingHandler)
	http.HandleFunc("/api/rule-mappings/delete", DeleteRuleMappingHandler)
	
	http.HandleFunc("/api/vpn-access-rules", GetVpnAccessRulesHandler)
	http.HandleFunc("/api/vpn-access-rules/update", UpdateVpnAccessRulesHandler)
	
	// Backup APIs
	http.HandleFunc("/api/backups", GetBackupListHandler)
	http.HandleFunc("/api/backups/restore", RestoreBackupHandler)
	
	// Export APIs
	http.HandleFunc("/api/config/export", ExportConfigHandler)
	
	go func() {
		log.Printf("[web] 服务运行中: http://%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()
}
