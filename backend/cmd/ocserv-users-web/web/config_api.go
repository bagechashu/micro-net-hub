package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ocserv-users/internal"
	"path/filepath"
	"strings"
)

// Global config manager (will be initialized in main.go)
var GlobalConfigManager *internal.ConfigManager

// Response represents a JSON API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Helper functions
func sendJSON(w http.ResponseWriter, code int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, code int, message string) {
	sendJSON(w, code, Response{Code: code, Message: message})
}

func sendSuccess(w http.ResponseWriter, message string, data interface{}) {
	sendJSON(w, http.StatusOK, Response{Code: 0, Message: message, Data: data})
}

// ==================== Rule Groups APIs ====================

// GetRuleGroups returns all rule groups
func GetRuleGroupsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Load config to get rule groups
	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	config, err := internal.LoadConfig(cfgPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	sendSuccess(w, "rule groups retrieved", config.RuleGroups)
}

// CreateRuleGroup creates a new rule group
func CreateRuleGroupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	// Load current config
	config, err := internal.LoadConfig(cfgPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	// Parse request body
	var ruleGroup internal.RuleGroup
	if err := json.NewDecoder(r.Body).Decode(&ruleGroup); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Check for duplicate names
	for _, g := range config.RuleGroups {
		if strings.ToLower(g.Name) == strings.ToLower(ruleGroup.Name) {
			sendError(w, http.StatusConflict, "rule group with same name already exists")
			return
		}
	}

	// Add new rule group
	config.RuleGroups = append(config.RuleGroups, ruleGroup)

	// Save config
	if err := GlobalConfigManager.SaveConfig(config); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save config: %v", err))
		return
	}

	sendSuccess(w, "rule group created", ruleGroup)
}

// UpdateRuleGroup updates an existing rule group
func UpdateRuleGroupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	// Load current config
	config, err := internal.LoadConfig(cfgPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	// Parse request body
	var updatedGroup internal.RuleGroup
	if err := json.NewDecoder(r.Body).Decode(&updatedGroup); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Find and update rule group
	found := false
	for i, g := range config.RuleGroups {
		if strings.ToLower(g.Name) == strings.ToLower(updatedGroup.Name) {
			config.RuleGroups[i] = updatedGroup
			found = true
			break
		}
	}

	if !found {
		sendError(w, http.StatusNotFound, "rule group not found")
		return
	}

	// Save config
	if err := GlobalConfigManager.SaveConfig(config); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save config: %v", err))
		return
	}

	sendSuccess(w, "rule group updated", updatedGroup)
}

// DeleteRuleGroup deletes a rule group
func DeleteRuleGroupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		sendError(w, http.StatusBadRequest, "name parameter is required")
		return
	}

	// Load current config
	config, err := internal.LoadConfig(cfgPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	// Find and remove rule group
	found := false
	for i, g := range config.RuleGroups {
		if strings.ToLower(g.Name) == strings.ToLower(name) {
			config.RuleGroups = append(config.RuleGroups[:i], config.RuleGroups[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		sendError(w, http.StatusNotFound, "rule group not found")
		return
	}

	// Save config
	if err := GlobalConfigManager.SaveConfig(config); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save config: %v", err))
		return
	}

	sendSuccess(w, "rule group deleted", nil)
}

// ==================== Rule Mappings APIs ====================

// GetRuleMappings returns all rule mappings
func GetRuleMappingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	config, err := internal.LoadConfig(cfgPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	sendSuccess(w, "rule mappings retrieved", config.RuleMappings)
}

// CreateRuleMapping creates a new rule mapping
func CreateRuleMappingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	// Load current config
	config, err := internal.LoadConfig(cfgPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	// Parse request body
	var ruleMapping internal.RuleMapping
	if err := json.NewDecoder(r.Body).Decode(&ruleMapping); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Add new rule mapping
	config.RuleMappings = append(config.RuleMappings, ruleMapping)

	// Save config
	if err := GlobalConfigManager.SaveConfig(config); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save config: %v", err))
		return
	}

	sendSuccess(w, "rule mapping created", ruleMapping)
}

// UpdateRuleMapping updates an existing rule mapping
func UpdateRuleMappingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	// Load current config
	config, err := internal.LoadConfig(cfgPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	// Parse request body
	var updatedMapping internal.RuleMapping
	if err := json.NewDecoder(r.Body).Decode(&updatedMapping); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Find and update rule mapping
	found := false
	for i, m := range config.RuleMappings {
		if strings.ToLower(m.Name) == strings.ToLower(updatedMapping.Name) {
			config.RuleMappings[i] = updatedMapping
			found = true
			break
		}
	}

	if !found {
		sendError(w, http.StatusNotFound, "rule mapping not found")
		return
	}

	// Save config
	if err := GlobalConfigManager.SaveConfig(config); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save config: %v", err))
		return
	}

	sendSuccess(w, "rule mapping updated", updatedMapping)
}

