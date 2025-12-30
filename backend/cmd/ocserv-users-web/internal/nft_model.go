package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

// Global_UsersRules 全局用户目标规则映射
var Global_UsersRules map[string][]Rule

type Rule struct {
	DestIp       string       `json:"dest_ip,omitempty"`
	DestPort     uint16       `json:"dest_port,omitempty"`
	Protocol     ProtocolType `json:"protocol,omitempty"` // tcp | udp | icmp, 默认 tcp
	ToLocal      bool         `json:"to_local,omitempty"` // 是否访问宿主机本地服务, 默认 false
	Action       ActionType   `json:"action,omitempty"`   // accept | drop，默认 accept
	SrcIp        string
	SrcIpSetName string
}

type RuleGroup struct {
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

type SrcIpSet struct {
	Name string   `json:"name"`
	Ips  []string `json:"ips"`
}

type RuleMapping struct {
	Name         string      `json:"name"`
	Type         MappingType `json:"mapping_type"` // [users | public | input_chain]
	SrcIps       []string    `json:"src_ips,omitempty"`
	SrcIpSet     SrcIpSet    `json:"src_ip_set,omitempty"`
	Users        []string    `json:"users,omitempty"`
	RuleGroupRef string      `json:"rule_group_ref"` // 引用的规则组名称
}

type Config struct {
	RuleGroups   []RuleGroup   `json:"rule_groups"`
	RuleMappings []RuleMapping `json:"rule_mappings"`
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
		if mapping.SrcIpSet.Name != "" {
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

type ProtocolType string

const (
	ProtocolTcp  ProtocolType = "tcp"
	ProtocolUdp  ProtocolType = "udp"
	ProtocolIcmp ProtocolType = "icmp"
)

func (t ProtocolType) Valid() bool {
	switch t {
	case ProtocolTcp, ProtocolUdp, ProtocolIcmp:
		return true
	default:
		return false
	}
}

type ActionType string

const (
	ActionAccept ActionType = "accept"
	ActionDrop   ActionType = "drop"
)

func (t ActionType) Valid() bool {
	switch t {
	case ActionAccept, ActionDrop:
		return true
	default:
		return false
	}
}

type MappingType string

const (
	MappingUsers           MappingType = "users"
	MappingPublic          MappingType = "public"
	MappingInputChain      MappingType = "input_chain"
	MappingInputChainIpSet MappingType = "input_chain_ip_set"
)

func (t MappingType) Valid() bool {
	switch t {
	case MappingUsers, MappingPublic, MappingInputChain, MappingInputChainIpSet:
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

// GetUserRulesMapping 构建用户到规则的映射
func GetUserRulesMapping(config *Config) map[string][]Rule {
	usersRules := make(map[string][]Rule)

	for _, rgm := range config.RuleMappings {
		// 为每个用户组构建规则映射
		if rgm.Type != MappingUsers {
			continue
		}
		// 查找该用户组引用的规则组
		rg := resolveRuleGroup(config.RuleGroups, rgm.RuleGroupRef)
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
func GetPublicRules(config *Config) map[string][]Rule {
	publicRules := make(map[string][]Rule)
	for _, rulemapping := range config.RuleMappings {
		if rulemapping.Type != MappingPublic {
			continue
		}
		ruleGroup := resolveRuleGroup(config.RuleGroups, rulemapping.RuleGroupRef)
		if ruleGroup == nil {
			log.Printf("[init] 未找到公共规则组: %s", rulemapping.RuleGroupRef)
			continue
		}

		rulemappingName := strings.ToLower(rulemapping.Name) // 用户名转小写，保持一致性
		// 将规则追加到现有规则中，以支持多个组的规则合并
		publicRules[rulemappingName] = append(publicRules[rulemappingName], ruleGroup.Rules...)
	}

	return publicRules
}

// GetInputChainRules 获取InputChain规则
func GetInputChainRules(config *Config) map[string][]Rule {
	inputChainRules := make(map[string][]Rule)
	for _, rulemapping := range config.RuleMappings {
		if rulemapping.Type != MappingInputChain {
			continue
		}
		ruleGroup := resolveRuleGroup(config.RuleGroups, rulemapping.RuleGroupRef)
		if ruleGroup == nil {
			log.Printf("[init] 未找到规则组: %s", rulemapping.RuleGroupRef)
			continue
		}

		rulemappingName := strings.ToLower(rulemapping.Name)

		// dest_rule 中 srcIp 赋值
		for _, srcIp := range rulemapping.SrcIps {
			// 为每个 srcIp 生成一份带有 SrcIp 的规则拷贝
			rulesWithSrc := make([]Rule, 0, len(ruleGroup.Rules))
			for _, r := range ruleGroup.Rules {
				r.SrcIp = srcIp
				rulesWithSrc = append(rulesWithSrc, r)
			}

			// 将规则追加到现有规则中，以支持多个组的规则合并
			inputChainRules[rulemappingName] = append(inputChainRules[rulemappingName], rulesWithSrc...)
		}
	}

	return inputChainRules
}

// GetInputChainIpSetRules 获取InputChainIpSet规则
func GetInputChainIpSetRules(config *Config) ([]SrcIpSet, map[string][]Rule) {
	srcIpSets := []SrcIpSet{}
	inputChainIpSetRules := make(map[string][]Rule)
	for _, rulemapping := range config.RuleMappings {
		if rulemapping.Type != MappingInputChainIpSet {
			continue
		}
		srcIpSets = append(srcIpSets, rulemapping.SrcIpSet)
		ruleGroup := resolveRuleGroup(config.RuleGroups, rulemapping.RuleGroupRef)
		if ruleGroup == nil {
			log.Printf("[init] 未找到规则组: %s", rulemapping.RuleGroupRef)
			continue
		}

		rulemappingName := strings.ToLower(rulemapping.Name)

		// dest_rule 中 srcIpSetName 赋值
		rulesWithSrcIpSetName := make([]Rule, 0, len(ruleGroup.Rules))
		for _, r := range ruleGroup.Rules {
			r.SrcIpSetName = rulemapping.SrcIpSet.Name
			rulesWithSrcIpSetName = append(rulesWithSrcIpSetName, r)
		}

		// 将规则追加到现有规则中，以支持多个组的规则合并
		inputChainIpSetRules[rulemappingName] = append(inputChainIpSetRules[rulemappingName], rulesWithSrcIpSetName...)
	}

	return srcIpSets, inputChainIpSetRules
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
