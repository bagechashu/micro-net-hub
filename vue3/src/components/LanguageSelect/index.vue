<template>
  <el-dropdown trigger="click" @command="handleSetLocale">
    <div>
      <svg-icon class-name="language-icon" icon-class="language" />
    </div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="item of localeOptions"
          :key="item.value"
          :disabled="locale === item.value"
          :command="item.value">
          {{ item.label }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import Cookies from 'js-cookie';
import { useI18n } from 'vue-i18n';

const localeOptions = [
  { label: '中文', value: 'zh' },
  { label: 'English', value: 'en' }
];

const { locale } = useI18n();

function handleSetLocale(value: string) {
  Cookies.set('locale', value);
  // 菜单标题在生成路由时已按当前语言翻译，切换语言后整页刷新最简单可靠
  location.reload();
}
</script>
