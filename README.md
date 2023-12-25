<!-- @format -->

# micro-net-hub

Basic tool for private network.

# References

- https://github.com/eryajf/go-ldap-admin (d00d6df)
- https://github.com/eryajf/go-ldap-admin-ui (c75476d)
- https://github.com/bjdgyc/anylink
- https://github.com/fivexl/golang-radius-server-ldap-with-mfa
- https://github.com/lework/lenav

# TODO

- UserManager
  - Delete Default Value of departmentNumber: 打工人; postalAddress: 地球.
  - Add comment for group entity of ou/cn.
- TOTPManager
- VPNManager
- CoreDnsManager

# Architechture

```mermaid
---
title: Micro Net Hub Architecture
---
flowchart LR
  %% Main
  subgraph Micro-Net-Hub
    %% Service provide by Micro-Net-Hub
    main[[Micro-Net-Hub:9000]]
    radius[[Micro-Net-Hub<br>RadiusService:1812/udp]]

    %% Architecture
    main --> CoreDnsController & LDAPController & VPNController & TOTPController
    radius --> VPNController & TOTPController
    ui([Embedded-UI])
  end
  Micro-Net-Hub ---> selfbuilt
  Micro-Net-Hub --> mysql-main 
  CoreDnsController --> mysql-coredns
  coredns --> mysql-coredns

  %% Service provide by third-party, need deployed by yourself.
  subgraph selfbuilt
   openldap[[OpenLDAP:389]]
   ocserv[[Ocserv:443]]
   coredns[[CoreDns:53/udp]]
  end

  %% Database
  subgraph MySQL
    mysql-coredns[(MySQL coredns)]
    mysql-main[(MySQL main)]
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

    ui-vpn-mgr --> ui-totp([TOTPManager])
    ui-vpn-mgr --> ui-vpn-config([VPN-Config])
    ui-vpn-mgr --> ui-vpn-status([VPN-Status])

    ui-coredns-mgr --> ui-coredns-config([CoreDns-Config])
    ui-coredns-mgr --> ui-coredns-status([CoreDns-Status])
  end

```