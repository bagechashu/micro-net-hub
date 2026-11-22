<!-- @format -->

# Install micro-net-hub/LDAP/Ocserv by docker-compose

> centos7.9 test success
> ubuntu22.04 test failed

```shell

git clone https://github.com/bagechashu/micro-net-hub.git

cd micro-net-hub/docs/docker-compose/all-in-one
# config/app/conf/ holds config.yml, the RSA keys (generated on first start) and rotation backups;
# the container runs as uid 1000
sudo chown -R 1000:1000 config/app/conf
docker-compose up -d

# Notice: openldap will start successfully until more than 5min later. check the log to get the detail.

```

# Install micro-net-hub(from source)/LDAP/Ocserv on centos step by step

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
mkdir -p /data/conf
cp /data/micro-net-hub/backend/conf/config.yml /data/conf/
# conf/ next to the binary is the default config directory; another one can be given with -conf <dir>.
# A new key pair is generated in /data/conf/ on first start.
# In /data/conf/config.yml set the LDAP passwords in plain text, e.g.
#   admin-pass: "plain:<ldap admin password>"
#   user-init-password: "plain:<initial user password>"

# run micro-net-hub (reads /data/conf/config.yml; logs/ is written to the working directory)
chmod +x /data/micro-net-hub-linux-amd64
cd /data && ./micro-net-hub-linux-amd64

```

Administrator of Micro-Net-Hub Dashboard:

- username: admin
- password: admin_pass (default password)

## RSA key

Micro-Net-Hub uses an RSA key pair to encrypt passwords sent by the login form and to store user passwords.

- **New deployment**: the key pair is generated automatically on first start (when the key files don't exist and the database has no users yet). They are written next to `config.yml` (relative key paths are relative to the directory of `config.yml`); with docker-compose that is `config/app/conf/`. **Back up this directory**: without it, existing users can no longer log in. Make sure the directory is writable by the container user (uid 1000): `sudo chown -R 1000:1000 config/app/conf`.
- **Passwords in config.yml** (`ldap.admin-pass`, `ldap.user-init-password`) can be written in plain text with the `plain:` prefix, e.g. `admin-pass: "plain:your-ldap-admin-password"`, or as ciphertext encrypted with this deployment's public key.
- **Existing deployments still using the key from this repository** (the service logs a warning at startup): stop the service, then run the rotation command. It generates a new key pair, re-encrypts all stored user passwords and the encrypted values in `config.yml`, and backs up the old keys and config to `rsa-backup-<time>/` next to the private key (with docker-compose: `config/app/conf/rsa-backup-<time>/` on the host). Users keep their passwords.

```shell
# docker-compose (all-in-one)
docker compose stop micro-net-hub
docker compose run --rm micro-net-hub ./micro-net-hub rotate-rsa-key --dry-run   # check first
docker compose run --rm micro-net-hub ./micro-net-hub rotate-rsa-key
docker compose start micro-net-hub

# binary
./micro-net-hub-linux-amd64 rotate-rsa-key --dry-run
./micro-net-hub-linux-amd64 rotate-rsa-key
```

> Older docker-compose files mounted `/app/config.yml` and the keys (`/app/keys/` or single `auth-*.pem` files) separately. They keep working: without `/app/conf/config.yml` the service falls back to `/app/config.yml`. But with single-file mounts the rotation backup stays inside the container and is lost with it; switch to the `conf/` directory mount first:
>
> ```shell
> docker compose stop micro-net-hub
> mkdir -p config/app/conf
> mv config/app/config.yml config/app/conf/
> mv config/app/keys config/app/conf/keys   # key paths `keys/auth-*.pem` in config.yml still resolve
> # (single-file key mounts: move the .pem files into config/app/conf/ and set
> #  rsa-public-key: auth-pub.pem / rsa-private-key: auth-priv-rsa.pem)
> sudo chown -R 1000:1000 config/app/conf
> # in docker-compose.yml replace the config.yml / keys / pem mounts with
> #   - ./config/app/conf/:/app/conf/
> docker compose up -d micro-net-hub
> ```

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
