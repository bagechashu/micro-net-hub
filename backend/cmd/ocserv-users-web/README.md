# nftables commands examples

```
nft list ruleset
nft list table inet vpn_filter
nft list chain inet vpn_filter vpn_forward

nft flush ruleset 
nft delete table inet filter
nft delete table ip vpn_filter
nft flush chain ip vpn_filter vpn_forward

nft delete chain ip vpn_filter vpn_forward
nft delete table ip vpn_filter

nft add table ip vpn_filter
nft add chain ip vpn_filter vpn_forward
nft add rule ip vpn_filter vpn_forward counter accept

nft add rule ip vpn_filter vpn_input ip saddr 1.2.3.4 tcp dport 22 accept
nft add rule ip vpn_filter vpn_input ip saddr 1.2.3.4 meta l4proto tcp accept

nft add rule vpn_filter vpn_input tcp dport 22 accept
nft add rule vpn_filter vpn_input meta l4proto tcp accept
nft add rule vpn_filter vpn_input iif lo accept comment "allow_loopback"

```

# nftables_permission_grant.sh

```sh
#!/bin/bash

GREEN="#08d417"
RED="#ff0000"

if [[ $REASON == "connect" ]]; then
  REASON_COLOR="<font color=\"$GREEN\"> $REASON </font>"
elif [[ $REASON == "disconnect" ]]; then
  REASON_COLOR="<font color=\"$RED\"> $REASON </font>"
else
  REASON_COLOR=$REASON
fi

CONNECT_INFO="[notice] $USERNAME\($IP_REAL\) $IP_REMOTE $REASON_COLOR Ocserv VPN"

/usr/bin/curl --connect-timeout 5 -XPOST "http://:9000/webhook/raw/ding?secret=securitykey" -d "${CONNECT_INFO}" >/dev/null 2>&1

(/usr/bin/curl -XPOST http://127.0.0.1:8080/vpnaccess 2>/dev/null &)
sleep 3
(/usr/bin/curl -XPOST http://127.0.0.1:8080/nftables 2>/dev/null &)

```


# ocserv configuration about script

```conf

# Script to call when a client connects and obtains an IP.
# The following parameters are passed on the environment.
# REASON, USERNAME, GROUPNAME, DEVICE, IP_REAL (the real IP of the client),
# IP_REAL_LOCAL (the local interface IP the client connected), IP_LOCAL
# (the local IP in the P-t-P connection), IP_REMOTE (the VPN IP of the client),
# IPV6_LOCAL (the IPv6 local address if there are both IPv4 and IPv6
# assigned), IPV6_REMOTE (the IPv6 remote address), IPV6_PREFIX, and
# ID (a unique numeric ID); REASON may be "connect" or "disconnect".
# In addition the following variables OCSERV_ROUTES (the applied routes for this
# client), OCSERV_NO_ROUTES, OCSERV_DNS (the DNS servers for this client),
# will contain a space separated list of routes or DNS servers. A version
# of these variables with the 4 or 6 suffix will contain only the IPv4 or
# IPv6 values. The connect script must return zero as exit code, or the
# client connection will be refused.

# The disconnect script will receive the additional values: STATS_BYTES_IN,
# STATS_BYTES_OUT, STATS_DURATION that contain a 64-bit counter of the bytes
# output from the tun device, and the duration of the session in seconds.

connect-script = /etc/ocserv/nftables_permission_grant.sh
disconnect-script = /etc/ocserv/nftables_permission_grant.sh


# Note the that following two firewalling options currently are available
# in Linux systems with iptables software.

# If set, the script /usr/bin/ocserv-fw will be called to restrict
# the user to its allowed routes and prevent him from accessing
# any other routes. In case of defaultroute, the no-routes are restricted.
# All the routes applied by ocserv can be reverted using /usr/bin/ocserv-fw
# --removeall. This option can be set globally or in the per-user configuration.
#restrict-user-to-routes = true

# This option implies restrict-user-to-routes set to true. If set, the
# script /usr/bin/ocserv-fw will be called to restrict the user to
# access specific ports in the network. This option can be set globally
# or in the per-user configuration.
#restrict-user-to-ports = "tcp(443), tcp(80), udp(443), sctp(99), tcp(583), icmp(), icmpv6()"

# You could also use negation, i.e., block the user from accessing these ports only.
#restrict-user-to-ports = "!(tcp(443), tcp(80))"

# When set to true, all client's iroutes are made visible to all
# connecting clients except for the ones offering them. This option
# only makes sense if config-per-user is set.
#expose-iroutes = true

```

# /etc/systemd/system/ocserv-users-web.service

```conf

[Unit]
Description=ocserv-users-web
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/ocserv-users -config /etc/ocserv/rules.json -webaddr ":8080"
Restart=on-failure
User=root
Environment=HOME=/root
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target

```