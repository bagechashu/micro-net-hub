# Vue2 → Vue3 迁移计划

## 步骤 1.1：依赖和工具链分析

### 1. 核心框架对比

| 依赖项 | Frontend (Vue2) | Vue3 | 迁移说明 |
|-------|-----------------|------|---------|
| **Vue** | 2.6.10 | 3.5.18 | 主版本升级，语法改变 |
| **Vue Router** | 3.0.2 | 4.5.1 | 主版本升级，API 变化 |
| **Vuex/Pinia** | 3.1.0 (Vuex) | 2.3.1 (Pinia) | 状态管理工具完全替换 |
| **Vue I18n** | 8.28.2 | ❌ 未在 vue3 package.json 中 | **需要安装 Vue I18n 9.x** |

### 2. 核心依赖分析

#### 🟢 无需更改（已兼容）
- **axios**: frontend 0.18.1 → vue3 1.11.0（推荐升级，保持兼容性）
- **file-saver**: frontend 2.0.1 → vue3 2.0.5（完全兼容）
- **fuse.js**: frontend 3.4.4 → vue3 6.6.2（功能向后兼容）
- **js-cookie**: frontend 2.2.0 → vue3 3.0.5（API 兼容）
- **jszip**: frontend 3.2.1 → vue3 3.10.1（完全兼容）
- **nprogress**: frontend 0.2.0 → vue3 0.2.0（完全兼容）
- **path-to-regexp**: frontend 2.4.0 → vue3 8.2.0（API 兼容性需要检查路由表）
- **xlsx**: frontend 0.14.1 → vue3 0.18.5（完全兼容）

#### 🟡 需要升级或替换
| 依赖项 | Frontend | Vue3 | 替换方案 |
|--------|----------|------|---------|
| **element-ui** | 2.15.14 | ❌ | **替换为 element-plus 2.10.5** ✅ 已在 vue3 中 |
| **vue-clipboard2** | 0.3.3 | ❌ | **替换为 clipboard 2.0.11** ✅ 已在 vue3 中 |
| **vue-lazyload** | 1.3.5 | ❌ | **选项1**: 继续使用（v10+支持Vue3）；**选项2**: 用 native lazy loading |
| **echarts** | 4.2.1 | 5.6.0 | 升级后 API 兼容，但新版本功能更丰富 |
| **tui-editor** | 1.3.3 | ❌ | **需要替换或升级到 @toast-ui/vue-editor** |
| **codemirror** | 5.45.0 | ❌ | **需要考虑替换为 codemirror 6.x（如使用代码编辑）** |
| **vue-count-to** | 1.0.13 | ❌ | **需要找 Vue3 替代方案（如 animejs）** |
| **vue-marquee-text-component** | 1.2.0 | ❌ | **需要找 Vue3 替代方案或自己实现** |
| **screenfull** | 4.2.0 | ❌ | 用户库，继续使用（检查兼容性） |

#### ❌ 不需要迁移（开发依赖）
这些是 Vue 2 CLI 特定的，Vue 3 已用 Vite + TypeScript 替代：
- `vue-cli-service` → **Vite** 替代 ✅ 已配置
- `@vue/cli-plugin-*` → 不需要
- `vue-template-compiler` → **@vitejs/plugin-vue** 替代 ✅ 已配置
- `babel-*` → **Vite 内置支持** ✅

#### ✅ Vue3 中已有的开发工具
- **TypeScript**: 5.8.3（官方推荐）
- **ESLint**: 9.10.0（已配置）
- **Vite**: 6.3.5（构建工具）
- **Pinia**: 2.3.1（状态管理）
- **Sass**: 1.89.2（样式处理）

### 3. 缺失的关键依赖

#### 必须安装
1. **Vue I18n 9.x** - 国际化支持（frontend 使用了）
2. **jsencrypt 3.0.1** - RSA 加密（frontend 有，vue3 需要添加）

#### 可选升级
1. **dropzone** (vue3 有) vs frontend 中没有 - 文件上传
2. **sortablejs** (vue3 有) - 拖拽排序
3. **driver.js** (vue3 有) - 功能引导

### 4. 迁移策略总结

