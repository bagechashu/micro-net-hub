<!-- @format -->
![](docs/logo/micro-net-hub.png)

A tool for managing your OpenLDAP/DNS/Navigation at a private network.

# How to install Micro-Net-Hub

[Click to see the doc](docs/README.md)

# How to set Ocserv Authentication with Radius which build in Micro-Net-Hub.

[Click to see the doc](backend/internal/radiussrv/README.md)

# How to set OpenLdap and IM manager

[Go-LDAP-Admin - eryajf](http://ldapdoc.eryajf.net/pages/5683c6/#%E5%88%9D%E5%A7%8B%E6%95%B0%E6%8D%AE)

# References

- https://github.com/eryajf/go-ldap-admin
- https://github.com/gnimli/go-web-mini
- https://github.com/LyricTian/gin-admin
- https://github.com/go-admin-team/go-admin
- https://github.com/m-vinc/go-ldap-pool
- https://github.com/bjdgyc/anylink
- https://github.com/fivexl/golang-radius-server-ldap-with-mfa
- https://github.com/lework/lenav
- https://github.com/kenshinx/godns
- https://github.com/snail2sky/coredns_mysql_extend.git 

# Architechture

```mermaid
---
title: Micro Net Hub Architecture
---
flowchart LR
  %% Database
  subgraph MySQL
    mysql-main[(MySQL main)]
  end

  subgraph Micro-Net-Hub
    admin[[AdminControlPanel:9000]]
    ui([Embedded-UI])

    admin --> UserManager
    UserManager --> TOTPModule

    admin --> ui
    admin --> SiteNavigationManager
    admin --> DNSManager
    admin --> NoticeManager
    
    radius[[Micro-Net-Hub<br>RadiusService:1812/udp]]
    embdns[[Micro-Net-Hub<br>DnsService:53/udp,tcp]]
    embLdap[[Micro-Net-Hub<br>LdapService:1389/tcp]]
    embLdapWithTotpVerify[[Micro-Net-Hub<br>LdapService:1390/tcp]]

    GoLDAPAdmin[GoLDAPAdmin<br>Sync Organization from<br>DingTalk, Wechat, Feishu, Openldap]
  end

  Micro-Net-Hub --> MySQL

  %% Service provide by third-party, need deployed by yourself.
  subgraph selfbuilt OpenLDAP
   openldap[[OpenLDAP:389<br>Not necessary]]
  end

  Micro-Net-Hub -.-> openldap
```
