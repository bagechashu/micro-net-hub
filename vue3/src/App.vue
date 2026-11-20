<template>
  <el-config-provider :size="size" :z-index="zIndex" :locale="locale">
    <watermark :text="userName" :font-size="16" color="rgba(0, 0, 0, 0.025)" :rotate="-30">
      <router-view />
    </watermark>
  </el-config-provider>
</template>

<script>
import { defineComponent, computed } from 'vue';
import zhLocale from 'element-plus/dist/locale/zh-cn.mjs';
import enLocale from 'element-plus/dist/locale/en.mjs';
import Cookies from 'js-cookie';
import appStore from '@/store/modules/app';
import userStore from '@/store/modules/user';
import Watermark from '@/components/Watermark/index.vue';

export default defineComponent({
  components: { Watermark },
  setup() {
    const app = appStore();
    const user = userStore();
    // 登录后以用户名作为页面水印（与 Vue2 版本一致）
    const userName = computed(() => user.name || '');
    const size = computed(() => app.size);
    const locale = computed(() => (Cookies.get('locale') === 'zh' ? zhLocale : enLocale));
    return { size, locale, userName };
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
