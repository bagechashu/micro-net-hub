<!-- @format -->

# Install LDAP/Ocserv/micro-net-hub on centos step by step

> This installation document is for reference only,
> and you can deployed in the way you need according to your actual situation.

## 0. Environment preparation

**only for centos7**

```shell
# network
setenforce 0
systemctl stop firewalld

# install git, docker
yum -y install git wget

sudo yum autoremove -y docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-engine
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io
sudo systemctl start docker
sudo systemctl enable docker

# install golang (https://go.dev/doc/install)
wget https://go.dev/dl/go1.21.7.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.7.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc

source ~/.bashrc
go version

# install node.js (https://nodejs.org/en/download/)
VERSION=v14.19.1
DISTRO=linux-x64
sudo mkdir -p /usr/local/lib/nodejs
wget https://nodejs.org/dist/v20.11.0/node-$VERSION-$DISTRO.tar.xz
sudo tar -xJvf node-$VERSION-$DISTRO.tar.xz -C /usr/local/lib/nodejs

echo "VERSION=v14.19.1" >> ~/.bashrc
echo "DISTRO=linux-x64" >> ~/.bashrc
echo 'export PATH=/usr/local/lib/nodejs/node-$VERSION-$DISTRO/bin:$PATH' >> ~/.bashrc

source ~/.bashrc
node -v

# create base dir
mkdir /data && cd /data
git clone https://github.com/bagechashu/micro-net-hub.git

```

## 1. Install LDAP

```shell
cd /data
cp -a micro-net-hub/docs/docker-compose/openldap /data
cd /data/openldap
docker compose up -d

```

## 2. Install Micro-Net-Hub

```shell
# install mysql
cd /data
cp -a micro-net-hub/docs/docker-compose/mysql /data
cd /data/mysql
docker compose up -d

# build micro-net-hub
cd /data/micro-net-hub
make all

cd /data
cp /data/micro-net-hub/bin/micro-net-hub-linux-amd64 /data
cp /data/micro-net-hub/backend/config.yml /data
cp /data/micro-net-hub/backend/auth-pub.pem /data
cp /data/micro-net-hub/backend/auth-priv-rsa.pem /data

# run micro-net-hub
chmod +x /data/micro-net-hub-linux-amd64
/data/micro-net-hub-linux-amd64

```
Administrator of Micro-Net-Hub Dashboard:
- username: admin
- password: admin_pass

## 3. Install Ocserv

```shell
# install ocserv
yum -y install ocserv
sysctl -w net.ipv4.ip_forward=1
iptables -t nat -A POSTROUTING -j MASQUERADE

```

##### adapt ocserv configuration:

```
# /etc/radcli/radiusclient.conf
authserver 127.0.0.1
servers /etc/radcli/servers

# /etc/radcli/servers
127.0.0.1 default-radius-secret

# /etc/ocserv/ocserv.conf
auth = "radius [config=/usr/local/etc/radcli/radiusclient.conf,groupconfig=true]"

```

```shell
# restart ocserv service
systemctl restart ocserv
```
