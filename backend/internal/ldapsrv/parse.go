package ldapsrv

import (
	"fmt"
	"micro-net-hub/internal/config"
	"strings"

	"github.com/merlinz01/ldapserver"
)

type Query struct {
	MemberOf    string
	Uid         string
	ObjectClass string
}

// TODO: https://www.rfc-editor.org/rfc/rfc4515.html
// Now, only implement the following filter:
// 1. (&(objectClass=person)(uid=xiaoxue)(memberOf:=cn=employees,dc=example,dc=com))
// 2. (&(uid=test02)(memberOf:=cn=t1,ou=allhands,dc=example,dc=com))
// x. (memberOf:=cn=t1,ou=allhands,dc=example,dc=com)
// x. (objectClass=*)
func ParseLdapQuery(query string) (*Query, error) {
	queryMap, err := parseLdapQueryToMap(query)
	if err != nil {
		return nil, err
	}
	if queryMap == nil {
		return nil, fmt.Errorf("ldap query format not support: %s", query)
	}

	q := &Query{}

	q.MemberOf = queryMap["memberOf"]
	q.Uid = queryMap["uid"]
	q.ObjectClass = queryMap["objectClass"]
	return q, nil
}

func parseLdapQueryToMap(query string) (map[string]string, error) {
	result := make(map[string]string)
	// 移除首尾可能存在的括号
	query = strings.Trim(query, "()")
	if strings.HasPrefix(query, "&") {
		// 根据")("分割查询字符串
		parts := strings.Split(query, ")(")

		for _, part := range parts {
			trimmedPart := strings.TrimSpace(part)
			// 通过检查前缀来优化性能，避免不必要的TrimPrefix调用
			trimmedPart = strings.TrimPrefix(trimmedPart, "&")
			trimmedPart = strings.TrimPrefix(trimmedPart, "(")

			// 这里可以添加对查询部分的额外验证来增强安全性
			equalParts := strings.SplitN(trimmedPart, "=", 2)
			if len(equalParts) == 2 {
				key := strings.TrimSpace(equalParts[0])
				value := strings.TrimSpace(equalParts[1])

				// 移除键末尾的冒号，如果存在的话, eg: memeberOf:=cn=employees,dc=example,dc=com
				key = strings.TrimSuffix(key, ":")

				// 验证key是否为空，增加边界条件处理
				if key == "" {
					return nil, fmt.Errorf("invalid query format: key is empty")
				}
				result[key] = value
			} else {
				// 提供详细的错误信息帮助调用者理解问题所在
				return nil, fmt.Errorf("invalid query format: %s", trimmedPart)
			}
		}
	} else {
		equalParts := strings.SplitN(query, "=", 2)
		if len(equalParts) == 2 {
			key := strings.TrimSpace(equalParts[0])
			value := strings.TrimSpace(equalParts[1])

			// 移除键末尾的冒号，如果存在的话, eg: memeberOf:=cn=employees,dc=example,dc=com
			key = strings.TrimSuffix(key, ":")

			// 验证key是否为空，增加边界条件处理
			if key == "" {
				return nil, fmt.Errorf("invalid query format: key is empty")
			}
			result[key] = value
		} else {
			// 提供详细的错误信息帮助调用者理解问题所在
			return nil, fmt.Errorf("invalid query format: %s", query)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("invalid ldap query format: %s", query)
	}
	return result, nil
}

type DN struct {
	Rdn string
	Ou  string
	DC  string
}

// cn=admin,dc=example,dc=com
// uid=user1,ou=people,dc=example,dc=com
func ParseLdapDN(dn ldapserver.DN) (*DN, error) {
	if !dn.IsSubordinate(ldapserver.MustParseDN(config.Conf.LdapServer.BaseDN)) {
		return nil, fmt.Errorf("dn is not find: %s", dn)
	}

	dnMap, err := parseLdapDNToMap(dn.String())
	if err != nil {
		return nil, err
	}
	if dnMap["uid"] == "" && dnMap["cn"] == "" {
		return nil, fmt.Errorf("failed to parse LDAP DN: %s", dnMap)
	}
	d := &DN{}
	if dnMap["uid"] != "" {
		d.Rdn = dnMap["uid"]
	} else {
		d.Rdn = dnMap["cn"]
	}
	d.Ou = dnMap["ou"]
	d.DC = dnMap["dc"]
	return d, nil
}

func parseLdapDNToMap(dn string) (map[string]string, error) {
	parts := strings.Split(dn, ",")

	// 初始化结果映射
	result := make(map[string]string)

	for _, part := range parts {
		// 分割键和值
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid LDAP DN format: %s", part)
		}

		// 将键值对添加到结果映射中
		if result[kv[0]] != "" {
			result[kv[0]] = result[kv[0]] + "." + kv[1]
		} else {
			result[kv[0]] = kv[1]
		}
	}

	return result, nil
}