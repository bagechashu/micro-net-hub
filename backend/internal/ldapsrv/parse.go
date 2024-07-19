package ldapsrv

import (
	"fmt"
	"micro-net-hub/internal/config"
	"strings"

	"github.com/merlinz01/ldapserver"
)

type Attributes struct {
	MemberOf string
}

func parseLdapAttributes(attributes []string) *Attributes {
	if len(attributes) == 0 {
		return nil
	}
	attr := &Attributes{}
	for _, attribute := range attributes {
		if strings.Contains(attribute, "=") {
			parts := strings.SplitN(attribute, "=", 2)
			key := strings.TrimSuffix(parts[0], ":")
			switch key {
			case "memberOf":
				attr.MemberOf = parts[1]
			default:
				return nil
			}
		}
	}
	return attr
}

type Query struct {
	MemberOf    string
	Uid         string
	ObjectClass string
}

const GroupOfUniqueNamesFields = "cn"

// TODO: https://www.rfc-editor.org/rfc/rfc4515.html
// Now, only implement the following filter:
// 1. (&(objectClass=objectClass)(uid=xiaoxue)(memberOf:=cn=employees,dc=example,dc=com))
// 2. (&(uid=test02)(memberOf:=cn=t1,ou=allhands,dc=example,dc=com))
// 3. (memberOf:=cn=t1,ou=allhands,dc=example,dc=com)
// 4. (|(objectClass=organizationalUnit)(objectClass=groupOfUniqueNames))
// 5. (&(objectclass=groupOfUniqueNames)(cn=*))
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

	q.ObjectClass = queryMap["objectclass"]
	q.MemberOf = queryMap["memberof"]

	if val, ok := queryMap["uid"]; ok {
		q.Uid = val
	} else if val, ok := queryMap["cn"]; ok {
		q.Uid = val
	}
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
			if len(equalParts) != 2 {
				return nil, fmt.Errorf("invalid query format: %s", trimmedPart)
			}

			key := strings.ToLower(strings.TrimSpace(equalParts[0]))
			value := strings.TrimSpace(equalParts[1])

			// 移除键末尾的冒号，如果存在的话, eg: memeberOf:=cn=employees,dc=example,dc=com
			key = strings.TrimSuffix(key, ":")

			// 验证key是否为空，增加边界条件处理
			if key == "" {
				return nil, fmt.Errorf("invalid query format: key is empty")
			}
			result[key] = value
		}
	} else if strings.HasPrefix(query, "|") {
		// 根据")("分割查询字符串
		parts := strings.Split(query, ")(")

		for _, part := range parts {
			trimmedPart := strings.TrimSpace(part)
			// 通过检查前缀来优化性能，避免不必要的TrimPrefix调用
			trimmedPart = strings.TrimPrefix(trimmedPart, "|")
			trimmedPart = strings.TrimPrefix(trimmedPart, "(")

			// 这里可以添加对查询部分的额外验证来增强安全性
			equalParts := strings.SplitN(trimmedPart, "=", 2)
			if len(equalParts) != 2 {
				return nil, fmt.Errorf("invalid query format: %s", trimmedPart)
			}

			key := strings.ToLower(strings.TrimSpace(equalParts[0]))
			value := strings.TrimSpace(equalParts[1])

			// 移除键末尾的冒号，如果存在的话, eg: memeberOf:=cn=employees,dc=example,dc=com
			key = strings.TrimSuffix(key, ":")

			// 验证key是否为空，增加边界条件处理
			if key == "" {
				return nil, fmt.Errorf("invalid query format: key is empty")
			}
			if existingValues, exists := result[key]; exists {
				// Use strings.Builder for efficient string concatenation
				var builder strings.Builder
				builder.WriteString(existingValues)
				builder.WriteString("|")
				builder.WriteString(value)
				result[key] = builder.String()
			} else {
				key = strings.ToLower(key)
				result[key] = value
			}
		}
	} else {
		equalParts := strings.SplitN(query, "=", 2)
		if len(equalParts) == 2 {
			key := strings.ToLower(strings.TrimSpace(equalParts[0]))
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
	d := &DN{}
	if dnMap["uid"] != "" {
		d.Rdn = dnMap["uid"]
	} else if dnMap["cn"] != "" {
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
