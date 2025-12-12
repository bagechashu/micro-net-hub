<template>
  <el-config-provider :size="size" :z-index="zIndex" :locale="locale">
    <router-view />
  </el-config-provider>
</template>

<script>
import { defineComponent, computed } from 'vue';
import zhCn from 'element-plus/dist/locale/zh-cn.mjs';
import appStore from '@/store/modules/app';

export default defineComponent({
  data() {
    return {
      locale: zhCn,
      zIndex: 3000
    };
  },
  setup() {
    const app = appStore();
    const size = computed(() => app.size);
    return { size };
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
