import { fileURLToPath, URL } from 'node:url';
import path from 'node:path';
import fs from 'node:fs';

import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';

// element plus 样式自动按需导入
import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers';

import svgSpritePlugin from '@pivanov/vite-plugin-svg-sprite';

function elementPlusStyleDeps(): string[] {
  const dir = path.resolve(process.cwd(), 'node_modules/element-plus/es/components');
  const deps = ['element-plus/es'];
  if (!fs.existsSync(dir)) return deps;
  for (const name of fs.readdirSync(dir)) {
    if (fs.existsSync(path.join(dir, name, 'style/css.mjs'))) {
      deps.push(`element-plus/es/components/${name}/style/css`);
    }
  }
  return deps;
}

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    // 后端把前端页面挂在 /ui/ 下（backend/ui/web.go）；开发时也用 /ui/，
    // 这样导航图标等 /ui/assets/... 地址在开发和生产环境一致
    base: '/ui/',
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      },
      extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.vue']
    },
    optimizeDeps: {
      // 开发时预先打包 Element Plus 各组件样式，避免首次打开某个页面时触发整页重载
      include: elementPlusStyleDeps()
    },
    plugins: [
      vue(),
      AutoImport({
        resolvers: [ElementPlusResolver()]
      }),
      Components({
        resolvers: [ElementPlusResolver()]
      }),
      svgSpritePlugin({
        iconDirs: [path.resolve(process.cwd(), 'src/icons/svg')],
        symbolId: 'icon-[name]',
        inject: 'body-last'
      })
    ],
    build: {
      // 产物输出到 dist/（含 public/web.go），再由 scripts/sync-ui.mjs 同步到 backend/ui
      outDir: 'dist',
      chunkSizeWarningLimit: 1500
    },
    server: {
      host: 'localhost',
      port: 8001,
      proxy: {
        '/api': {
          target: env.VITE_PROXY_TARGET || 'http://localhost:9000/',
          changeOrigin: true
        }
      }
    }
  };
});
