<template>
  <section class="app-main">
    <router-view v-slot="{ Component, route }">
      <transition name="fade-transform" mode="out-in">
        <keep-alive :include="cachedViews">
          <component :is="Component" :key="route.fullPath" />
        </keep-alive>
      </transition>
    </router-view>
    <el-footer class="footer-copyright" height="20px">
      <div>
        <span>Since 2024 </span>
        <el-divider direction="vertical" />
        <span>Powered by Open Source</span>
        <el-divider direction="vertical" />
        <span>LICENSE </span>
        <a href="https://github.com/bagechashu/micro-net-hub/blob/main/LICENSE" target="_blank" rel="noopener">GPLv3</a>
      </div>
    </el-footer>
  </section>
</template>

<script>
import { defineComponent } from 'vue';
import tagsViewStore from '@/store/modules/tagsView';

export default defineComponent({
  name: 'AppMain',
  computed: {
    cachedViews() {
      return tagsViewStore().cachedViews;
    }
  }
});
</script>

<style lang="scss" scoped>
.footer-copyright {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  font-size: xx-small;
  line-height: 10px;
  text-align: center;
}

.app-main {
  /* 50= navbar  50  */
  min-height: calc(100vh - 50px);
  width: 100%;
  position: relative;
  overflow: hidden;
}

.fixed-header+.app-main {
  padding-top: 50px;
}

.hasTagsView {
  .app-main {
    /* 84 = navbar + tags-view = 50 + 34 */
    min-height: calc(100vh - 84px);
  }

  .fixed-header+.app-main {
    padding-top: 84px;
  }
}
</style>

<style lang="scss">
// fix css style bug in open el-dialog
.el-popup-parent--hidden {
  .fixed-header {
    padding-right: 15px;
  }
}
</style>
