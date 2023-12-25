<!-- @format -->

# micro-net-hub

Basic tool for private network.

# References

- https://github.com/eryajf/go-ldap-admin (d00d6df)
- https://github.com/eryajf/go-ldap-admin-ui (c75476d)
- https://github.com/bjdgyc/anylink
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
    main>Micro-Net-Hub] --> ui[[Embeded-UI]]
    ui --> UserManager
    ui --> VPNManager
    ui --> CoreDnsManager

    UserManager --> User
    UserManager --> Group

    VPNManager --> TOTPManager
    VPNManager --> VPN-Config
    VPNManager --> VPN-Status

    CoreDnsManager --> CoreDns-Config
    CoreDnsManager --> CoreDns-Status

    main --> srv[[Service]]
    srv --> LDAPController --> OpenLDAP
    LDAPController --> my-main[(MySQL main)]
    srv --> my-main[(MySQL main)]

    srv --> VPNController --> Ocserv --> TOTPController
    VPNController --> TOTPController --> my-totp[(MySQL totp)]

    srv --> CoreDnsController --> CoreDns --> my-coredns[(MySQL coredns)]
    CoreDnsController --> my-coredns[(MySQL coredns)]

```
