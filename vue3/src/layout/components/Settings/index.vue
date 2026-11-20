<template>
  <div class="drawer-container">
    <div>
      <h3 class="drawer-title">系统设置</h3>

      <!-- <div class="drawer-item">
        <span>Theme Color</span>
        <theme-picker style="float: right;height: 26px;margin: -3px 8px 0 0;" @change="themeChange" />
      </div> -->

      <div class="drawer-item">
        <span>开启 Tags-View</span>
        <el-switch v-model="tagsView" class="drawer-switch" />
      </div>

      <div class="drawer-item">
        <span>固定 Header</span>
        <el-switch v-model="fixedHeader" class="drawer-switch" />
      </div>

      <div class="drawer-item">
        <span>侧边栏 Logo</span>
        <el-switch v-model="sidebarLogo" class="drawer-switch" />
      </div>

      <div class="drawer-item">
        <span>侧边栏子目录气泡显示</span>
        <el-switch v-model="secondMenuPopup" class="drawer-switch" />
      </div>

    </div>
  </div>
</template>

<script>
// import ThemePicker from '@/components/ThemePicker';
import { defineComponent, computed } from 'vue';
import settingsStore from '@/store/modules/settings';

export default defineComponent({
  components: {
    // ThemePicker
  },
  setup() {
    const settings = settingsStore();

    const fixedHeader = computed({
      get: () => settings.fixedHeader,
      set: (val) => settings.changeSetting({ key: 'fixedHeader', value: val })
    });

    const tagsView = computed({
      get: () => settings.tagsView,
      set: (val) => settings.changeSetting({ key: 'tagsView', value: val })
    });

    const sidebarLogo = computed({
      get: () => settings.sidebarLogo,
      set: (val) => settings.changeSetting({ key: 'sidebarLogo', value: val })
    });

    const secondMenuPopup = computed({
      get: () => settings.secondMenuPopup,
      set: (val) => settings.changeSetting({ key: 'secondMenuPopup', value: val })
    });

    function themeChange(val) {
      settings.changeSetting({ key: 'theme', value: val });
    }

    return {
      fixedHeader,
      tagsView,
      sidebarLogo,
      secondMenuPopup,
      themeChange
    };
  }
});
</script>

<style lang="scss" scoped>
.drawer-container {
  padding: 24px;
  font-size: 14px;
  line-height: 1.5;
  word-wrap: break-word;

  .drawer-title {
    margin-bottom: 12px;
    color: var(--el-text-color-primary);
    font-size: 14px;
    line-height: 22px;
  }

  .drawer-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    color: var(--el-text-color-regular);
    font-size: 14px;
    padding: 12px 0;
  }

  .drawer-switch {
    float: right
  }
}
</style>
