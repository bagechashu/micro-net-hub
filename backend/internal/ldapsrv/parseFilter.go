package ldapsrv

import (
	"fmt"
	"strings"
)

type Query struct {
	MemberOf    string
	Uid         string
	ObjectClass string
}

// (&(objectClass=person)(uid=xiaoxue)(memberOf:=cn=employees,dc=example,dc=com))
func parseLdapQuery(query string) (*Query, error) {
	queryMap, err := parseLdapQueryToMap(query)
	if err != nil {
		return nil, err
	}
	q := &Query{
		MemberOf:    queryMap["memberOf"],
		Uid:         queryMap["uid"],
		ObjectClass: queryMap["objectClass"],
	}
	return q, nil
}

func parseLdapQueryToMap(query string) (map[string]string, error) {
	// 移除首尾可能存在的括号
	query = strings.Trim(query, "()")
	// 根据")("分割查询字符串
	parts := strings.Split(query, ")(")

	result := make(map[string]string)

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

	return result, nil
}
