// ESLint flat config（ESLint 10）
import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import pluginVue from 'eslint-plugin-vue';
import vueParser from 'vue-eslint-parser';
import prettierConfig from 'eslint-config-prettier';
import globals from 'globals';

export default tseslint.config(
  { ignores: ['node_modules/**', 'dist/**', '.vscode/**', 'auto-imports.d.ts', 'components.d.ts'] },

  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...pluginVue.configs['flat/essential'],

  {
    files: ['**/*.{js,mjs,ts,vue}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.node,
        // unplugin-auto-import 自动引入的 Element Plus 服务
        ElLoading: 'readonly',
        ElMessage: 'readonly',
        ElMessageBox: 'readonly',
        ElNotification: 'readonly'
      },
      parser: vueParser,
      parserOptions: {
        parser: tseslint.parser,
        extraFileExtensions: ['.vue']
      }
    },
    rules: {
      'vue/multi-word-component-names': 'off',
      'vue/no-v-html': 'off',
      'vue/no-multiple-template-root': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-unused-vars': ['error', { args: 'none', caughtErrors: 'none' }],
      'no-unused-vars': 'off',
      eqeqeq: ['error', 'always', { null: 'ignore' }],
      'prefer-const': 'error',
      'no-console': 'off',
      'no-debugger': 'warn'
    }
  },

  // 关闭与 Prettier 冲突的格式规则
  prettierConfig
);
