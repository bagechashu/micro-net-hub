<template>
  <div :class="{ 'has-logo': showLogo }">
    <logo v-if="showLogo" :collapse="isCollapse" />
    <el-scrollbar wrap-class="scrollbar-wrapper">
      <el-menu class="left-menu" :default-active="activeMenu" :collapse="isCollapse"
               :unique-opened="false" :collapse-transition="false" mode="vertical">
        <sidebar-item v-for="route in permission_routes" :key="route.path" :item="route" :base-path="route.path" :is-top-route="true" />
      </el-menu>
    </el-scrollbar>
  </div>
</template>

<script>
import { defineComponent, computed } from 'vue';
import { useRoute } from 'vue-router';
import Logo from './Logo';
import SidebarItem from './SidebarItem';
import appStore from '@/store/modules/app';
import permissionStore from '@/store/modules/permission';
import settingsStore from '@/store/modules/settings';

export default defineComponent({
  name: 'Sidebar',
  components: {
    SidebarItem,
    Logo
  },
  setup() {
    const app = appStore();
    const permission = permissionStore();
    const settings = settingsStore();
    const route = useRoute();


    const sidebar = computed(() => app.sidebar);
    const permission_routes = computed(() => permission.routes);
    const secondMenuPopup = computed(() => settings.secondMenuPopup);

    const activeMenu = computed(() => {
      const meta = route.meta || {};
      if (meta.activeMenu) return meta.activeMenu;
      return route.path;
    });

    const showLogo = computed(() => settings.sidebarLogo);

    const isCollapse = computed(() => {
      if (secondMenuPopup.value) return true;
      return !sidebar.value.opened;
    });

    return {
      sidebar,
      permission_routes,
      secondMenuPopup,
      activeMenu,
      showLogo,
      isCollapse
    };
  }
});
</script>
