package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RuleGroups     []RuleGroup    `json:"rule_groups,omitempty" yaml:"rule_groups,omitempty"`
	RuleMappings   []RuleMapping  `json:"rule_mappings,omitempty" yaml:"rule_mappings,omitempty"`
	VpnAccessRules []VpnAccessRule `json:"vpn_access_rules,omitempty" yaml:"vpn_access_rules,omitempty"`
}

// -------------------- Config Manager --------------------
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse yaml config %q: %w", path, err)
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse json config %q: %w", path, err)
		}
	default:
		// Try json first, then yaml
		if err := json.Unmarshal(data, &cfg); err == nil {
			break
		}
		if err2 := yaml.Unmarshal(data, &cfg); err2 == nil {
			break
		} else {
			return nil, fmt.Errorf("failed to parse config %q as json or yaml: json: %v; yaml: %v", path, err, err2)
		}
	}

	// set default values for rules
	cfg.setRuleDefaults()

	if err := cfg.Check(); err != nil {
		return nil, fmt.Errorf("invalid config %q: %w", path, err)
	}

	return &cfg, nil
}

// Check: check the Config according to the following rules:
// 1) RuleGroups must have unique Name (case-insensitive)
// 2) SrcIpSet names (from mappings) must be unique (case-insensitive)
// 3) RuleMappings Names must be unique within the same MappingType (case-insensitive)
func (cfg Config) Check() error {
	seenGroups := make(map[string]bool)
	for _, g := range cfg.RuleGroups {
		if g.Name == "" {
			continue
		}
		k := strings.ToLower(g.Name)
		if seenGroups[k] {
			return fmt.Errorf("duplicate DestRuleGroup name %q", g.Name)
		}
		seenGroups[k] = true

		for _, r := range g.Rules {
			if !r.Protocol.Valid() {
				return fmt.Errorf("invalid protocol %q [tcp | udp | icmp]", r.Protocol)
			}
			if !r.Action.Valid() {
				return fmt.Errorf("invalid action %q [accept | drop]", r.Action)
			}
		}
	}

	seenSrcIpSets := make(map[string]bool)
	seenMappingByType := make(map[MappingType]map[string]bool)

	for _, mapping := range cfg.RuleMappings {
		if !mapping.Type.Valid() {
			return fmt.Errorf("invalid mapping type: %s [users | public | input_chain | input_chain_ip_set]", mapping.Type)
		}

		// check SrcIpSet name uniqueness when present
		if mapping.SrcIpSet != nil && mapping.SrcIpSet.Name != "" {
			sk := strings.ToLower(mapping.SrcIpSet.Name)
			if seenSrcIpSets[sk] {
				return fmt.Errorf("duplicate SrcIpSet name %q", mapping.SrcIpSet.Name)
			}
			seenSrcIpSets[sk] = true
		}

		if mapping.Name == "" {
			// unnamed mappings: skip name-based validation
			continue
		}
		nameKey := strings.ToLower(mapping.Name)
		if _, ok := seenMappingByType[mapping.Type]; !ok {
			seenMappingByType[mapping.Type] = make(map[string]bool)
		}
		if seenMappingByType[mapping.Type][nameKey] {
			return fmt.Errorf("duplicate mapping name %q for type %s", mapping.Name, mapping.Type)
		}
		seenMappingByType[mapping.Type][nameKey] = true
	}

	for _, r := range cfg.VpnAccessRules {
		if !r.Action.Valid() {
			return fmt.Errorf("invalid VPN action %q [block | logonly]", r.Action)
		}
	}
	return nil
}

// setRuleDefaults sets default and normalized values for rules when fields are omitted
// protocol default: tcp, action default: accept
func (cfg *Config) setRuleDefaults() {
	for gi := range cfg.RuleGroups {
		for ri := range cfg.RuleGroups[gi].Rules {
			r := &cfg.RuleGroups[gi].Rules[ri]
			// default protocol to tcp and normalize to lower-case
			if r.Protocol == "" {
				r.Protocol = ProtocolTcp
			} else {
				r.Protocol = ProtocolType(strings.ToLower(string(r.Protocol)))
			}

			// default action to accept and normalize to lower-case
			if r.Action == "" {
				r.Action = ActionAccept
			} else {
				r.Action = ActionType(strings.ToLower(string(r.Action)))
			}
		}
	}
	for vi := range cfg.VpnAccessRules {
		r := &cfg.VpnAccessRules[vi]
		// default action to drop and normalize to lower-case
		if r.Action == "" {
			r.Action = VpnActionBlock
		} else {
			r.Action = VpnActionType(strings.ToLower(string(r.Action)))
		}
	}
}

