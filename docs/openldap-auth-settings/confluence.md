<!-- @format -->

# confluence LDAP directory integration

#### refer

- https://www.jianshu.com/p/a33d5305ba5b
- https://blog.csdn.net/mnasd/article/details/119810951

#### Settings Example

##### Server Settings

- Directory Server Type: LDAP

##### LDAP Schema

- Base DN: dc=example,dc=com
- Additional User DN: ou=people
- Additional Group DN: ou=allhands # your config maybe "ou=groups"

##### LDAP Permissions

- Read Only, with Local Groups
- Default Group Memeberships: Confluence-users

##### Advanced Settings

- Synchronization Interval: 10 # sync user info every 10 minutes

##### User Schema Settings

User Object Class: inetorgperson
User Object Filter: (objectclass=inetorgperson)
User Name Attribute: uid
User Name RDN Attribute cn
User First Name Attribute: givenName
User Last Name Attribute: sn
User Display Name Attribute: displayName
User Email Attribute: mail
User Password Attribute: userPassword
User Password Encryption SSHA
User Unique ID Attribute uid

##### Group Schema Settings

- Group Object Class: groupOfUniqueNames
- Group Object Filter: (objectClass=groupOfUniqueNames)
- Group Name Attribute: cn
- Group Description Attribute: description

##### Membership Schema Settings

- Group Members Attribute: uniqueMember
- User Membership Attribute: memberOf
- Use the User Membership Attribute: false # unchecked, **Very import**. 组是 groupOfUniqueNames 对象，靠 uniqueMember 反向指向用户的 DN。
