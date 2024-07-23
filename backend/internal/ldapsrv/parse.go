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
			case "memberOf", "memberof":
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

// 定义键名常量
const (
	KeyObjectClass = "objectclass"
	KeyMemberOf    = "memberof"
	KeyUid         = "uid"
	KeyCN          = "cn"
)

// TODO: https://www.rfc-editor.org/rfc/rfc4515.html
// Now, only implement the following filter:
// 1. (&(objectClass=objectClass)(uid=xiaoxue)(memberOf:=cn=employees,dc=example,dc=com))
// 2. (&(uid=test02)(memberOf:=cn=t1,ou=allhands,dc=example,dc=com))
// 3. (memberOf:=cn=t1,ou=allhands,dc=example,dc=com)
// 4. (|(objectClass=organizationalUnit)(objectClass=groupOfUniqueNames))
// 5. (&(objectclass=groupOfUniqueNames)(cn=*))
// x. (objectClass=*)
func ParseLdapQuery(query string) (*Query, error) {
	if query == "" {
		return nil, fmt.Errorf("not supported empty query string")
	}

	queryMap, err := parseLdapQueryToMap(query)
	if err != nil {
		return nil, err
	}
	if queryMap == nil {
		return nil, fmt.Errorf("not supported ldap query format: %s", query)
	}

	// 统一处理键的大小写
	for key, value := range queryMap {
		lowerKey := strings.ToLower(key)
		if lowerKey != key {
			delete(queryMap, key)
			queryMap[lowerKey] = value
		}
	}

	q := &Query{}
	q.ObjectClass = queryMap[KeyObjectClass]
	q.MemberOf = queryMap[KeyMemberOf]

	if val, ok := queryMap[KeyUid]; ok {
		q.Uid = val
	} else if val, ok := queryMap[KeyCN]; ok {
		q.Uid = val
	}

	return q, nil
}

func parseLdapQueryToMap(query string) (map[string]string, error) {
	query = strings.Trim(query, "()")
	switch {
	case strings.HasPrefix(query, "&"):
		return parseLogicalOperatorQuery(query, "&")
	case strings.HasPrefix(query, "|"):
		return parseLogicalOperatorQuery(query, "|")
	case strings.HasPrefix(query, "!"):
		return nil, fmt.Errorf("not supported query format: %s", query)
	default:
		return parseSingleQuery(query)
	}
}

func parseLogicalOperatorQuery(query string, operator string) (map[string]string, error) {
	result := make(map[string]string)
	query = trimLogicSymbol(query)
	// check subquery
	if strings.ContainsAny(query, "&|!") {
		return nil, fmt.Errorf("not supported subquery")
	}
	// 根据")("分割查询字符串
	parts := strings.Split(query, ")(")
	for _, part := range parts {
		if key, value, err := parseQueryPart(part); err != nil {
			return nil, err
		} else {
			// Handle key repetition for logical operator queries.
			if existingValues, exists := result[key]; exists {
				var builder strings.Builder
				builder.WriteString(existingValues)
				builder.WriteString(operator)
				builder.WriteString(value)
				result[key] = builder.String()
			} else {
				result[key] = value
			}
		}
	}
	return result, nil
}

func parseSingleQuery(query string) (map[string]string, error) {
	result := make(map[string]string, 1)
	equalParts := strings.SplitN(query, "=", 2)
	if len(equalParts) != 2 {
		return nil, fmt.Errorf("not supported query format: %s", query)
	}
	if key, value, err := parseQueryPart(query); err != nil {
		return nil, err
	} else {
		result[key] = value

	}
	return result, nil
}

func parseQueryPart(part string) (string, string, error) {
	// Split key=value pair.
	equalParts := strings.SplitN(part, "=", 2)
	if len(equalParts) != 2 {
		return "", "", fmt.Errorf("not supported query format: %s", part)
	}
	key := strings.TrimSpace(equalParts[0])
	value := strings.TrimSpace(equalParts[1])
	// Remove potential colon from key.
	key = strings.TrimSuffix(key, ":")
	// Validate key.
	if key == "" {
		return "", "", fmt.Errorf("not supported query format")
	}
	return key, value, nil
}

func trimLogicSymbol(query string) string {
	trimmed := strings.TrimSpace(query)
	// Remove prefix operators if present.
	trimmed = strings.TrimPrefix(trimmed, "&")
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimPrefix(trimmed, "!")
	trimmed = strings.TrimPrefix(trimmed, "(")
	return trimmed
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
