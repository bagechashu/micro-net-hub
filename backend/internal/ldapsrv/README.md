## Introduction
- Exclusively designed for internal network LDAP servers.
- Incorporates functionalities for binding, searching, and filtering.
- User authentication need both static passwords and TOTP (Time-based One-Time Password) mechanisms.

## LDAP command

```
ldapsearch -LLL -w password -x -H ldap://127.0.0.1:1389 -D "cn=admin,dc=example,dc=com" -b "dc=example,dc=com" "(groupOfUniqueNames=employees)" 

```
