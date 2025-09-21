package internal

import (
	"encoding/json"
	"log"
	"os"
)

type RuleConfig struct {
	IP       string `json:"ip"`
	Protocol string `json:"protocol"`
	Port     uint16 `json:"port,omitempty"`
}

type RuleGroup struct {
	Name  string       `json:"name"`
	Rules []RuleConfig `json:"rules"`
}

type UserGroup struct {
	Name    string   `json:"name"`
	Users   []string `json:"users"`
	RuleRef string   `json:"rule_ref"` // 引用的规则组名称
}

type FullConfig struct {
	Rules         []RuleGroup `json:"rules"`
	Users         []UserGroup `json:"users"`
	PublicRuleRef string      `json:"public_rule_ref"` // 引用的规则组名称
}

// -------------------- Config Manager --------------------
func LoadConfig(path string) (*FullConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg FullConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// resolveRuleGroup 根据规则组名称查找规则组
func resolveRuleGroup(groups []RuleGroup, name string) *RuleGroup {
	for _, group := range groups {
		if group.Name == name {
			return &group
		}
	}
	return nil
}

// buildUserRulesMapping 构建用户到规则的映射
func buildUserRulesMapping(config *FullConfig) map[string][]RuleConfig {
	userRules := make(map[string][]RuleConfig)

	// 为每个用户组构建规则映射
	for _, userGroup := range config.Users {
		// 查找该用户组引用的规则组
		ruleGroup := resolveRuleGroup(config.Rules, userGroup.RuleRef)
		if ruleGroup == nil {
			log.Printf("[nft] 未找到规则组: %s", userGroup.RuleRef)
			continue
		}

		// 为该组中的每个用户分配规则
		for _, username := range userGroup.Users {
			userRules[username] = ruleGroup.Rules
		}
	}

	return userRules
}

// getPublicRules 获取公共规则
func getPublicRules(config *FullConfig) []RuleConfig {
	ruleGroup := resolveRuleGroup(config.Rules, config.PublicRuleRef)
	if ruleGroup == nil {
		log.Printf("[nft] 未找到公共规则组: %s", config.PublicRuleRef)
		return []RuleConfig{}
	}
	return ruleGroup.Rules
}
