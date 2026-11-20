# Micro Net Hub UI (Vue 3)

Micro Net Hub 的管理后台前端，由 `frontend/`（Vue 2 + Element UI）迁移而来。

技术栈：Vue 3、Vite、TypeScript、Element Plus、Pinia、Vue Router、vue-i18n。

## 开发

```bash
npm ci
npm run dev          # http://localhost:8001，/api 代理到 http://localhost:9000
```

可在 `.env.dev_test` 等文件中设置：

- `VITE_PUBLIC_KEY`：登录时加密口令用的 RSA 公钥（与后端 `system.rsa-public-key` 对应）
- `VITE_PROXY_TARGET`：开发代理的后端地址（默认 `http://localhost:9000/`）

## 构建

```bash
npm run build        # 产物输出到 dist/，base 为 /ui/
npm run build:ui     # 构建并同步到 ../backend/ui，供后端 go:embed 嵌入
```

仓库根目录的 `make fe` 会执行 `npm ci && npm run build:ui`。

## 检查

```bash
npm run type-check
npm run lint
```

## 路由与菜单

基础页面在 `src/router/index.ts`；登录后根据后端菜单（`/api/menu/access/tree`）生成动态路由，菜单的 `component` 字段对应 `src/views` 下的文件路径（例如 `/personnel/user/index` → `src/views/personnel/user/index.vue`）。
