package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ocserv-users/internal"
	"time"

	"gopkg.in/yaml.v3"
)

func configWebHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "config.html", nil)
}

// ConfigEditorWebHandler serves the configuration editor page
func configEditorWebHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "config-editor.html", nil)
}

// Global config manager (will be initialized in main.go)
var GlobalConfigManager *internal.ConfigManager

// Response represents a JSON API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ConfigPreviewRequest is used for config preview/validation
type ConfigPreviewRequest struct {
	Content string `json:"content"`
	Format  string `json:"format"` // "json" or "yaml"
}

// ConfigChangesSummary shows what changed in a config
type ConfigChangesSummary struct {
	Added    int      `json:"added"`
	Modified int      `json:"modified"`
	Deleted  int      `json:"deleted"`
	Changes  []string `json:"changes"`
}

// BackupInfo contains metadata about a backup
type BackupInfo struct {
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"mod_time"`
	Description string    `json:"description,omitempty"`
}

// ConfigDiff shows the difference between two versions
type ConfigDiff struct {
	OldFile string `json:"old_file"`
	NewFile string `json:"new_file"`
	Diff    string `json:"diff"`
}

// ==================== Config Export/Import APIs ====================
// ExportConfigHandler exports the entire config as JSON
func ExportConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Load config from cache
	config := GlobalConfigManager.GetConfig()

	type ExportData struct {
		RuleGroups     []internal.RuleGroup     `json:"rule_groups"`
		RuleMappings   []internal.RuleMapping   `json:"rule_mappings"`
		VpnAccessRules []internal.VpnAccessRule `json:"vpn_access_rules,omitempty"`
	}

	vpnRules := internal.GetVpnAccessRules()
	if vpnRules == nil {
		vpnRules = []internal.VpnAccessRule{}
	}

	exportData := ExportData{
		RuleGroups:     config.RuleGroups,
		RuleMappings:   config.RuleMappings,
		VpnAccessRules: vpnRules,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=config.json")
	json.NewEncoder(w).Encode(exportData)
}

// ==================== Unified Config View API ====================
// ConfigViewHandler returns complete config with relationships
func ConfigViewHandler(w http.ResponseWriter, r *http.Request) {
	cfg := GlobalConfigManager.GetConfig()

	sendSuccess(w, "unified config view", cfg)
}

// ConfigValidateHandler validates a configuration
func ConfigValidateHandler(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to 10MB to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	var req ConfigPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	// Parse and validate the configuration
	var cfg internal.Config

	if req.Format == "yaml" {
		if err := yaml.Unmarshal([]byte(req.Content), &cfg); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("YAML parse error: %v", err))
			return
		}
	} else {
		if err := json.Unmarshal([]byte(req.Content), &cfg); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("JSON parse error: %v", err))
			return
		}
	}

	// Run validation
	if err := cfg.Check(); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("validation error: %v", err))
		return
	}

	sendSuccess(w, "validation passed", nil)
}

// ConfigPreviewHandler shows what would change in a configuration
func ConfigPreviewHandler(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to 10MB to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	var req ConfigPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	// Parse new configuration
	var newConfig internal.Config

	if req.Format == "yaml" {
		if err := yaml.Unmarshal([]byte(req.Content), &newConfig); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("YAML parse error: %v", err))
			return
		}
	} else {
		if err := json.Unmarshal([]byte(req.Content), &newConfig); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("JSON parse error: %v", err))
			return
		}
	}

	// Calculate changes
	oldConfig := GlobalConfigManager.GetConfig()
	summary := calculateConfigChanges(oldConfig, &newConfig)

	sendSuccess(w, "preview generated", summary)
}

// ConfigSaveHandler saves a new configuration
func ConfigSaveHandler(w http.ResponseWriter, r *http.Request) {

	// Limit request body size to 10MB to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	var req ConfigPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	// Parse configuration
	var cfg internal.Config

	if req.Format == "yaml" {
		if err := yaml.Unmarshal([]byte(req.Content), &cfg); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("YAML parse error: %v", err))
			return
		}
	} else {
		if err := json.Unmarshal([]byte(req.Content), &cfg); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Sprintf("JSON parse error: %v", err))
			return
		}
	}

	// Define atomic apply function
	applyFunc := func(applyCfg *internal.Config) error {
		// Prepare rules
		publicRules := internal.GetPublicRules(applyCfg)
		inputChainRules := internal.GetInputChainRules(applyCfg)
		srcIpSets, inputChainIpSetRules := internal.GetInputChainIpSetRules(applyCfg)

		// Reinitialize nftables
		if err := internal.InitNftables(publicRules, inputChainRules, inputChainIpSetRules, srcIpSets); err != nil {
			return fmt.Errorf("nftables init failed: %w", err)
		}

		// Update global rules (after nftables succeeds)
		internal.UpdateUserRules(internal.GetUserRulesMapping(applyCfg))
		internal.UpdateVpnAccessRules(applyCfg.VpnAccessRules)

		return nil
	}

	// Save config and apply rules atomically
	if err := GlobalConfigManager.SaveAndApplyConfig(&cfg, applyFunc); err != nil {
		log.Printf("[config] save and apply failed: %v", err)
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendSuccess(w, "config saved successfully", nil)
}

// ruleEqual compares two Rule objects for equality
func ruleEqual(a, b internal.Rule) bool {
	return a.DestIp == b.DestIp &&
		a.DestPort == b.DestPort &&
		a.Protocol == b.Protocol &&
		a.ToLocal == b.ToLocal &&
		a.Action == b.Action &&
		a.SrcIp == b.SrcIp &&
		a.SrcIpSetName == b.SrcIpSetName
}

// ruleGroupEqual compares two RuleGroup objects for equality
func ruleGroupEqual(a, b internal.RuleGroup) bool {
	if a.Name != b.Name || len(a.Rules) != len(b.Rules) {
		return false
	}

	// Compare each rule
	for i := range a.Rules {
		if !ruleEqual(a.Rules[i], b.Rules[i]) {
			return false
		}
	}

	return true
}

// ruleMappingEqual compares two RuleMapping objects for equality
func ruleMappingEqual(a, b internal.RuleMapping) bool {
	if a.Name != b.Name || a.Type != b.Type || a.RuleGroupRef != b.RuleGroupRef {
		return false
	}

	// Compare SrcIps slices
	if len(a.SrcIps) != len(b.SrcIps) {
		return false
	}
	srcIpsMap := make(map[string]bool)
	for _, ip := range a.SrcIps {
		srcIpsMap[ip] = true
	}
	for _, ip := range b.SrcIps {
		if !srcIpsMap[ip] {
			return false
		}
	}

	// Compare Users slices
	if len(a.Users) != len(b.Users) {
		return false
	}
	usersMap := make(map[string]bool)
	for _, user := range a.Users {
		usersMap[user] = true
	}
	for _, user := range b.Users {
		if !usersMap[user] {
			return false
		}
	}

	// Compare SrcIpSet
	if (a.SrcIpSet == nil) != (b.SrcIpSet == nil) {
		return false
	}
	if a.SrcIpSet != nil && b.SrcIpSet != nil {
		if a.SrcIpSet.Name != b.SrcIpSet.Name || len(a.SrcIpSet.Ips) != len(b.SrcIpSet.Ips) {
			return false
		}
		ipsMap := make(map[string]bool)
		for _, ip := range a.SrcIpSet.Ips {
			ipsMap[ip] = true
		}
		for _, ip := range b.SrcIpSet.Ips {
			if !ipsMap[ip] {
				return false
			}
		}
	}

	return true
}

// vpnAccessRuleEqual compares two VpnAccessRule objects for equality
func vpnAccessRuleEqual(a, b internal.VpnAccessRule) bool {
	// Compare Users slices
	if len(a.Users) != len(b.Users) {
		return false
	}
	usersMap := make(map[string]bool)
	for _, user := range a.Users {
		usersMap[user] = true
	}
	for _, user := range b.Users {
		if !usersMap[user] {
			return false
		}
	}

	// Compare RemoteIPs slices
	if len(a.RemoteIPs) != len(b.RemoteIPs) {
		return false
	}
	ipMap := make(map[string]bool)
	for _, ip := range a.RemoteIPs {
		ipMap[ip] = true
	}
	for _, ip := range b.RemoteIPs {
		if !ipMap[ip] {
			return false
		}
	}

	// Compare TimeRange
	if (a.TimeRange == nil) != (b.TimeRange == nil) {
		return false
	}
	if a.TimeRange != nil && b.TimeRange != nil {
		// Compare Start
		if (a.TimeRange.Start == nil) != (b.TimeRange.Start == nil) {
			return false
		}
		if a.TimeRange.Start != nil && b.TimeRange.Start != nil {
			if a.TimeRange.Start.Hour != b.TimeRange.Start.Hour || a.TimeRange.Start.Minute != b.TimeRange.Start.Minute {
				return false
			}
		}
		// Compare End
		if (a.TimeRange.End == nil) != (b.TimeRange.End == nil) {
			return false
		}
		if a.TimeRange.End != nil && b.TimeRange.End != nil {
			if a.TimeRange.End.Hour != b.TimeRange.End.Hour || a.TimeRange.End.Minute != b.TimeRange.End.Minute {
				return false
			}
		}
	}

	return true
}

// getVpnAccessRuleKey generates a unique key for a VPN access rule based on its users
func getVpnAccessRuleKey(rule internal.VpnAccessRule) string {
	users := make([]string, len(rule.Users))
	copy(users, rule.Users)
	return fmt.Sprintf("%v", users)
}

// calculateConfigChanges compares two configurations and returns a summary
func calculateConfigChanges(oldCfg, newCfg *internal.Config) ConfigChangesSummary {
	summary := ConfigChangesSummary{
		Changes: make([]string, 0),
	}

	// Rule groups changes
	oldRuleGroupMap := make(map[string]internal.RuleGroup)
	for _, rg := range oldCfg.RuleGroups {
		oldRuleGroupMap[rg.Name] = rg
	}

	for _, rg := range newCfg.RuleGroups {
		if old, exists := oldRuleGroupMap[rg.Name]; !exists {
			summary.Added++
			summary.Changes = append(summary.Changes, fmt.Sprintf("Added rule group: %s (%d rules)", rg.Name, len(rg.Rules)))
		} else if !ruleGroupEqual(old, rg) {
			summary.Modified++
			summary.Changes = append(summary.Changes, fmt.Sprintf("Modified rule group: %s (%d → %d rules)", rg.Name, len(old.Rules), len(rg.Rules)))
		}
		delete(oldRuleGroupMap, rg.Name)
	}

	for name := range oldRuleGroupMap {
		summary.Deleted++
		summary.Changes = append(summary.Changes, fmt.Sprintf("Deleted rule group: %s", name))
	}

	// Rule mappings changes
	oldMappingMap := make(map[string]internal.RuleMapping)
	for _, m := range oldCfg.RuleMappings {
		oldMappingMap[m.Name] = m
	}

	for _, m := range newCfg.RuleMappings {
		if old, exists := oldMappingMap[m.Name]; !exists {
			summary.Added++
			summary.Changes = append(summary.Changes, fmt.Sprintf("Added mapping: %s (%s)", m.Name, m.Type))
		} else if !ruleMappingEqual(old, m) {
			summary.Modified++
			summary.Changes = append(summary.Changes, fmt.Sprintf("Modified mapping: %s", m.Name))
		}
		delete(oldMappingMap, m.Name)
	}

	for name := range oldMappingMap {
		summary.Deleted++
		summary.Changes = append(summary.Changes, fmt.Sprintf("Deleted mapping: %s", name))
	}

	// VPN access rules changes
	oldVpnRulesMap := make(map[string]internal.VpnAccessRule)
	for _, rule := range oldCfg.VpnAccessRules {
		key := getVpnAccessRuleKey(rule)
		oldVpnRulesMap[key] = rule
	}

	for _, rule := range newCfg.VpnAccessRules {
		key := getVpnAccessRuleKey(rule)
		if old, exists := oldVpnRulesMap[key]; !exists {
			summary.Added++
			userList := fmt.Sprintf("%v", rule.Users)
			summary.Changes = append(summary.Changes, fmt.Sprintf("Added VPN access rule for users: %s", userList))
		} else if !vpnAccessRuleEqual(old, rule) {
			summary.Modified++
			userList := fmt.Sprintf("%v", rule.Users)
			summary.Changes = append(summary.Changes, fmt.Sprintf("Modified VPN access rule for users: %s", userList))
		}
		delete(oldVpnRulesMap, key)
	}

	for key := range oldVpnRulesMap {
		summary.Deleted++
		rule := oldVpnRulesMap[key]
		userList := fmt.Sprintf("%v", rule.Users)
		summary.Changes = append(summary.Changes, fmt.Sprintf("Deleted VPN access rule for users: %s", userList))
	}

	return summary
}
