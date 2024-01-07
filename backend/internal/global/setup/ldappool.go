package setup

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/pkg/ldappool"
)

func InitLdapPool() {
	global.LdapPool = *ldappool.NewLdapPool(
		config.Conf.Ldap.MaxConn,
		config.Conf.Ldap.Url,
		config.Conf.Ldap.AdminDN,
		config.Conf.Ldap.AdminPass,
	)
}
