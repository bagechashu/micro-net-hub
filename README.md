<!-- @format -->

<pre style="font-size: 32px; font-weight: bold; background-color: #000000; display: inline-table; text-align: center; line-height: 52px; padding: 12px 10px; border-radius: 5px;">
<span style="color: #ffffff; display: inline-block; padding: 0 8px;">Micro</span>
<span style="color: #000000; background-color: #eeb968; display: inline-block; padding: 0 10px; border-radius: 10px;">Net</span>
<span style="color: #ffffff; display: inline-block; padding: 0 8px;">Hub</span>
</pre>

A tool for managing your OpenLDAP/Ocserv/Navigation at a private network.

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

# TODO

- VPNManager
- CoreDnsManager

# Architechture

```mermaid
---
title: Micro Net Hub Architecture
---
flowchart LR
  %% Service provide by third-party, need deployed by yourself.
  subgraph selfbuilt OpenLDAP
   openldap[[OpenLDAP:389]]
  end

  subgraph selfbuilt Ocserv
   ocserv[[Ocserv:443]]
  end

  subgraph 3rd-Services
   gitlab[[Gitlab]]
   nexus[[Nexus]]
   other[[...]]
  end
  %% Database
  subgraph MySQL
    mysql-main[(MySQL main)]
  end

  subgraph Micro-Net-Hub
    main --> ui
    main --> SiteNavigationManager
    main[[Micro-Net-Hub:9000]]
    ui([Embedded-UI])

    main --> UserManager
    UserManager --> TOTPModule

    radius[[Micro-Net-Hub<br>RadiusService:1812/udp]]
    radius --> TOTPModule

    UserManager ---> GoLDAPAdmin
  end

  Micro-Net-Hub --> MySQL
  GoLDAPAdmin --> openldap
  3rd-Services --> openldap
  ocserv --> radius

```
