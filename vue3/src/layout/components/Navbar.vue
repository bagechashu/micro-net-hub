<template>
  <div class="navbar">
    <hamburger id="hamburger-container" :is-active="sidebar.opened" class="hamburger-container"
               @toggleClick="toggleSidebar" />

    <breadcrumb id="breadcrumb-container" class="breadcrumb-container" />

    <div class="right-menu">
      <template v-if="device !== 'mobile'">
        <header-search id="header-search" class="right-menu-item" />

        <error-log class="errLog-container right-menu-item hover-effect" />

        <screenfull id="screenfull" class="right-menu-item hover-effect" />

        <el-tooltip content="全局组件尺寸" effect="dark" placement="bottom">
          <size-select id="size-select" class="right-menu-item hover-effect" />
        </el-tooltip>

        <el-tooltip :content="$t('theme.tooltip')" effect="dark" placement="bottom">
          <theme-switch id="theme-switch" class="right-menu-item hover-effect" />
        </el-tooltip>

        <el-tooltip content="Language" effect="dark" placement="bottom">
          <language-select id="language-select" class="right-menu-item hover-effect" />
        </el-tooltip>

        <el-tooltip content="Doc Refs [eryajf]" effect="dark" placement="bottom">
          <a class="right-menu-item hover-effect doc-link" href="http://ldapdoc.eryajf.net" target="_blank" rel="noopener">
            <el-icon :size="20"><Document /></el-icon>
          </a>
        </el-tooltip>

      </template>

      <el-dropdown v-if="token" class="avatar-container right-menu-item hover-effect" trigger="click">
        <div class="avatar-wrapper">
          <img :src="avatar || defaultAvatar" class="user-avatar">
          <el-icon class="el-icon-caret-bottom" size="small">
            <CaretBottom />
          </el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <router-link to="/profile/index">
              <el-dropdown-item>{{ $t('profile.profile') }}</el-dropdown-item>
            </router-link>
            <el-dropdown-item divided @click="logout">
              <span style="display:block;">{{ $t('loginform.logout') }}</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <div v-else class="avatar-container right-menu-item hover-effect">
        <el-button type="primary" size="small" plain @click="updateLoginFormVisible(true)">{{ $t('loginform.login') }}</el-button>
      </div>
    </div>
    <Login v-model="loginFormVisible" />
  </div>
</template>

<script>
import appStore from '@/store/modules/app';
import userStore from '@/store/modules/user';
import Breadcrumb from '@/components/Breadcrumb';
import Hamburger from '@/components/Hamburger';
import ErrorLog from '@/components/ErrorLog';
import Screenfull from '@/components/Screenfull';
import SizeSelect from '@/components/SizeSelect';
import HeaderSearch from '@/components/HeaderSearch';
import LanguageSelect from '@/components/LanguageSelect/index.vue';
import ThemeSwitch from '@/components/ThemeSwitch/index.vue';
import { defineComponent } from 'vue';
import { CaretBottom, Document } from '@element-plus/icons-vue';
import Login from '@/layout/components/Login';

export default defineComponent({
  components: {
    Breadcrumb,
    Hamburger,
    ErrorLog,
    Screenfull,
    SizeSelect,
    HeaderSearch,
    LanguageSelect,
    ThemeSwitch,
    CaretBottom,
    Document,
    Login
  },
  data() {
    return {
      loginFormVisible: false,
      defaultAvatar: 'https://q1.qlogo.cn/g?b=qq&nk=10002&s=100'
    };
  },
  computed: {
    sidebar() {
      return appStore().sidebar;
    },
    device() {
      return appStore().device;
    },
    avatar() {
      return userStore().avatar;
    },
    token() {
      return userStore().token;
    }
  },
  methods: {
    updateLoginFormVisible(value) {
      this.loginFormVisible = value;
    },
    toggleSidebar() {
      appStore().toggleSidebar();
    },
    async logout() {
      await userStore().logout();
      this.$router.push('/');
    }
  }
});
</script>

<style lang="scss" scoped>
.doc-link {
  display: inline-flex !important;
  align-items: center;
}

.navbar {
  height: 50px;
  overflow: hidden;
  position: relative;
  background: var(--app-header-bg);
  box-shadow: var(--app-header-shadow);

  .hamburger-container {
    line-height: 46px;
    height: 100%;
    float: left;
    cursor: pointer;
    transition: background .3s;
    -webkit-tap-highlight-color: transparent;

    &:hover {
      background: var(--app-hover-bg)
    }
  }

  .breadcrumb-container {
    float: left;
  }

  .errLog-container {
    display: inline-block;
    vertical-align: top;
  }

  .right-menu {
    float: right;
    height: 100%;
    line-height: 50px;

    &:focus {
      outline: none;
    }

    .right-menu-item {
      display: inline-block;
      padding: 0 8px;
      height: 100%;
      line-height: 50px;
      font-size: 18px;
      color: var(--el-text-color-regular);
      vertical-align: text-bottom;

      &.hover-effect {
        cursor: pointer;
        transition: background .3s;

        &:hover {
          background: var(--app-hover-bg)
        }
      }
    }

    .avatar-container {
      margin-right: 30px;

      .avatar-wrapper {
        margin-top: 5px;
        position: relative;
        height: 45px;

        .user-avatar {
          cursor: pointer;
          width: 40px;
          height: 40px;
          border-radius: 10px;
        }

        .el-icon-caret-bottom {
          cursor: pointer;
          position: absolute;
          right: -20px;
          top: 25px;
          font-size: 12px;
        }
      }
    }
  }
}
</style>
