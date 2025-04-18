# System settings --> Authentication --> LDAP

Server: ldap://127.0.0.1:389

Bind DN: cn=admin,dc=example,dc=com

Password: admin_pass

Search OU: ou=people,dc=example,dc=com

Search filter eg1: (&(uid=%(user)s)(memberOf=cn=backend,ou=allhands,dc=example,dc=com))
Search filter eg2: (&(uid=%(user)s)(|(memberOf=cn=backend,ou=allhands,dc=example,dc=com)(memberOf=cn=dba,ou=allhands,dc=example,dc=com)))

Search filter eg2 unfolding:
(&
  (uid=%(user)s)
  (|
    (memberOf=cn=backend,ou=allhands,dc=example,dc=com)
    (memberOf=cn=dba,ou=allhands,dc=example,dc=com)
  )
)


User attribute: 

```
{
  "username": "cn",
  "name": "sn",
  "email": "mail"
}

```