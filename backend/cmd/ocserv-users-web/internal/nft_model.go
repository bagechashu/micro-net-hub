package internal

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// GlobalUsersDestRules 全局用户目标规则映射
var Global_UsersDestRules map[string][]DestRule

type DestRule struct {
	Ip           string `json:"ip,omitempty"`
	Protocol     string `json:"protocol"`
	Port         uint16 `json:"port,omitempty"`
	ToLocal      bool   `json:"to_local,omitempty"` // 是否访问宿主机本地服务
	srcIp        string
	srcIpSetName string
}

type DestRuleGroup struct {
	Name  string     `json:"name"`
	Rules []DestRule `json:"rules"`
}

type SrcIpSet struct {
	Name string   `json:"name"`
	Ips  []string `json:"ips"`
}

type DestRuleMapping struct {
	Name         string      `json:"name"`
	Type         MappingType `json:"mapping_type"` // [users | public | input_chain]
	SrcIps       []string    `json:"src_ips,omitempty"`
	SrcIpSet     SrcIpSet    `json:"src_ip_set,omitempty"`
	Users        []string    `json:"users,omitempty"`
	RuleGroupRef string      `json:"rule_group_ref"` // 引用的规则组名称
}

type Config struct {
	DestRuleGroups   []DestRuleGroup   `json:"dest_rule_groups"`
	DestRuleMappings []DestRuleMapping `json:"dest_rule_mappings"`
}

type MappingType string

const (
	TypeUsers           MappingType = "users"
	TypePublic          MappingType = "public"
	TypeInputChain      MappingType = "input_chain"
	TypeInputChainIpSet MappingType = "input_chain_ip_set"
)

func (t MappingType) Valid() bool {
	switch t {
	case TypeUsers, TypePublic, TypeInputChain, TypeInputChainIpSet:
		return true
	default:
		return false
	}
}

// -------------------- Config Manager --------------------
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	for _, mapping := range cfg.DestRuleMappings {
		if !mapping.Type.Valid() {
			log.Printf("[init] 无效的规则映射类型: %s", mapping.Type)
		}
	}

	return &cfg, nil
}

// GetUserRulesMapping 构建用户到规则的映射
func GetUserRulesMapping(config *Config) map[string][]DestRule {
	usersRules := make(map[string][]DestRule)

	for _, rgm := range config.DestRuleMappings {
		// 为每个用户组构建规则映射
		if rgm.Type != TypeUsers {
			continue
		}
		// 查找该用户组引用的规则组
		rg := resolveRuleGroup(config.DestRuleGroups, rgm.RuleGroupRef)
		if rg == nil {
			log.Printf("[init] 未找到规则组: %s", rgm.RuleGroupRef)
			continue
		}

		// 为该组中的每个用户分配规则
		for _, username := range rgm.Users {
			username = strings.ToLower(username) // 用户名转小写，保持一致性
			// 将规则追加到现有规则中，以支持多个组的规则合并
			usersRules[username] = append(usersRules[username], rg.Rules...)
		}
	}

	return usersRules
}

// GetPublicRules 获取公共规则
func GetPublicRules(config *Config) map[string][]DestRule {
	publicRules := make(map[string][]DestRule)
	for _, rulemapping := range config.DestRuleMappings {
		if rulemapping.Type != TypePublic {
			continue
		}
		ruleGroup := resolveRuleGroup(config.DestRuleGroups, rulemapping.RuleGroupRef)
		if ruleGroup == nil {
			log.Printf("[init] 未找到公共规则组: %s", rulemapping.RuleGroupRef)
			continue
		}

		rules_name := strings.ToLower(ruleGroup.Name) // 用户名转小写，保持一致性
		// 将规则追加到现有规则中，以支持多个组的规则合并
		publicRules[rules_name] = append(publicRules[rules_name], ruleGroup.Rules...)
	}

	return publicRules
}

// GetInputChainRules 获取InputChain规则
func GetInputChainRules(config *Config) map[string][]DestRule {
	inputChainRules := make(map[string][]DestRule)
	for _, rulemapping := range config.DestRuleMappings {
		if rulemapping.Type != TypeInputChain {
			continue
		}
		ruleGroup := resolveRuleGroup(config.DestRuleGroups, rulemapping.RuleGroupRef)
		if ruleGroup == nil {
			log.Printf("[init] 未找到规则组: %s", rulemapping.RuleGroupRef)
			continue
		}

		// 将规则追加到现有规则中，以支持多个组的规则合并
		for _, srcIp := range rulemapping.SrcIps {
			rules_name := srcIp // 使用源 srcIp 作为规则名称,方便 addNftRules 查找
			inputChainRules[rules_name] = append(inputChainRules[rules_name], ruleGroup.Rules...)
		}
	}

	return inputChainRules
}

// GetInputChainIpSetRules 获取InputChainIpSet规则
func GetInputChainIpSetRules(config *Config) ([]SrcIpSet, map[string][]DestRule) {
	srcIpSets := []SrcIpSet{}
	inputChainIpSetRules := make(map[string][]DestRule)
	for _, rulemapping := range config.DestRuleMappings {
		if rulemapping.Type != TypeInputChainIpSet {
			continue
		}
		srcIpSets = append(srcIpSets, rulemapping.SrcIpSet)
		ruleGroup := resolveRuleGroup(config.DestRuleGroups, rulemapping.RuleGroupRef)
		if ruleGroup == nil {
			log.Printf("[init] 未找到规则组: %s", rulemapping.RuleGroupRef)
			continue
		}

		rules_name := strings.ToLower(rulemapping.SrcIpSet.Name) // 使用集合名称作为规则名称,方便 addNftRulesIpSet 查找
		// 将规则追加到现有规则中，以支持多个组的规则合并
		inputChainIpSetRules[rules_name] = append(inputChainIpSetRules[rules_name], ruleGroup.Rules...)
	}

	return srcIpSets, inputChainIpSetRules
}

// resolveRuleGroup 根据规则组名称查找规则组
func resolveRuleGroup(groups []DestRuleGroup, name string) *DestRuleGroup {
	for _, group := range groups {
		if group.Name == name {
			return &group
		}
	}
	return nil
}
