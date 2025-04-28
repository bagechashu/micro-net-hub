# Nexus LDAP Settings

### Connection
- Search base DN: ou=people,dc=example,dc=com
- Authentication Method: Simple Authentication
- Username or DN: cn=admin,dc=example,dc=com

### User and Group
- User subtree: checked
- Object Class: inetOrgPerson
- User filter: (memberOf=cn=backend,ou=groups,dc=example,dc=com)
- User ID attribute: uid
- Map LDAP groups as roles: checked
- Group Type: Dynamic Groups
- Group member of attribute: memberOf

