# radius-client

micro-net-hub 内置 RADIUS 认证服务(`internal/radiussrv`)的命令行测试客户端.

相当于 freeradius-utils 的 `radtest`, 但:

- 无需安装额外依赖, 直接 `go run` 即可
- 密码按本项目的约定拼装为 **数据库密码 + TOTP 六位验证码**(服务端取最后 6 位作为 TOTP)
- 可携带 `NAS-Identifier` / `NAS-IP-Address` 等属性, 便于验证人工审批单上的留痕信息
- 支持连续多次请求, 便于验证 "失败次数锁定 5 分钟" 与 "人工审批" 流程
- 密码在输出中自动脱敏, 不会把明文写进终端或 CI 日志

## 快速开始

服务端先跑起来(需先准备好 MySQL 与 LDAP 数据):

```shell
cd backend
go run cmd/micro-net-hub/main.go
```

另开一个终端发认证请求:

```shell
# -pass 填数据库密码, -otp 填认证器中当前的 6 位验证码
go run ./cmd/radius-client -user admin -pass admin_pass -otp 000000
```

若密码与 TOTP 都正确, 输出 `[1/1] Access-Accept`, 退出码为 `0`.

## 参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-server` | `127.0.0.1:1812` | 服务端地址, 支持 `host` 或 `host:port`, 对应 `radius.listen-addr` |
| `-secret` | `default-radius-secret` | 共享密钥, 需与 `config.yml` 中 `radius.secret` 一致 |
| `-user` | 空 | 用户名, 也可作为第一个位置参数 |
| `-pass` | 空 | 数据库密码(不含 TOTP), 也可作为第二个位置参数 |
| `-otp` | 空 | TOTP 六位验证码, 直接拼接到 `-pass` 之后 |
| `-nas-id` | `micro-net-hub-radius-client` | `NAS-Identifier` 属性, 留空则不发送 |
| `-nas-ip` | 空 | `NAS-IP-Address` 属性, 留空时自动探测本机出接口地址 |
| `-nas-port` | `0` | `NAS-Port` 属性, `0` 表示不发送 |
| `-timeout` | `30s` | 单次请求总超时, 需大于 `radius.approval.wait-seconds` |
| `-retry` | `0` | 报文重传间隔, `0` 表示不重传; 设为 `1s` 可模拟 ocserv/radcli 的重试 |
| `-count` | `1` | 请求次数 |
| `-interval` | `1s` | 多次请求之间的间隔 |
| `-verbose` | `false` | 打印请求与响应的全部属性 |

## 退出码

| 退出码 | 含义 |
| --- | --- |
| `0` | 全部请求收到 `Access-Accept` |
| `1` | 收到 `Access-Reject`(或其他非 Accept 响应) |
| `2` | 请求失败: 无响应 / 超时 / 报文非法, 或参数错误 |

出现传输层错误时优先返回 `2`; 参数错误也返回 `2`.

## 常见测试场景

### 1. 基本认证与报文排错

```shell
go run ./cmd/radius-client -user admin -pass admin_pass -otp 000000 -verbose
```

`-verbose` 会打印请求属性(如 `NAS-Identifier`、`Service-Type`)与响应属性. 认证被拒时服务端只返回
`Access-Reject`, 具体原因需查看 micro-net-hub 日志中的 `Error while performing auth for user`.

### 2. 触发失败次数锁定

`config.yml` 中 `radius.fail-times-before-block5min` 默认 9 次(真实客户端失败后会重试, 因此密码错
3 次左右即可能被锁定). 用 `-count` 连续发错密码即可复现:

```shell
go run ./cmd/radius-client -user admin -pass wrong_pass -otp 000000 -count 5 -interval 1s
```

配合 `-retry 1s` 可模拟 ocserv/radcli 的重传行为, 观察重传对失败计数的影响:

```shell
go run ./cmd/radius-client -user admin -pass wrong_pass -otp 000000 -retry 1s -timeout 5s
```

### 3. 验证人工审批流程

开启 `radius.approval.enable` 并命中时间窗口后, 服务端会等待审批人(默认 `wait-seconds: 6`), 因此
`-timeout` 必须大于该值:

```shell
go run ./cmd/radius-client -user admin -pass admin_pass -otp 000000 \
    -nas-id ocserv-1 -nas-ip 192.168.5.10 -timeout 60s -count 2 -interval 5s
```

- 审批通过: 本次请求返回 `Access-Accept`, 且有效期内再次连接会被放行凭证直接放行
- 审批被拒: 返回 `Access-Reject`, 冷却期内重试会被 `reject-cooldown-seconds` 限流
- 等待超时: 请求无响应(客户端超时), 审批单仍然有效, 审批通过后重连即可

### 4. 与 radtest 的对照

```shell
# 等价于: radtest admin admin_pass000000 127.0.0.1 1812 default-radius-secret
go run ./cmd/radius-client -user admin -pass admin_pass -otp 000000
go run ./cmd/radius-client admin admin_pass000000        # 位置参数写法
```

## 测试

客户端自身带有单元测试与端到端测试(内置一个假 RADIUS 服务端, 不依赖外部服务):

```shell
cd backend
go test ./cmd/radius-client/ -v
```
