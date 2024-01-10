<!-- @format -->

# Radius Server

Based on [golang-radius-server-ldap-with-mfa](https://github.com/fivexl/golang-radius-server-ldap-with-mfa).

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
radtest <username> <password> localhost 1812 <radius secret>
radtest admin admin_pass 127.0.0.1 1812 default-radius-secret

```

Checking packets sent to server

```
sudo tshark -f "udp port 1812" -i any -V
```
