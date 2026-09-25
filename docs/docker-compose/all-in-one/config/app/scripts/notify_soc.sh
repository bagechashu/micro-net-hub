#!/bin/sh
# notify_soc.sh - 安全审计通知脚本示例
#
# 由 scripthook 以 `/bin/sh notify_soc.sh` 异步调用, 环境变量由调用方通过
# scripthook.WithEnv 注入, key 自动转为大写并加 SCRIPTHOOK_ 前缀:
#   SCRIPTHOOK_MESSAGE     审计通知消息正文(来自审批通过后的 sendPlain 内容)
#   SCRIPTHOOK_USERNAME    登录用户名
#   SCRIPTHOOK_SOURCE_ADDR 登录来源地址
#
# 脚本 stdout/stderr 会被 scripthook 记录到服务端日志。
# 这里仅做示例: 打印收到的上下文; 实际使用时可替换为调用 SOC 的 webhook 等。

set -eu

MESSAGE="${SCRIPTHOOK_MESSAGE:-}"
USERNAME="${SCRIPTHOOK_USERNAME:-unknown}"
SOURCE_ADDR="${SCRIPTHOOK_SOURCE_ADDR:-unknown}"

echo "[notify_soc] user=${USERNAME} source=${SOURCE_ADDR}"
if [ -n "$MESSAGE" ]; then
	echo "$MESSAGE"
fi

# 示例: 通过 webhook 转发到 SOC(取消注释并按需填写)
# curl -sS -X POST 'https://soc.example.com/webhook' \
#   -H 'Content-Type: application/json' \
#   -d "{\"user\":\"${USERNAME}\",\"source\":\"${SOURCE_ADDR}\",\"message\":\"${MESSAGE}\"}"

exit 0
