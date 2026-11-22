<!-- @format -->

# Micro-Net-Hub Backend

Base on [go-ldap-admin](https://github.com/eryajf/go-ldap-admin), but **change too much** according to my personal habits.

# Notes

##### Init SQL for site-nav

[Click to see SQL](internal/module/sitenav/model/README.md)

##### Config directory and RSA key pairs

Config directory: the service reads `config.yml` from the `conf/` directory next to the binary (`/app/conf/` in the container), or from the directory given with `-conf <dir>` (e.g. `micro-net-hub -conf /etc/micro-net-hub`, also `rotate-rsa-key -conf <dir>`). Without `-conf` it falls back to `config.yml` next to the binary, then `conf/config.yml` / `config.yml` in the working directory (for `go run`). Relative paths in `config.yml` (RSA keys, `site-logo`) are relative to the directory of `config.yml`; `logs.path` and the sqlite `database.source` stay relative to the working directory.

For local development `conf/config.yml` is used (`go run ./cmd/micro-net-hub` from `backend/`). `conf/*.pem` are not committed (`conf/.gitignore`): on a fresh database a key pair is generated there on first start. Existing deployments rotate keys with `micro-net-hub rotate-rsa-key` (see the [README](../README.md#rsa-key-and-key-rotation)).

##### Gen encrypted password string

```shell
go run cmd/gen_pass/main.go
# JrHB7jOVjZOKLa46/bv96rg80aYPRzdsxl5kQJhAVdnMH/nwsqAq696suIwE5+CbgW+W6Shec0mO4tZeojcCPRyAwdNNG9+OAMuH2R5+edfaE2OBe57S07ZBg8uJfmSjgFYxOx1FOSUtCr9bdKgjWFWTtMR714AB23TZ8unSvHY=   # admin_pass
```

or

```shell
curl "http://127.0.0.1:9000/api/base/encryptpwd?passwd=admin_pass"
```

```sql
# Update admin password after renew public key
UPDATE `users` SET password='JrHB7jOVjZOKLa46/bv96rg80aYPRzdsxl5kQJhAVdnMH/nwsqAq696suIwE5+CbgW+W6Shec0mO4tZeojcCPRyAwdNNG9+OAMuH2R5+edfaE2OBe57S07ZBg8uJfmSjgFYxOx1FOSUtCr9bdKgjWFWTtMR714AB23TZ8unSvHY=' WHERE username='admin';

INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`username`,`password`,`nickname`,`given_name`,`mail`,`job_number`,`mobile`,`avatar`,`postal_address`,`departments`,`position`,`introduction`,`status`,`creator`,`source`,`department_id`,`source_user_id`,`source_union_id`,`user_dn`,`sync_state`,`id`) VALUES ('2023-12-20 18:02:28.026','2023-12-20 18:02:28.026',NULL,'admin','JrHB7jOVjZOKLa46/bv96rg80aYPRzdsxl5kQJhAVdnMH/nwsqAq696suIwE5+CbgW+W6Shec0mO4tZeojcCPRyAwdNNG9+OAMuH2R5+edfaE2OBe57S07ZBg8uJfmSjgFYxOx1FOSUtCr9bdKgjWFWTtMR714AB23TZ8unSvHY=','Super Admin','Super Admin','admin@example.com','0000','18888888888','https://q1.qlogo.cn/g?b=qq&nk=10002&s=100','default','default','default','default',1,'System','','','','','cn=admin,dc=example,dc=com',1,1)

```