// DeleteRuleMapping deletes a rule mapping
func DeleteRuleMappingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		sendError(w, http.StatusBadRequest, "name parameter is required")
		return
	}

	// Load current config
	config, err := internal.LoadConfig(cfgPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	// Find and remove rule mapping
	found := false
	for i, m := range config.RuleMappings {
		if strings.ToLower(m.Name) == strings.ToLower(name) {
			config.RuleMappings = append(config.RuleMappings[:i], config.RuleMappings[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		sendError(w, http.StatusNotFound, "rule mapping not found")
		return
	}

	// Save config
	if err := GlobalConfigManager.SaveConfig(config); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save config: %v", err))
		return
	}

	sendSuccess(w, "rule mapping deleted", nil)
}

// ==================== VPN Access Rules APIs ====================

// GetVpnAccessRules returns all VPN access rules
func GetVpnAccessRulesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if len(internal.Global_VpnAccessRules) == 0 {
		sendSuccess(w, "vpn access rules retrieved", []interface{}{})
		return
	}

	sendSuccess(w, "vpn access rules retrieved", internal.Global_VpnAccessRules)
}

// UpdateVpnAccessRules updates VPN access rules
func UpdateVpnAccessRulesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	// Parse request body
	var vpnRules []internal.VpnAccessRule
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &vpnRules); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Save config
	if err := GlobalConfigManager.SaveVpnAccessConfig(vpnRules); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save config: %v", err))
		return
	}

	// Update global rules
	internal.Global_VpnAccessRules = vpnRules

	sendSuccess(w, "vpn access rules updated", vpnRules)
}

// ==================== Backup APIs ====================

// GetBackupList returns a list of backups
func GetBackupListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	backups, err := GlobalConfigManager.GetBackupList()
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get backup list: %v", err))
		return
	}

	sendSuccess(w, "backup list retrieved", backups)
}

// RestoreBackupHandler restores from a backup
func RestoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	backupFilename := r.URL.Query().Get("filename")
	if backupFilename == "" {
		sendError(w, http.StatusBadRequest, "filename parameter is required")
		return
	}

	backupPath := filepath.Join(GlobalConfigManager.BackupDirPath, backupFilename)

	if err := GlobalConfigManager.RestoreBackup(backupPath); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to restore backup: %v", err))
		return
	}

	// Reload configs
	var err error
	internal.Global_UsersRules, err = reloadUserRules()
	if err != nil {
		log.Printf("[api] failed to reload user rules: %v", err)
	}

	internal.Global_VpnAccessRules, err = reloadVpnAccessRules()
	if err != nil {
		log.Printf("[api] failed to reload vpn access rules: %v", err)
	}

	sendSuccess(w, "backup restored successfully", nil)
}

// ==================== Config Export/Import APIs ====================

// ExportConfigHandler exports the entire config as JSON
func ExportConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cfgPath := r.Header.Get("X-Config-Path")
	if cfgPath == "" {
		cfgPath = "rules.yaml"
	}

	// Load config to get all data
	config, err := internal.LoadConfig(cfgPath)
	if err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load config: %v", err))
		return
	}

	type ExportData struct {
		RuleGroups     []internal.RuleGroup     `json:"rule_groups"`
		RuleMappings   []internal.RuleMapping   `json:"rule_mappings"`
		VpnAccessRules []internal.VpnAccessRule `json:"vpn_access_rules,omitempty"`
	}

	vpnRules := []internal.VpnAccessRule{}
	if internal.Global_VpnAccessRules != nil {
		vpnRules = internal.Global_VpnAccessRules
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

// Helper function to reload user rules
func reloadUserRules() (map[string][]internal.Rule, error) {
	config, err := internal.LoadConfig("rules.yaml")
	if err != nil {
		return nil, err
	}
	return internal.GetUserRulesMapping(config), nil
}

// Helper function to reload VPN access rules
func reloadVpnAccessRules() ([]internal.VpnAccessRule, error) {
	cfg, err := internal.LoadConfig("rules.yaml")
	if err != nil {
		return nil, err
	}
	return cfg.VpnAccessRules, nil
}
