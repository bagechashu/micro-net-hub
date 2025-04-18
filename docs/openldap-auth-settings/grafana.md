# grafana LDAP settings

#### /etc/grafana/grafana.ini

```
[auth.ldap]
enabled = true
config_file = /etc/grafana/ldap.toml
allow_sign_up = true
;skip_org_role_sync = false
;sync_cron = "0 1 * * *"
;active_sync_enabled = true

```


#### /etc/grafana/ldap.toml

At your ldap server, `ou=allhands,dc=example,dc=com` may be `ou=groups,dc=example,dc=com`.

```
[[servers]]
host = "127.0.0.1"
port = 389
use_ssl = false
start_tls = false
tls_ciphers = []
min_tls_version = ""
ssl_skip_verify = false
bind_dn = "cn=admin,dc=example,dc=com"
bind_password = 'admin_pass'
timeout = 10
search_filter = "(uid=%s)"
search_base_dns = ["ou=people,dc=example,dc=com"]
group_search_filter = "(objectClass=groupOfUniqueNames)"
group_search_base_dns = ["ou=allhands,dc=example,dc=com"]
group_search_filter_user_attribute = "uid"
[servers.attributes]
name = "givenName"
surname = "sn"
username = "cn"
member_of = "memberOf"
email =  "email"
[[servers.group_mappings]]
group_dn = "cn=admin,ou=allhands,dc=example,dc=com"
org_role = "Admin"
[[servers.group_mappings]]
group_dn = "cn=backend,ou=allhands,dc=example,dc=com"
org_role = "Editor"
[[servers.group_mappings]]
group_dn = "cn=dba,ou=allhands,dc=example,dc=com"
org_role = "Editor"
[[servers.group_mappings]]
group_dn = "*"
org_role = "Viewer"

```