#### 🔴 立即处理
1. **安装 Vue I18n 9.x** 到 vue3 项目
2. **安装 jsencrypt 3.0.1** 到 vue3 项目
3. **替换 element-ui 为 element-plus**（已完成，vue3 中已有）
4. **替换 vue-clipboard2 为 clipboard**（已完成，vue3 中已有）

#### 🟡 迁移过程中处理
1. **Vuex 迁移到 Pinia** - 语法改写
2. **Axios 版本升级** - 更新 interceptor 写法（如有）
3. **路由表升级** - path-to-regexp 版本变化
4. **组件库 API 变化** - Element UI → Element Plus

#### 🟢 保持或跳过
1. **tui-editor, codemirror, vue-marquee-text-component** - 如果前端代码中未使用这些，可以先跳过
2. **vue-count-to** - 可以用更现代的 JavaScript 动画库替代

### 5. 推荐的安装操作

在 vue3 目录执行：

```bash
# 安装缺失的关键依赖
npm install vue-i18n@^9.x
npm install jsencrypt@^3.0.1

# 可选：升级 axios 到与 frontend 更接近的版本
npm install axios@^0.27.x  # 或保持现在的 1.11.0

# 如果需要 tui-editor
npm install @toast-ui/vue-editor

# 如果需要 codemirror 6
npm install codemirror @codemirror/lang-javascript
```

### 6. 文件配置变化

| 配置项 | Frontend | Vue3 | 备注 |
|-------|----------|------|-----|
| **构建工具** | Vue CLI | Vite | 完全不同 |
| **主配置文件** | vue.config.js | vite.config.ts | ✅ 已有 vite.config.ts |
| **入口文件** | src/main.js | src/main.ts | ✅ 已有 main.ts |
| **tsconfig** | 无 | tsconfig.json | ✅ 已有 |
| **ESLint** | .eslintrc.js | eslint.config.js | ✅ 已有 |
| **代码生成** | plop | ❌ 未在 vue3 中 | 可选保留或迁移 plopfile.js |

---

## 后续步骤建议

✅ **完成度**: 步骤 1.1 的依赖分析完成

**下一步**: 执行步骤 1.2 - 理解现有实现结构，或开始安装缺失的依赖

## Vue2 → Vue3 迁移计划

### **第一阶段：准备工作**

#### 步骤 1.1：分析依赖和工具链
- 查看 frontend 中所有依赖包和 vue3 中的对应版本
- 确认所有第三方库在 Vue3 中是否有兼容版本
- 对比 frontend 中的自定义工具函数和 vue3 中已有的实现

#### 步骤 1.2：理解现有实现结构
- 整理 frontend 中的 API 调用模式和请求拦截配置
- 了解 frontend 中 Vuex store 的模块结构和通信方式
- 审查 layout 组件中的嵌套和依赖关系
- 评估 views 中的页面依赖和复用逻辑

---

### **第二阶段：状态管理迁移（Vuex → Pinia）**

#### 步骤 2.1：迁移 Store 基础结构
- 确认 vue3 中的 Pinia 初始化配置
- 理解 frontend 中的 getters.js 和模块化结构
- 备份现有的 vue3 store 模块

#### 步骤 2.2：逐个迁移 Store 模块
- **步骤 2.2.1**：迁移 app.js → app.ts（应用状态）
- **步骤 2.2.2**：迁移 `user.js` → `user.ts`（用户认证）
- **步骤 2.2.3**：迁移 `permission.js` → `permission.ts`（权限管理）
- **步骤 2.2.4**：迁移 `tagsView.js` → `tagsView.ts`（标签页）
- **步骤 2.2.5**：迁移 `settings.js` → `settings.ts`（系统设置）
- **步骤 2.2.6**：迁移 `errorLog.js` → `errorLog.ts`（错误日志）

#### 步骤 2.3：更新 Store 调用语法
- 将所有 `this.$store.commit/dispatch` 改为 Pinia 的 action/setter
- 更新 computed 中的 `mapState/mapGetters` 用法

---

### **第三阶段：API 层迁移**

#### 步骤 3.1：分析并设计 API 架构
- 审查 frontend 的 `src/api/` 目录结构
- 确认 axios 实例配置和请求/响应拦截器
- 理解认证、错误处理等机制

