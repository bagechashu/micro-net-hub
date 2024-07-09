package ldapsrv

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"

	"github.com/merlinz01/ldapserver"
)

func NewLdapServer() *ldapserver.LDAPServer {
	handler := &LdapSrvHandler{
		abandonment: make(map[ldapserver.MessageID]bool),
	}
	server := ldapserver.NewLDAPServer(handler)

	global.Log.Infof("New Ldap server on: %s", config.Conf.LdapServer.ListenAddr)
	return server
}

func Run() error {
	server := NewLdapServer()
	return server.ListenAndServe(config.Conf.LdapServer.ListenAddr)
}
