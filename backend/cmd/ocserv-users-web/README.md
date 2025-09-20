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

```

