package tools

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

func ConvertToPinYin(src string) (dst string) {
	args := pinyin.NewArgs()
	args.Fallback = func(r rune, args pinyin.Args) []string {
		return []string{string(r)}
	}

	for _, singleResult := range pinyin.Pinyin(src, args) {
		for _, result := range singleResult {
			dst = dst + result
		}
	}
	return
}

func ConvertBaseDNToDomain(baseDN string) string {
	// Split the baseDN string by the commas
	parts := strings.Split(baseDN, ",")
	// Get the last two parts of the string
	parts = parts[len(parts)-2:]
	// Remove the "dc=", "cn=" and "ou="prefixes from the parts
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.TrimPrefix(parts[i], "dc=")
		// parts[i] = strings.TrimPrefix(parts[i], "cn=")
		// parts[i] = strings.TrimPrefix(parts[i], "ou=")
	}
	// Join the parts with a dot
	result := strings.Join(parts, ".")
	return result
}
