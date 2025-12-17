<template>
  <el-config-provider :size="size" :z-index="zIndex" :locale="locale">
    <router-view />
  </el-config-provider>
</template>

<script>
import { defineComponent, computed } from 'vue';
import zhLocale from 'element-plus/dist/locale/zh-cn.mjs';
import enLocale from 'element-plus/dist/locale/en.mjs';
import Cookies from 'js-cookie';
import appStore from '@/store/modules/app';

export default defineComponent({
  setup() {
    const app = appStore();
    const size = computed(() => app.size);
    const locale = computed(() => (Cookies.get('locale') === 'zh' ? zhLocale : enLocale));
    return { size, locale };
  },
  data() {
    return {
      zIndex: 3000
    };
  }
});
</script>

<style lang="scss">
html,
body,
#app {
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 0;
}
</style>
