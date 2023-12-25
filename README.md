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
    main[[<b style="color:yellow;">Micro-Net-Hub:9000</b>]]
    radius[[<b style="color:yellow;">Micro-Net-Hub<br>RadiusService:1812/udp</b>]]

    %% Architecture
    main --> LDAPController & CoreDnsController & VPNController & TOTPController
    radius --> VPNController & TOTPController

    ui([Embedded-UI])
  end
  Micro-Net-Hub --> OpenLDAP-selfbuilt
  Micro-Net-Hub --> Ocserv-selfbuilt 
  Micro-Net-Hub --> CoreDNS-selfbuilt
  Micro-Net-Hub --> MySQL
  CoreDNS-selfbuilt --> mysql-coredns

  %% Service provide by third-party, need deployed by yourself.
  subgraph OpenLDAP-selfbuilt
   openldap[[<b style="color:green;">OpenLDAP:389</b>]]
  end
  subgraph Ocserv-selfbuilt
   ocserv[[<b style="color:green;">Ocserv:443</b>]]
  end
  subgraph CoreDNS-selfbuilt
   coredns[[<b style="color:green;">CoreDns:53/udp</b>]]
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

    ui-vpn-mgr --> ui-totp([TOTPManager])
    ui-vpn-mgr --> ui-vpn-config([VPN-Config])
    ui-vpn-mgr --> ui-vpn-status([VPN-Status])

    ui-coredns-mgr --> ui-coredns-config([CoreDns-Config])
    ui-coredns-mgr --> ui-coredns-status([CoreDns-Status])
  end

```