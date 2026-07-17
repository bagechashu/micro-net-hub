
```yaml
# /config/configuration.yml

# 极其关键的 LDAP 认证后端配置
authentication_backend:
  ldap:
    implementation: custom
    address: 'ldaps://127.0.0.1:636'
    tls:
      skip_verify: true
    base_dn: 'dc=example,dc=com'
    additional_users_dn: 'ou=people'
    users_filter: '(&({username_attribute}={input})(objectClass=inetOrgPerson))'
    additional_groups_dn: 'ou=groups'

    # 使用 filter mode, 复杂度最坏为 O(N)
    #group_search_mode: 'filter'
    #groups_filter: '(&(uniqueMember={dn})(objectClass=groupOfUniqueNames))'

    # 使用 memberof  mode, 复杂度为 O(1)
    group_search_mode: 'memberof'
    groups_filter: '(&(objectClass=groupOfUniqueNames)(|{memberof:rdn}))'

    user: 'CN=admin,DC=example,DC=com'
    password: 'ldap_admin_password'
    attributes:
      username: 'uid'
      display_name: 'displayName'
      mail: 'mail'
      member_of: 'memberOf'
      group_name: 'cn'
      # group_server_mode=memberof时， 必须要有 distinguished_name
      distinguished_name: 'entryDN'

access_control:
  default_policy: 'deny'
  rules:
    - domain: 'admin.example.com'
      policy: 'one_factor'
      subject:
        - 'group:admin'
    - domain_regex:
        - '^test-.*\.example\.com$'
      policy: 'one_factor'
      subject: 'group:backend'

```
