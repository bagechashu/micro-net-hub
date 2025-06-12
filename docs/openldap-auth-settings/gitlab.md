# gitlab.rb

```rb
gitlab_rails['ldap_enabled'] = true
gitlab_rails['ldap_servers'] = YAML.load <<-'EOS'
  main: # 'main' is the GitLab 'provider ID' of this LDAP server
    label: 'LDAP'
    host: '127.0.0.1'
    port: 389
    uid: 'uid'
    bind_dn: 'cn=admin,dc=example,dc=com'
    password: 'admin_pass'
    encryption: 'plain'
    verify_certificates: false
    active_directory: true
    allow_username_or_email_login: true
    lowercase_usernames: false
    block_auto_created_users: false
    base: 'ou=people,dc=example,dc=com'
    user_filter: '(memberOf=cn=all,ou=allhands,dc=example,dc=com)'
    ## EE only
    group_base: 'ou=allhands,dc=example,dc=com'
    admin_group: ''
    sync_ssh_keys: true
EOS

```
