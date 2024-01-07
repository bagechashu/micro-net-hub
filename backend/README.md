<!-- @format -->

# Micro-Net-Hub Backend

Base on [go-ldap-admin](https://github.com/eryajf/go-ldap-admin), but <span style="color:red; font-weight:bold">change too much</span> according to my personal habits.

# Notes

##### RSA key pairs

> http://ldapdoc.eryajf.net/pages/119ea3/

Empty database after use new RSA key pairs.

```shell
# Please ensure that the generated RSA key does not exceed 1024 bits in size, as the password is too long to be inserted into the database.
$ openssl genrsa -out auth-priv.pem 1024

# Openssl3 convert to traditional format
$ openssl rsa -in auth-priv.pem -out auth-priv-rsa.pem -traditional

# generate public key
$ openssl rsa -in auth-priv.pem -pubout -out auth-pub.pem
# YYHkG5kAFOyvJmmyVEt6rdDWM5q46F+Rkh3oxosFXHA86JMk4QhJQeVBknMNwvyFWLoGme2gF4eIp2WhpLUj9kxDQKrLj7AwnhILJrFmcykPPXgfBpVGA5aPrtrlucHuIsCBgyrSavHLhnKjdE0O5SbtamiVgfC+PBABY19vX2s=  # admin_pass

```

##### Gen encrypted password string

```shell
go run cmd/gen_pass/main.go
# JrHB7jOVjZOKLa46/bv96rg80aYPRzdsxl5kQJhAVdnMH/nwsqAq696suIwE5+CbgW+W6Shec0mO4tZeojcCPRyAwdNNG9+OAMuH2R5+edfaE2OBe57S07ZBg8uJfmSjgFYxOx1FOSUtCr9bdKgjWFWTtMR714AB23TZ8unSvHY=   # admin_pass
```

```sql
# Update admin password after renew public key
UPDATE `users` SET password='JrHB7jOVjZOKLa46/bv96rg80aYPRzdsxl5kQJhAVdnMH/nwsqAq696suIwE5+CbgW+W6Shec0mO4tZeojcCPRyAwdNNG9+OAMuH2R5+edfaE2OBe57S07ZBg8uJfmSjgFYxOx1FOSUtCr9bdKgjWFWTtMR714AB23TZ8unSvHY=' WHERE username='admin';

INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`username`,`password`,`nickname`,`given_name`,`mail`,`job_number`,`mobile`,`avatar`,`postal_address`,`departments`,`position`,`introduction`,`status`,`creator`,`source`,`department_id`,`source_user_id`,`source_union_id`,`user_dn`,`sync_state`,`id`) VALUES ('2023-12-20 18:02:28.026','2023-12-20 18:02:28.026',NULL,'admin','JrHB7jOVjZOKLa46/bv96rg80aYPRzdsxl5kQJhAVdnMH/nwsqAq696suIwE5+CbgW+W6Shec0mO4tZeojcCPRyAwdNNG9+OAMuH2R5+edfaE2OBe57S07ZBg8uJfmSjgFYxOx1FOSUtCr9bdKgjWFWTtMR714AB23TZ8unSvHY=','管理员','最强后台','admin@eryajf.net','0000','18888888888','https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif','地球','研发中心','打工人','最强后台的管理员',1,'系统','','','','','cn=admin,dc=example,dc=com',1,1)

```
