# Micro Net Hub UI (Vue 3)

Micro Net Hub 的管理后台前端，由旧版 Vue 2 + Element UI 前端迁移而来（旧版已移除）。

技术栈：Vue 3、Vite、TypeScript、Element Plus、Pinia、Vue Router、vue-i18n。

## 开发

```bash
npm ci
npm run dev          # http://localhost:8001，读取 .env.dev_test
npm run dev:prod     # 同上，读取 .env.dev_prod（用于连另一个后端）
```

`.env.dev_test` / `.env.dev_prod` 里唯一被用到的配置项：

- `VITE_PROXY_TARGET`：开发服务器把 `/api` 代理到的后端地址（默认 `http://localhost:9000/`）

构建不读取任何 `.env` 文件。登录、改密时加密口令用的 RSA 公钥在运行时从后端 `GET /api/base/pubkey` 获取（见 `src/utils/rsa.ts`）。

## 构建

```bash
npm run build        # 产物输出到 dist/，base 为 /ui/
npm run build:ui     # 构建并同步到 ../backend/ui，供后端 go:embed 嵌入
npm run gen:brands   # 修改 scripts/gen-brand-icons.mjs 里的品牌列表后，重新生成 src/icons/brands.ts
```

仓库根目录的 `make fe` 会执行 `npm ci && npm run build:ui`。

## 检查

```bash
npm run type-check
npm run lint
```

## 路由与菜单

基础页面在 `src/router/index.ts`；登录后根据后端菜单（`/api/menu/access/tree`）生成动态路由，菜单的 `component` 字段对应 `src/views` 下的文件路径（例如 `/personnel/user/index` → `src/views/personnel/user/index.vue`）。
