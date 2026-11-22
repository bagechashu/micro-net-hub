# TODO

> Vue3 前端已合入 `main`（v0.6.0，代码在 `frontend/`），以下功能可以开始实现。旧版 Vue 2 前端已移除。

## 1. 导航页按用户组展示链接

现状：导航页（`/sitenav`，接口 `GET /api/sitenav/getnav`）对所有人公开，任何人看到的都是全部导航。

目标：

- 所有人（包括未登录用户）仍然可以打开导航页。
- **未登录用户**：只显示“公开”的导航分组/链接（大家都能看到的内网地址）。
- **已登录用户**：在公开链接之外，再显示其所在用户组有权访问的导航分组/链接。

需要做的事：

- [ ] 后端：导航分组（`NavGroup`）增加可见性设置——公开，或关联若干用户组（LDAP 分组 `groups`）。
- [ ] 后端：`/api/sitenav/getnav` 支持可选登录：无令牌时只返回公开分组；带有效令牌时返回公开分组 + 用户所在组可见的分组。注意该接口目前挂在无鉴权的路由组上，需要“可选 JWT”中间件，而不是强制鉴权。
- [ ] 后端：导航管理接口（新增/编辑分组）支持设置可见范围；初始化数据迁移时把现有分组设为“公开”，保持升级前后行为一致。
- [ ] 前端：导航管理页（`views/sitenav/manager.vue`）编辑分组时可选择“公开 / 指定用户组”。
- [ ] 前端：导航首页（`views/sitenav/index.vue`）登录/退出后重新拉取导航。
- [ ] 顺带：后端增加站点链接的协议校验（只允许 http/https），前端校验之外的兜底（安全扫描发现 `javascript:` 链接可被直接通过接口写入）。

## 2. 支持 OIDC 登录

现状：只有基于本地数据库（用户表）的用户名 + 口令登录。

目标：增加 OIDC 登录，方便用户登录后查看自己有权限的导航。

背景：OIDC 用户就是本项目所管理的 LDAP 账户，经 Authelia 认证（参考 `docs/docker-compose/openldap/`），因此 **不需要区分 OIDC 用户和本地用户**——按用户名映射到同一个本地用户即可。

需要做的事：

- [ ] 后端：配置项增加 OIDC（issuer、client_id、client_secret、redirect_uri、scopes、用户名使用的 claim，如 `preferred_username`）。
- [ ] 后端：实现授权码流程（建议带 PKCE 和 `state`/`nonce` 校验），回调后校验 ID Token，按用户名查找本地用户；找不到或已禁用的用户拒绝登录。
- [ ] 后端：登录成功后签发与现有登录相同的 JWT，复用现有的鉴权、角色与菜单逻辑。
- [ ] 前端：登录页 / 登录弹窗增加“使用 OIDC 登录”按钮，处理回调并保存令牌。
- [ ] 文档：补充 Authelia 中注册本项目为 OIDC client 的配置示例。

## 3. 安全扫描待修复项（2026-11-22）

> 来源：`sec-scan-reports/20261122-153604/report.md`（扫描目录已被 git 忽略，含漏洞细节，不要外传）。
> 仓库里的 `config.yml` 和密钥只为方便本地测试，下面的修复目标是：**生产部署不可能沿用这些默认值**。

### P0（期限 2026-11-29）

- [ ] **SEC-001 JWT 签名密钥为默认值，可伪造任意用户令牌**（`backend/conf/config.yml:78`，docs 下两套 compose 的 config.yml 同）
  - 后端启动时校验 `jwt.key`：为空、等于 `secret key` 或过短时，release 模式下拒绝启动，debug 模式下告警。
  - 部署样例改为占位值，README 写明必须修改；或首次启动自动生成随机密钥并写回/提示。
  - 顺带修复 `config.go:67` 的 `AutomaticEnv` 不生效，让密钥可以通过环境变量注入。
- [ ] **SEC-002 默认超级管理员口令可由仓库配置还原**（`backend/conf/config.yml:154`，`init_mysql_data.go:280`）
  - 初始化时 admin 口令不再直接取 `ldap.admin-pass`：改为随机生成并在首次启动日志里输出一次，或要求通过环境变量提供。
  - 增加首次登录强制改密。
- [ ] **SEC-003 all-in-one 样例把 MySQL/LDAP 以默认口令暴露到宿主所有网卡**（`docs/docker-compose/all-in-one/docker-compose.yml:11-56`，另见 mysql、openldap 两个分步样例）
  - 去掉 3306、389/636、phpldapadmin 8000 到宿主的端口映射，或只绑定 `127.0.0.1`。
  - 口令改为从 `.env` 读取的必填变量，README 增加修改默认口令的步骤。
  - 这是攻击链 CHAIN-1（批量还原全员 VPN 凭证）的最佳阻断点。

### P1（期限 2026-12-22）

- [ ] **SEC-004 OpenLDAP 同步的用户口令为空，可直接登录**（`backend/internal/module/goldap/usermgr/ldap.go:62-79`）
  - 同步时为新用户生成随机口令（或标记为“需重置”），登录和改密都拒绝空口令。
  - 这是攻击链 CHAIN-2（空口令账户经 RADIUS 爆破 TOTP 接入 VPN）的最佳阻断点。
- [x] **SEC-005 RSA 私钥提交在仓库，用户口令可整体还原**（第 1–4 步已完成；长期项未做）（`backend/conf/auth-priv-rsa.pem` 等）
  - 根因：前端在**构建时**把公钥写死（`frontend/.env.*` 里的 `VITE_PUBLIC_KEY`），镜像里的前端只认仓库这把公钥，所以每个用官方镜像的部署都被迫使用仓库里的私钥。
  - 修复：
    1. 后端新增公开接口（如 `GET /api/base/pubkey`）返回当前公钥；前端登录、改密、用户管理三处改为运行时获取，去掉 `VITE_PUBLIC_KEY`。之后换密钥不用重新编译前端。
    2. 后端启动时若密钥文件不存在，自动生成 2048 位密钥对（私钥权限 0600）；release 模式下检测到仍是仓库里的公开密钥时告警。
    3. 部署样例不再附带私钥，改为首次启动生成并持久化到挂载目录。
    4. 迁移工具：已有部署换密钥时，用旧私钥解出 `users.password` 和 config.yml 里加密的口令（`ldap.admin-pass`、`user-init-password` 等），再用新公钥重新加密；可在 `backend/cmd/gen_pass` 基础上扩展。
  - 已完成：`GET /api/base/pubkey` + 前端运行时取公钥；启动时自动生成 2048 位密钥（已有用户却缺密钥时拒绝启动）、使用仓库公开密钥时告警；部署样例改为 `keys/` 目录挂载、口令支持 `plain:` 明文写法；`micro-net-hub rotate-rsa-key [--dry-run]` 轮换密钥（重新加密库里的用户口令和 config.yml 里的加密口令，并自动备份）。
  - [ ] 线上部署执行一次 `rotate-rsa-key`，替换仓库公开密钥（见 docs/README.md 的 RSA key 一节）。
  - 长期：口令改用 bcrypt/argon2 存储（目前是可逆加密，见报告 U004-4），RSA 只用于传输；TOTP 种子加密存储。
- [ ] **SEC-006 数据库使用 root 账户和弱默认口令**（`backend/conf/config.yml:53`）
  - 部署样例改用专用库账户（只授权 `micro_net_hub` 库）和随机口令，与 SEC-003 一起改。

其余中低危问题（如 `GET /dns/getall` 未鉴权、找回密码接口无限频、内置 DNS 开放递归、ocserv-users-web 若干问题）见报告，排在这些之后处理。