I'm going to read key API files in api to map required migrations and priorities. I'll fetch base.js, user.js, menu.js, dns.js, and dashboard.js.

Read [](file:///Users/op/PersonalSpace/0_project/micro-net-hub/frontend/src/api/system/base.js)

Read [](file:///Users/op/PersonalSpace/0_project/micro-net-hub/frontend/src/api/system/user.js)

Read [](file:///Users/op/PersonalSpace/0_project/micro-net-hub/frontend/src/api/system/menu.js)

Read [](file:///Users/op/PersonalSpace/0_project/micro-net-hub/frontend/src/api/dns/dns.js)

Read [](file:///Users/op/PersonalSpace/0_project/micro-net-hub/frontend/src/api/dashboards/dashboard.js)

Updating the TODO list to mark API file listing complete and add an API migration planning item.

Completed: *列出 API 模块* (4/14)

已更新 TODO：标记 API 列表完成并开始“API 迁移优先级与映射”。

迁移建议（API 映射与优先级）：

- 优先级 高（必须先迁移，直接影响登录、路由与权限）
  - base.js → 在 api 新增 `base.ts`（或 `system/base.ts`）  
    - 原因：`login` / `logout` / `refreshToken` 被 `user` store 使用。
  - user.js → 迁移为 user.ts（已有 user.js，需合并/替换）  
    - 原因：`getInfo`、用户 CRUD 被 `user` store 与页面依赖。
  - menu.js → 迁移为 `vue3/src/api/menu.ts`（用于权限菜单与路由生成）
    - 原因：`permission` store 需要 `getUserMenuTreeByUserId`。

- 优先级 中（业务页面常用）
  - dashboard.js → `vue3/src/api/dashboard.ts`
  - dns.js → `vue3/src/api/dns.ts`
  - notice.js → `vue3/src/api/notice.ts`
  - sitenav.js → `vue3/src/api/sitenav.ts`

- 优先级 低（次级功能，可接着迁移）
  - `frontend/src/api/personnel/*` (user.js, group.js, fieldRelation.js) → 合并/分别迁移到 personnel 下
  - operationLog.js → `vue3/src/api/log/operationLog.ts`（或合并到 `log` 目录）
  - role.js / api.js → `vue3/src/api/system/role.ts` 等

迁移注意事项：
- 统一 `request` 实例路径：frontend 使用 `@/utils/request`，vue3 也应放在相同位置（或合并已有实现），并确保 Axios 版本与拦截器行为一致（token header、错误格式）。
- 类型：把新的 API 文件改为 `.ts` 并添加返回类型接口（可逐步完善）。
- 合并冲突：user.js 已存在，需决定是替换（推荐：用 frontend 的更完整实现替换/合并）还是合并两者功能。
- 路由 / permission：`menu` API 返回结构需转换为 `RouteRecordRaw` 兼容格式或在 permission store 中转换。

下一步我可以：
- 自动创建 api 下的占位 `.ts` 文件并将 frontend 的实现复制到相应文件（并把 `request` 引入改为项目现有路径），或
- 先生成每个 API 文件的迁移模板和示例类型，逐个迁移并运行测试。

要我先为上面“高优先级”的三个 API（`base`, `user`, `menu`）生成迁移模板并创建对应文件吗？

#### 步骤 3.2：迁移 API 请求配置
- 将 axios 实例配置从 frontend 迁移到 vue3（如有差异）
- 迁移请求拦截器（token 处理、error handling）
- 迁移响应拦截器

#### 步骤 3.3：逐个迁移 API 模块
- **步骤 3.3.1**：迁移 `dashboards.js` → `dashboards.ts`
- **步骤 3.3.2**：迁移 `dns.js` → `dns.ts`
- **步骤 3.3.3**：迁移 `log.js` → `log.ts`
- **步骤 3.3.4**：迁移 `notice.js` → `notice.ts`
- **步骤 3.3.5**：迁移 `personnel.js` → `personnel.ts`
- **步骤 3.3.6**：迁移 `sitenav.js` → `sitenav.ts`
- **步骤 3.3.7**：迁移 `system.js` → `system.ts`

#### 步骤 3.4：更新 API 调用方式
- 使用 TypeScript 类型定义请求和响应
- 确保所有 API 调用都有正确的返回类型

---

### **第四阶段：Layout（布局）迁移**

#### 步骤 4.1：分析 Layout 结构
- 查看 frontend layout 中的子组件
- 了解父子通信和状态管理的调用方式

#### 步骤 4.2：迁移主 Layout 组件
- 将 index.vue 转换为 Vue3 语法
- 更新组件引入方式（动态导入）
- 更新 store 调用为 Pinia

#### 步骤 4.3：迁移 Layout 子组件
- **步骤 4.3.1**：迁移 Layout 中的各子组件（Navbar, Sidebar 等）
- **步骤 4.3.2**：更新组件间通信（props, emits）
- **步骤 4.3.3**：迁移 mixins 为 Composition API（可选）

#### 步骤 4.4：测试 Layout 功能
- 测试侧边栏切换、响应式设计
- 测试权限控制是否正常工作

---

### **第五阶段：Views（页面）迁移**

#### 步骤 5.1：分析 Views 目录
- 统计所有 view 页面数量和依赖关系
- 按业务模块分组

#### 步骤 5.2：迁移基础错误页面
- 迁移 `error-page/` 下的 404、500 等错误页面
- 迁移 `login/` 登录页面
- 迁移 `redirect/` 重定向页面

#### 步骤 5.3：迁移业务页面（按模块）
- **步骤 5.3.1**：迁移 `dashboard/` 模块
- **步骤 5.3.2**：迁移 `personnel/` 人员管理模块
- **步骤 5.3.3**：迁移 `system/` 系统管理模块
- **步骤 5.3.4**：迁移 `dns/` DNS 模块
- **步骤 5.3.5**：迁移 `log/` 日志模块
- **步骤 5.3.6**：迁移 `notice/` 通知模块
- **步骤 5.3.7**：迁移 `sitenav/` 站点导航模块
- **步骤 5.3.8**：迁移其他页面（profile, changePassword 等）

#### 步骤 5.4：更新页面内 API 调用
- 替换所有 API 调用为新的模块
- 更新 store 调用为 Pinia

#### 步骤 5.5：更新页面内组件
- 迁移自定义组件引入
- 更新 directive 用法

---

### **第六阶段：路由配置迁移**

#### 步骤 6.1：分析现有路由
- 查看 frontend 的路由守卫逻辑
- 了解权限控制的实现方式

#### 步骤 6.2：迁移路由表
- 将 frontend 中的所有路由配置迁移到 vue3 的路由模块
- 确保所有路由都指向正确的组件

#### 步骤 6.3：迁移路由守卫
- 迁移 `permission.js` 中的路由守卫逻辑
- 更新 store 调用为 Pinia

#### 步骤 6.4：测试路由功能
- 测试路由导航
- 测试权限拦截

---

### **第七阶段：清理和优化**

#### 步骤 7.1：检查 TypeScript 类型
- 为所有迁移的代码添加类型定义
- 确保没有 `any` 类型

#### 步骤 7.2：代码审查和测试
- 全面测试所有功能
- 检查是否有遗漏的文件或模块

#### 步骤 7.3：清理旧文件
- 备份 frontend 目录
- 更新项目文档

---

### **关键要点总结**

| 方面 | Vue2 (frontend) | Vue3 (vue3) | 迁移策略 |
|------|---|---|---|
| **状态管理** | Vuex | Pinia | 逐模块转换，语法改写 |
| **API** | axios (callback) | axios (Promise) | 复用 interceptors，更新类型 |
| **Layout** | Options API | Vue3 SFC | 组件语法转换 |
| **Views** | Options API + mixin | Vue3 SFC + Composition | 语法转换，props/emits 改写 |
| **路由** | Vue Router 3 | Vue Router 4 | 路由表迁移，守卫更新 |
| **I18n** | vue-i18n 8 | vue-i18n 9+ (需要检查) | 如需要则更新 |

这个计划将确保系统化、有序的迁移，降低出错风险。建议按顺序完成，每个步骤完成后都要测试验证。