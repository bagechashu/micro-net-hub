FROM golang:1.21  AS builder

# ENV GOPROXY      https://goproxy.io
RUN mkdir /app && apk add --no-cache --virtual .build-deps \
    ca-certificates \
    gcc \
    g++

ADD . /app/
WORKDIR /app
RUN sed -i 's@localhost:389@openldap:389@g' /app/config.yml && \
    sed -i 's@host: localhost@host: mysql@g'  /app/config.yml && \
    go build -o micro-net-hub ./cmd/micro-net-hub/main.go

### build final image
FROM alpine:3.16

# we set the timezone `Asia/Shanghai` by default, you can be modified
# by `docker build --build-arg="TZ=Other_Timezone ..."`

# ARG TZ="Asia/Shanghai"
# ENV TZ ${TZ}
# RUN ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime && \
#     echo ${TZ} > /etc/timezone

RUN mkdir /app
WORKDIR /app

COPY --from=builder /app/micro-net-hub .
COPY config.yml /app/config.yml
COPY auth-pub.pem /app/auth-pub.pem
COPY auth-priv-rsa.pem /app/auth-priv-rsa.pem

RUN chmod +x micro-net-hub

CMD ./micro-net-hub