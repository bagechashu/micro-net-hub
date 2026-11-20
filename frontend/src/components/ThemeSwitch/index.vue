<template>
  <el-dropdown trigger="click" @command="setMode">
    <div class="theme-switch">
      <el-icon :size="20"><component :is="currentIcon" /></el-icon>
    </div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="item in options"
          :key="item.value"
          :command="item.value"
          :disabled="theme.mode === item.value"
          :icon="item.icon">
          {{ item.label }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Monitor, Sunny, Moon } from '@/icons/tabler';
import useThemeStore from '@/store/modules/theme';
import type { ThemeMode } from '@/store/modules/theme';

defineOptions({ name: 'ThemeSwitch' });

const { t } = useI18n();
const theme = useThemeStore();

const options = computed(() => [
  { value: 'system' as ThemeMode, label: t('theme.system'), icon: Monitor },
  { value: 'light' as ThemeMode, label: t('theme.light'), icon: Sunny },
  { value: 'dark' as ThemeMode, label: t('theme.dark'), icon: Moon }
]);

// 按钮图标反映当前选择的模式
const currentIcon = computed(() => options.value.find(o => o.value === theme.mode)?.icon || Monitor);

function setMode(mode: ThemeMode) {
  theme.setMode(mode);
}
</script>

<style scoped>
.theme-switch {
  display: flex;
  align-items: center;
  height: 100%;
}
</style>
