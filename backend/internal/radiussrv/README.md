<!-- @format -->

# Radius Server

Based on [golang-radius-server-ldap-with-mfa](https://github.com/fivexl/golang-radius-server-ldap-with-mfa).

# How to set Ocserv Authentication with Radius.

> https://ocserv.openconnect-vpn.net/recipes-ocserv-authentication-radius-radcli.html

Key Configuration

```
# /etc/radcli/radiusclient.conf
authserver 192.168.5.5 # 192.168.5.5 is Radius service that build in Micro-Net-Hub.
servers /etc/radcli/servers

# /etc/radcli/servers
192.168.5.5 default-radius-secret  # You can change it at config.yml

# /etc/ocserv/ocserv.conf
auth = "radius [config=/usr/local/etc/radcli/radiusclient.conf,groupconfig=true]"

```

After setting Ocserv auth with Micro-net-hub 1812 radius service, you can login Ocserv VPN with Micro-net-hub accouts

> eg:

```
VPN Username: Username (admin)
VPN Password: Password+OTPcode (admin_pass000000)

```

# MFA otp tools (Scan QRcode which at Profile page)：

- https://github.com/andOTP/andOTP（OpenSource Software，recommended）
- google authenticator
- 阿里云 手机客户端
- ...

# Ocserv VPN Client

- https://www.infradead.org/openconnect/
- https://github.com/openconnect/openconnect
- https://github.com/openconnect/openconnect-gui

# Testing

First get server up and running

```
go run cmd/micro-net-hub/main.go
```

For the test we need a client

```
# Ubuntu
sudo apt-get install freeradius-utils
# MacOS
brew install freeradius-server
# test
#eg: radtest <username> <password(admin_pass)+OTP(000000)> localhost 1812 <radius secret>
radtest admin admin_pass000000 127.0.0.1 1812 default-radius-secret

```

Checking packets sent to server

```
sudo tshark -f "udp port 1812" -i any -V
```
