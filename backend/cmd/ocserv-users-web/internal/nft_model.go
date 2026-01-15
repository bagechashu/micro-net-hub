package internal

import (
	"log"
	"strings"
)

var (
	// Global_UsersRules 全局用户目标规则映射
	Global_UsersRules map[string][]Rule
)

type Rule struct {
	DestIp   string       `json:"dest_ip,omitempty" yaml:"dest_ip,omitempty"`
	DestPort uint16       `json:"dest_port,omitempty" yaml:"dest_port,omitempty"`
	Protocol ProtocolType `json:"protocol,omitempty" yaml:"protocol,omitempty"` // tcp | udp | icmp, 默认 tcp
	ToLocal  bool         `json:"to_local,omitempty" yaml:"to_local,omitempty"` // 是否访问宿主机本地服务, 默认 false
	Action   ActionType   `json:"action,omitempty" yaml:"action,omitempty"`     // accept | drop，默认 accept

	// SrcIp, SrcIpSetName 默认不配置, 通过 RuleMapping 去补充
	SrcIp        string `json:"src_ip,omitempty" yaml:"src_ip,omitempty"`
	SrcIpSetName string `json:"src_ip_set_name,omitempty" yaml:"src_ip_set_name,omitempty"`
}

type RuleGroup struct {
	Name  string `json:"name" yaml:"name"`
	Rules []Rule `json:"rules,omitempty" yaml:"rules,omitempty"`
}

type SrcIpSet struct {
	Name string   `json:"name,omitempty" yaml:"name,omitempty"`
	Ips  []string `json:"ips,omitempty" yaml:"ips,omitempty"`
}

type RuleMapping struct {
	Name         string       `json:"name,omitempty" yaml:"name,omitempty"`
	Type         MappingType  `json:"mapping_type" yaml:"mapping_type"` // [users | public | input_chain | input_chain_ip_set]
	SrcIps       []string     `json:"src_ips,omitempty" yaml:"src_ips,omitempty"`
	SrcIpSet     *SrcIpSet    `json:"src_ip_set,omitempty" yaml:"src_ip_set,omitempty"`
	Users        []string     `json:"users,omitempty" yaml:"users,omitempty"`
	RuleGroupRef string       `json:"rule_group_ref,omitempty" yaml:"rule_group_ref,omitempty"` // 引用的规则组名称
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
		if rulemapping.SrcIpSet != nil {
			srcIpSets = append(srcIpSets, *rulemapping.SrcIpSet)
		}
		ruleGroup := resolveRuleGroup(config.RuleGroups, rulemapping.RuleGroupRef)
		if ruleGroup == nil {
			log.Printf("[init] 未找到规则组: %s", rulemapping.RuleGroupRef)
			continue
		}

		rulemappingName := strings.ToLower(rulemapping.Name)

		// dest_rule 中 srcIpSetName 赋值
		rulesWithSrcIpSetName := make([]Rule, 0, len(ruleGroup.Rules))
		for _, r := range ruleGroup.Rules {
			if rulemapping.SrcIpSet != nil {
				r.SrcIpSetName = rulemapping.SrcIpSet.Name
			}
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
