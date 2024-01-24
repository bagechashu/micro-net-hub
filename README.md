<!-- @format -->

# micro-net-hub

A tool for managing your OpenLDAP/Ocserv/Navigation at a private network.

# How to set Ocserv Authentication with Radius.

> https://ocserv.openconnect-vpn.net/recipes-ocserv-authentication-radius-radcli.html

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
