package web

import (
	"fmt"
	"net/http"
	"ocserv-users/internal"
)

func nftCheckTriggerHandler(w http.ResponseWriter, r *http.Request) {
	// Get a thread-safe copy of user rules
	userRules := internal.GetUserRules()
	
	// 执行一次 nftables 规则更新
	err := internal.UpdateUsersNftablesRules(userRules)
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

func vpnAccessCheckTriggerHandler(w http.ResponseWriter, r *http.Request) {
	// Get a thread-safe copy of VPN access rules
	vpnRules := internal.GetVpnAccessRules()
	
	// 执行一次 EnforceVpnAccess 规则检查
	err := internal.EnforceVpnAccess(vpnRules)
	if err != nil {
		http.Error(w, fmt.Sprintf("vpn access enforcement failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("vpn access enforced\n"))
}
