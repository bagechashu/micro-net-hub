# sonar.properties

```yaml
# LDAP CONFIGURATION

# Enable the LDAP feature
sonar.security.realm=LDAP

# Set to true when connecting to a LDAP server using a case-insensitive setup.
# sonar.authenticator.downcase=true

# URL of the LDAP server. Note that if you are using ldaps, then you should install the server certificate into the Java truststore.
ldap.url=ldap://127.0.0.1:389

# Bind DN is the username of an LDAP user to connect (or bind) with. Leave this blank for anonymous access to the LDAP directory (optional)
ldap.bindDn=cn=admin,dc=example,dc=com

# Bind Password is the password of the user to connect with. Leave this blank for anonymous access to the LDAP directory (optional)
ldap.bindPassword=admin_pass

# Possible values: simple | CRAM-MD5 | DIGEST-MD5 | GSSAPI See http://java.sun.com/products/jndi/tutorial/ldap/security/auth.html (default: simple)
ldap.authentication=simple

# USER MAPPING

# Distinguished Name (DN) of the root node in LDAP from which to search for users (mandatory)
ldap.user.baseDn=dc=example,dc=com

# LDAP user request. (default: (&(objectClass=inetOrgPerson)(uid={login})) )
ldap.user.request=(&(objectClass=inetOrgPerson)(uid={login}))

# Attribute in LDAP defining the user’s real name. (default: cn)
ldap.user.realNameAttribute=cn

# Attribute in LDAP defining the user’s email. (default: mail)
ldap.user.emailAttribute=mail

# GROUP MAPPING

# Distinguished Name (DN) of the root node in LDAP from which to search for groups. (optional, default: empty)
ldap.group.baseDn=ou=groups,dc=example,dc=com

# LDAP group request (default: (&(objectClass=groupOfUniqueNames)(uniqueMember={dn})) )
ldap.group.request=(&(|(objectClass=organizationalUnit)(objectClass=groupOfUniqueNames))(memberOf={dn}))

# Property used to specifiy the attribute to be used for returning the list of user groups in the compatibility mode. (default: cn)
# ldap.group.idAttribute=cn

# ldap.allowAnonymous=false
# ldap.realm=AD

```