// 把 vue3/dist 同步到 backend/ui，供后端通过 go:embed 嵌入。
// backend/ui 整个目录都是构建产物（web.go 来自 public/web.go，随构建一起输出）。
import { cpSync, existsSync, readdirSync, rmSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const root = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const dist = path.join(root, 'dist');
const target = path.join(root, '..', 'backend', 'ui');

if (!existsSync(path.join(dist, 'index.html')) || !existsSync(path.join(dist, 'web.go'))) {
  console.error(`[sync-ui] ${dist} 中缺少 index.html 或 web.go，请先执行构建`);
  process.exit(1);
}

for (const name of readdirSync(target)) {
  rmSync(path.join(target, name), { recursive: true, force: true });
}
cpSync(dist, target, { recursive: true });
console.log(`[sync-ui] ${dist} -> ${target}`);
