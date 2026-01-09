package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sigs.k8s.io/yaml"
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

		for ri, r := range g.Rules {
			if !r.Protocol.Valid() {
				return fmt.Errorf("invalid protocol %q in rule group %q rule index %d", r.Protocol, g.Name, ri)
			}
			if !r.Action.Valid() {
				return fmt.Errorf("invalid action %q in rule group %q rule index %d", r.Action, g.Name, ri)
			}
		}
	}

	seenSrcIpSets := make(map[string]bool)
	seenMappingByType := make(map[MappingType]map[string]bool)

	for _, mapping := range cfg.RuleMappings {
		if !mapping.Type.Valid() {
			return fmt.Errorf("无效的规则映射类型: %s", mapping.Type)
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
}

// ConfigManager handles configuration file operations with backup support
type ConfigManager struct {
	ConfigPath    string
	BackupDirPath string
}

// NewConfigManager creates a new config manager
func NewConfigManager(configPath string) *ConfigManager {
	backupDir := filepath.Join(filepath.Dir(configPath), ".config_backups")
	return &ConfigManager{
		ConfigPath:    configPath,
		BackupDirPath: backupDir,
	}
}

// EnsureBackupDir creates the backup directory if it doesn't exist
func (cm *ConfigManager) EnsureBackupDir() error {
	return os.MkdirAll(cm.BackupDirPath, 0755)
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

	// Write backup file
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	log.Printf("[config] backup created: %s", backupPath)
	return backupPath, nil
}

// SaveConfig saves the configuration to file, creating a backup first
// Empty fields (zeros values) are omitted from the output
func (cm *ConfigManager) SaveConfig(config *Config) error {
	// Create backup before modifying
	_, err := cm.BackupConfig()
	if err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	// Marshal the config to the appropriate format
	var data []byte
	ext := strings.ToLower(filepath.Ext(cm.ConfigPath))

	switch ext {
	case ".yaml", ".yml":
		// Use YAML marshaling with proper ordering
		data, err = yaml.Marshal(config)
	case ".json":
		// Marshal to JSON, then unmarshal and re-marshal to remove null values
		data, err = json.MarshalIndent(config, "", "  ")
	default:
		// Default to YAML
		data, err = yaml.Marshal(config)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cm.ConfigPath, data, 0644); err != nil {
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
	// Verify backup path is within backup directory (security check)
	absBackupPath, err := filepath.Abs(backupPath)
	if err != nil {
		return err
	}

	absBackupDir, err := filepath.Abs(cm.BackupDirPath)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(absBackupPath, absBackupDir) {
		return fmt.Errorf("invalid backup path")
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

	// Write to config file
	if err := os.WriteFile(cm.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	log.Printf("[config] config restored from: %s", absBackupPath)
	return nil
}
