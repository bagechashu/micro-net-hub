# https://github.com/eryajf/go-ldap-admin/blob/main/docs/docker-compose/docker-compose.yaml

```
  openldap:
    image: registry.cn-hangzhou.aliyuncs.com/ali_eryajf/openldap:1.4.0
    container_name: go-ldap-admin-openldap
    hostname: go-ldap-admin-openldap
    restart: always
    environment:
      TZ: Asia/Shanghai
      LDAP_ORGANISATION: "eryajf.net"
      LDAP_DOMAIN: "eryajf.net"
      LDAP_ADMIN_PASSWORD: "123456"
    command: [ '--copy-service' ]
    volumes:
      - ./data/openldap/database:/var/lib/ldap
      - ./data/openldap/config:/etc/ldap/slapd.d
      - ./config/init.ldif:/container/service/slapd/assets/config/bootstrap/ldif/custom/init.ldif
    ports:
      - 388:389
    networks:
      - go-ldap-admin

```

# https://github.com/osixia/docker-openldap
#### Edit your server configuration

Do not edit slapd.conf it's not used. To modify your server configuration use ldap utils: **ldapmodify / ldapadd / ldapdelete**

#### Seed ldap database with ldif

This image can load ldif files at startup with either `ldapadd` or `ldapmodify`.
Mount `.ldif` in `/container/service/slapd/assets/config/bootstrap/ldif` directory if you want to overwrite image default bootstrap ldif files or in `/container/service/slapd/assets/config/bootstrap/ldif/custom` (recommended) to extend image config.

Files containing `changeType:` attributes will be loaded with `ldapmodify`.

The startup script provides some substitutions in bootstrap ldif files. Following substitutions are supported:

- `{{ LDAP_BASE_DN }}`
- `{{ LDAP_BACKEND }}`
- `{{ LDAP_DOMAIN }}`
- `{{ LDAP_READONLY_USER_USERNAME }}`
- `{{ LDAP_READONLY_USER_PASSWORD_ENCRYPTED }}`

Other `{{ * }}` substitutions are left unchanged.

Since startup script modifies `ldif` files, you **must** add `--copy-service`
argument to entrypoint if you don't want to overwrite them.

```sh
# single file example:
docker run \
	--volume ./bootstrap.ldif:/container/service/slapd/assets/config/bootstrap/ldif/50-bootstrap.ldif \
	osixia/openldap:1.5.0 --copy-service

# directory example:
docker run \
	--volume ./ldif:/container/service/slapd/assets/config/bootstrap/ldif/custom \
	osixia/openldap:1.5.0 --copy-service
```