// ConfigManager handles configuration file operations with backup support
type ConfigManager struct {
	ConfigPath    string
	BackupDirPath string
	// In-memory cache of current config
	mu     sync.RWMutex
	config *Config
}

// NewConfigManager creates a new config manager
func NewConfigManager(configPath string) *ConfigManager {
	backupDir := filepath.Join(filepath.Dir(configPath), ".config_backups")
	cm := &ConfigManager{
		ConfigPath:    configPath,
		BackupDirPath: backupDir,
	}
	// Load initial config
	if cfg, err := LoadConfig(configPath); err == nil {
		cm.config = cfg
	} else {
		log.Printf("[config] failed to load initial config: %v", err)
		cm.config = &Config{}
	}
	return cm
}

// GetConfig returns a copy of the current in-memory config
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.config == nil {
		return &Config{}
	}
	// Return a deep copy to prevent external modifications
	data, _ := json.Marshal(cm.config)
	var cfg Config
	json.Unmarshal(data, &cfg)
	return &cfg
}

// UpdateConfig updates the in-memory config (without saving to disk yet)
func (cm *ConfigManager) UpdateConfig(config *Config) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = config
}

// EnsureBackupDir creates the backup directory if it doesn't exist
// Uses 0700 permissions to restrict access to the owner only (sensitive config backups)
func (cm *ConfigManager) EnsureBackupDir() error {
	return os.MkdirAll(cm.BackupDirPath, 0700)
}

