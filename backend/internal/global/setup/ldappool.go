package setup

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/pkg/ldappool"
	"micro-net-hub/internal/tools"
)

func InitLdapPool() {
	global.LdapPool = *ldappool.NewLdapPool(
		config.Conf.Ldap.MaxConn,
		config.Conf.Ldap.Url,
		config.Conf.Ldap.AdminDN,
		tools.NewParsePasswd(config.Conf.Ldap.AdminPass),
	)
}
