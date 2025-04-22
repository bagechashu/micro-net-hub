<!-- @format -->

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
group_search_filter = "(&(objectClass=groupOfUniqueNames)(uniqueMember=%s))"
group_search_base_dns = ["ou=allhands,dc=example,dc=com"]
group_search_filter_user_attribute = "dn"
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

```

Possible base_dn:

```
search_base_dns = ["ou=users,dc=example,dc=com"]
group_search_base_dns = ["ou=groups,dc=example,dc=com"]

```

#### Explanation

Micro-Net-Hub 管理下的 OpenLDAP 环境是 标准 OpenLDAP 风格，
用户条目里是没有 memberOf 这个属性，
组是 groupOfUniqueNames 对象，靠 uniqueMember 反向指向用户的 DN。

也就是说：
用户对象本身 不含组信息。
组对象中通过 uniqueMember 指定哪些用户属于它。

因此，Grafana 的默认 memberOf 逻辑是不适用您的环境的。
我们需要启用 Grafana 的 group search（让 Grafana 反向去组里找当前用户），而不是指望从用户信息里拿到 group。

关键配置:

```toml
group_search_filter = "(&(objectClass=groupOfUniqueNames)(uniqueMember=%s))"
group_search_filter_user_attribute = "dn"       # 注意！一定是 dn，不是 uid，不是 cn.
```

OpenLDAP Log

```log
6807a029 conn=1981 fd=14 ACCEPT from IP=127.0.0.1:49138 (IP=0.0.0.0:389)
6807a029 conn=1981 op=0 BIND dn="cn=admin,dc=example,dc=com" method=128
6807a029 conn=1981 op=0 BIND dn="cn=admin,dc=example,dc=com" mech=SIMPLE ssf=0
6807a029 conn=1981 op=0 RESULT tag=97 err=0 text=
6807a029 conn=1981 op=1 SRCH base="ou=people,dc=example,dc=com" scope=2 deref=0 filter="(|(uid=father.god))"
6807a029 conn=1981 op=1 SRCH attr=cn sn email givenName memberOf dn
6807a029 conn=1981 op=1 SEARCH RESULT tag=101 err=0 nentries=1 text=
6807a029 conn=1981 op=2 SRCH base="ou=allhands,dc=example,dc=com" scope=2 deref=0 filter="(&(objectClass=groupOfUniqueNames)(uniqueMember=uid=father.god,ou=people,dc=example,dc=com))"
6807a029 conn=1981 op=2 SRCH attr=dn
6807a029 <= mdb_equality_candidates: (uniqueMember) not indexed
6807a029 conn=1981 op=2 SEARCH RESULT tag=101 err=0 nentries=2 text=
6807a029 conn=1981 fd=14 closed (connection lost)

```
