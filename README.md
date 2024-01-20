<!-- @format -->

# micro-net-hub

Basic tool for private network.

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
  %% Service provide by Micro-Net-Hub
  subgraph Micro-Net-Hub
    main[[Micro-Net-Hub:9000]]
    radius[[Micro-Net-Hub<br>RadiusService:1812/udp]]
    ui([Embedded-UI])

    %% Architecture
    main --> ui
    main --> CoreDnsHandler & UserHandler & VPNHandler
    UserHandler --> LDAPHandler
    UserHandler --> TOTPModule
    TOTPModule --> radius
  end
  Micro-Net-Hub --> MySQL
  coredns --> MySQL
  LDAPHandler --> openldap
  ocserv --> radius

  %% Service provide by third-party, need deployed by yourself.
  subgraph selfbuilt
   openldap[[OpenLDAP:389]]
   ocserv[[Ocserv:443]]
   coredns[[CoreDns:53/udp]]
  end

  %% Database
  subgraph MySQL
    mysql-main[(MySQL main)]
    mysql-coredns[(MySQL coredns)]
  end


```

```mermaid
---
title: Micro Net Hub Architecture
---
flowchart LR
  %% UI
  subgraph Embedded-UI
    ui-user-mgr([UserManager])
    ui-vpn-mgr([VPNManager])
    ui-coredns-mgr([CoreDnsManager])

    ui-user-mgr --> ui-user([User])
    ui-user-mgr --> ui-group([Group])
    ui-user-mgr --> ui-totp([TOTP])

    ui-vpn-mgr --> ui-vpn-config([VPN-Config])
    ui-vpn-mgr --> ui-vpn-status([VPN-Status])

    ui-coredns-mgr --> ui-coredns-config([CoreDns-Config])
    ui-coredns-mgr --> ui-coredns-status([CoreDns-Status])
  end

```
