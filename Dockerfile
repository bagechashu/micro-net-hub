FROM alpine:3.19

WORKDIR /app
COPY bin/micro-net-hub-linux-amd64 ./micro-net-hub
RUN chmod +x micro-net-hub

CMD ./micro-net-hub
