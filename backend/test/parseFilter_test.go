package test

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global/setup"
	"micro-net-hub/internal/ldapsrv"
	"testing"

	"github.com/merlinz01/ldapserver"
)

func initConfig() {
	// 加载配置文件到全局配置结构体
	config.InitConfig()

	// 初始化日志
	setup.InitLogger()
}

func TestParseLdapQuery(t *testing.T) {
	initConfig()
	tests := []struct {
		query     string
		want      *ldapsrv.Query
		wantError bool
	}{
		{
			query: "(&(objectClass=person)(uid=user01)(memberOf:=cn=employees,dc=example,dc=com))",
			want: &ldapsrv.Query{
				MemberOf:    "cn=employees,dc=example,dc=com",
				Uid:         "user01",
				ObjectClass: "person",
			},
			wantError: false,
		},
		{
			query: "(&(uid=user01)(memberOf:=cn=employees,dc=example,dc=com))",
			want: &ldapsrv.Query{
				MemberOf:    "cn=employees,dc=example,dc=com",
				Uid:         "user01",
				ObjectClass: "",
			},
			wantError: false,
		},
		{
			query: "(objectClass=*)",
			want: &ldapsrv.Query{
				MemberOf:    "",
				Uid:         "",
				ObjectClass: "*",
			},
			wantError: true,
		},
		{
			query:     "invalid_query",
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := ldapsrv.ParseLdapQuery(tt.query)
			if (err != nil) != tt.wantError {
				t.Errorf("parseLdapQuery() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && !equalQueries(got, tt.want) {
				t.Errorf("parseLdapQuery() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalQueries(a, b *ldapsrv.Query) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Uid == b.Uid && a.MemberOf == b.MemberOf && a.ObjectClass == b.ObjectClass
}

func TestParseLdapDN(t *testing.T) {
	initConfig()
	// Define test cases
	tests := []struct {
		name       string
		dn         ldapserver.DN
		wantDN     *ldapsrv.DN
		wantErr    bool
		errMessage string
	}{
		{
			name: "Valid DN uid",
			dn:   ldapserver.MustParseDN("uid=user1,ou=Users,dc=example,dc=com"),
			wantDN: &ldapsrv.DN{
				Rdn: "user1",
				Ou:  "Users",
				DC:  "example.com",
			},
			wantErr: false,
		},
		{
			name: "Valid DN cn",
			dn:   ldapserver.MustParseDN("cn=user1,dc=example,dc=com"),
			wantDN: &ldapsrv.DN{
				Rdn: "user1",
				Ou:  "",
				DC:  "example.com",
			},
			wantErr: false,
		},
		{
			name:       "Invalid DN - not suboardinatie",
			dn:         ldapserver.MustParseDN("ou=Users,dc=example1,dc=com"),
			wantDN:     nil,
			wantErr:    true,
			errMessage: "dn is not find: ou=Users,dc=example1,dc=com",
		},
		{
			name:       "Invalid DN - missing cn",
			dn:         ldapserver.MustParseDN("ou=Users,dc=example,dc=com"),
			wantDN:     nil,
			wantErr:    true,
			errMessage: "failed to parse LDAP DN: map[dc:example.com ou:Users]",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDN, err := ldapsrv.ParseLdapDN(tt.dn)

			// Check for error
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLdapDN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMessage {
				t.Errorf("parseLdapDN() error = %v, wantErrMessage %v", err.Error(), tt.errMessage)
				return
			}

			// Check for expected DN
			if !tt.wantErr && !compareDNs(gotDN, tt.wantDN) {
				t.Errorf("parseLdapDN() gotDN = %v, want %v", gotDN, tt.wantDN)
			}
		})
	}
}

func compareDNs(got *ldapsrv.DN, want *ldapsrv.DN) bool {
	if got.Rdn != want.Rdn {
		return false
	}
	if got.Ou != want.Ou {
		return false
	}
	if got.DC != want.DC {
		return false
	}
	return true
}
