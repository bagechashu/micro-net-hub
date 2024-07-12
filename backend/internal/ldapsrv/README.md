## Introduction
- Exclusively designed for internal network LDAP servers.
- User authentication can choose "passwords" or "password+TOTP" mechanisms.
- Only users with the role same as "BindDNRoleKeyword" setting can be used as the service account bind DN.
- Incorporates functionalities for "BIND", "SEARCH". 
    - Search filter support rules eg:
        - (&(objectClass=person)(uid=xiaoxue)(memberOf:=cn=group01,dc=example,dc=com))
        - (&(objectClass=person)(uid=xiaoxue))

## Usage

```
ldapsearch -LLL -W -x -H ldap://127.0.0.1:1389 -D "cn=admin,dc=example,dc=com" -b "dc=example,dc=com" "(&(objectClass=person)(uid=xiaoxue)(memberOf:=cn=group01,dc=example,dc=com))" 
ldapsearch -LLL -w admin_pass -x -H ldap://127.0.0.1:1389 -D "cn=admin,dc=example,dc=com" -b "dc=example,dc=com" "(&(objectClass=person)(uid=xiaoxue)(memberOf:=cn=group01,dc=example,dc=com))" 

```

## Test Result
- "anylink" success
