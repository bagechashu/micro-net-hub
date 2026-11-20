# 镜像只包含编译好的 linux/amd64 二进制（前端已通过 go:embed 打进二进制），需先执行 make fe be。
# 只固定小版本：重新构建时会拿到 3.24.x 的最新补丁。
FROM alpine:3.24

# 升级基础镜像里已有的包（拿到镜像发布后才修复的漏洞），并创建运行用户 app（uid/gid 1000）
RUN apk upgrade --no-cache \
    && addgroup -S -g 1000 app \
    && adduser -S -D -H -u 1000 -G app -h /app app

WORKDIR /app
COPY --chown=app:app --chmod=755 bin/micro-net-hub-linux-amd64 ./micro-net-hub
# 程序会在工作目录下写 logs/；配置文件和 RSA 密钥运行时挂载到 /app，需要对 uid 1000 可读
RUN mkdir -p /app/logs && chown -R app:app /app

USER app
EXPOSE 9000

# exec 形式：程序是 1 号进程，docker stop 的 SIGTERM 能直接送达
CMD ["./micro-net-hub"]
