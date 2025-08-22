# nginx http context

> https://github.com/kvspb/nginx-auth-ldap

```conf
http {
    auth_ldap_cache_enabled on;
    auth_ldap_cache_expiration_time 3600;
    auth_ldap_cache_size 1000;

    ldap_server example-ldap {
        url "ldap://127.0.0.1:389/ou=people,dc=example,dc=com?uid?sub?(&(objectClass=inetorgperson))";
        binddn "cn=admin,dc=example,dc=com";
        binddn_passwd "admin_pass";
        group_attribute uniquemember;
        group_attribute_is_dn on;
        satisfy any;
        require group "cn=backend,ou=allhands,dc=example,dc=com";
        referral off;
    }
}

server {
    auth_ldap "Please contact the Admin to apply for permission.";
    auth_ldap_servers example-ldap;
}

```
