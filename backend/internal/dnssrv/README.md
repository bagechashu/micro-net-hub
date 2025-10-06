## reference
- https://github.com/kenshinx/godns
- https://mp.weixin.qq.com/s/G1KtkeyVQUXvU6qZBePWSA
- https://github.com/snail2sky/coredns_mysql_extend.git 

## Introduce
- only for intranet dns server
- support A/CNAME/TXT record
- use MySQL database to store domain information

## TODO
- Use state to control whether queries can be made
- implementation of dns records to add, delete, change and search

## Ocserv configuration

```
# The advertized DNS server. Use multiple lines for
# multiple servers.
dns = 192.168.0.10

## "split-dns" limitation
# "split-dns" functionality has compatibility issues with certain clients. 
# Test results show that Cisco AnyConnect on macOS does not support the split-DNS feature, 
# causing all intranet domain name resolution requests to fail 
# to be proxied to the DNS server provided by Ocserv.
# x split-dns = corp.example.com

```