// BackupConfig creates a timestamped backup of the current config file
// Returns the backup file path
func (cm *ConfigManager) BackupConfig() (string, error) {
	if err := cm.EnsureBackupDir(); err != nil {
		return "", err
	}

	// Read the current config file
	data, err := os.ReadFile(cm.ConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Generate backup filename with timestamp (millisecond precision)
	timestamp := time.Now().Format("20060102_150405.000")
	ext := filepath.Ext(cm.ConfigPath)
	basename := strings.TrimSuffix(filepath.Base(cm.ConfigPath), ext)
	backupPath := filepath.Join(cm.BackupDirPath, fmt.Sprintf("%s_%s%s", basename, timestamp, ext))

	// Write backup file with restricted permissions (0600: owner read/write only)
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	log.Printf("[config] backup created: %s", backupPath)
	return backupPath, nil
}

// SaveConfig saves the configuration to file, creating a backup first
// Empty fields (zeros values) are omitted from the output
func (cm *ConfigManager) SaveConfig(config *Config) error {
	return cm.saveConfigWithBackup(config, true)
}

// SaveConfigWithBackup saves the configuration to file with optional backup
func (cm *ConfigManager) saveConfigWithBackup(config *Config, createBackup bool) error {
	if createBackup {
		// Create backup before modifying
		_, err := cm.BackupConfig()
		if err != nil {
			return fmt.Errorf("failed to backup config: %w", err)
		}
	}

	// Update in-memory cache
	cm.UpdateConfig(config)

	// Marshal the config to the appropriate format
	var data []byte
	var err error
	ext := strings.ToLower(filepath.Ext(cm.ConfigPath))

	switch ext {
	case ".yaml", ".yml":
		// Use YAML marshaling with proper ordering and 2-space indentation
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err = encoder.Encode(config); err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		data = buf.Bytes()
	case ".json":
		// Marshal to JSON, then unmarshal and re-marshal to remove null values
		data, err = json.MarshalIndent(config, "", "  ")
	default:
		// Default to YAML
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err = encoder.Encode(config); err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		data = buf.Bytes()
	}

	// Write to file with restricted permissions (0600: owner read/write only)
	// This prevents other users from reading sensitive configuration data
	if err := os.WriteFile(cm.ConfigPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("[config] config saved to: %s", cm.ConfigPath)
	return nil
}

// SaveVpnAccessConfig saves the VPN access configuration to file, creating a backup first
func (cm *ConfigManager) SaveVpnAccessConfig(rules []VpnAccessRule) error {
	// Read current config to merge with existing rules
	fullConfig, err := LoadConfig(cm.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load existing config: %w", err)
	}

	// Update VPN access rules
	fullConfig.VpnAccessRules = rules

	// Use SaveConfig to save the complete config
	return cm.SaveConfig(fullConfig)
}

// SaveAndApplyConfig atomically saves config and applies rules
// If any step fails, the config is restored from backup
// This ensures consistency between disk config and in-memory rules
func (cm *ConfigManager) SaveAndApplyConfig(config *Config, applyFunc func(*Config) error) error {
	// Step 1: Validate configuration
	if err := config.Check(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Step 2: Create backup of current state
	backupPath, err := cm.BackupConfig()
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Step 3: Save config to disk
	if err := cm.saveConfigWithBackup(config, false); err != nil {
		return fmt.Errorf("save config failed: %w", err)
	}

	// Step 4: Apply rules (nftables, global rules, etc.)
	if err := applyFunc(config); err != nil {
		// ROLLBACK: Restore from backup on failure
		if restoreErr := cm.RestoreBackup(backupPath); restoreErr != nil {
			log.Printf("[config] CRITICAL: restore from backup failed: %v (config may be in inconsistent state)", restoreErr)
			return fmt.Errorf("apply failed and recovery failed: apply err: %w, recovery err: %w", err, restoreErr)
		}
		log.Printf("[config] rollback successful, restored from: %s", backupPath)
		return fmt.Errorf("apply rules failed (rolled back): %w", err)
	}

	// Step 5: Update in-memory cache (already done by saveConfigWithBackup)
	log.Printf("[config] config saved and rules applied successfully")
	return nil
}

// GetBackupList returns a list of backup files
func (cm *ConfigManager) GetBackupList() ([]map[string]string, error) {
	if err := cm.EnsureBackupDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(cm.BackupDirPath)
	if err != nil {
		return nil, err
	}

	var backups []map[string]string
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			backups = append(backups, map[string]string{
				"filename": entry.Name(),
				"path":     filepath.Join(cm.BackupDirPath, entry.Name()),
				"size":     fmt.Sprintf("%d bytes", info.Size()),
				"modified": info.ModTime().Format("2006-01-02 15:04:05.000"),
			})
		}
	}

	return backups, nil
}

// RestoreBackup restores config from a backup file
func (cm *ConfigManager) RestoreBackup(backupPath string) error {
	// Verify backup path is within backup directory (strict security check)
	absBackupPath, err := filepath.Abs(backupPath)
	if err != nil {
		return err
	}

	absBackupDir, err := filepath.Abs(cm.BackupDirPath)
	if err != nil {
		return err
	}

	// Use filepath.Rel() for secure path validation to prevent path traversal
	// This prevents directory traversal attacks like /config escaping via /config-evil prefix
	relPath, err := filepath.Rel(absBackupDir, absBackupPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("invalid backup path: path escapes backup directory")
	}

	// Verify the resolved path is actually within the backup directory
	// (additional validation to catch symbolic link attacks)
	if !strings.HasPrefix(absBackupPath, absBackupDir+string(filepath.Separator)) {
		// Check if it's the backup dir itself (not valid)
		if absBackupPath != absBackupDir {
			return fmt.Errorf("invalid backup path: must be inside backup directory")
		}
	}

	// Read backup file
	data, err := os.ReadFile(absBackupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Create backup before restoring (backup the current state)
	_, err = cm.BackupConfig()
	if err != nil {
		return fmt.Errorf("failed to backup current config: %w", err)
	}

	// Write to config file with restricted permissions (0600: owner read/write only)
	if err := os.WriteFile(cm.ConfigPath, data, 0600); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	log.Printf("[config] config restored from: %s", absBackupPath)
	return nil
}
