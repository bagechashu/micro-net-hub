package web

import (
	"log"
	"net/http"
)

func RunWebServer(addr string) {
	// corefunc
	http.HandleFunc("/nftables", nftablesHandler)
	http.HandleFunc("/vpnaccess", vpnAccessHandler)

	// static
	registerStatic()

	// index
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/partials/users.html", usersPartialHandler)
	http.HandleFunc("/partials/rule-groups.html", RuleGroupsPartialHandler)
	http.HandleFunc("/partials/rule-mappings.html", RuleMappingsPartialHandler)
	http.HandleFunc("/partials/vpn-access.html", VpnAccessPartialHandler)
	http.HandleFunc("/partials/backups.html", BackupsPartialHandler)

	// Config Management APIs
	http.HandleFunc("/config.html", configHandler)
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
