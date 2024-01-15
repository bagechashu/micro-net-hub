<template>
  <div class="navbar">
    <hamburger
      id="hamburger-container"
      :is-active="sidebar.opened"
      class="hamburger-container"
      @toggleClick="toggleSideBar"
    />

    <breadcrumb id="breadcrumb-container" class="breadcrumb-container" />

    <div class="right-menu">
      <template v-if="device !== 'mobile'">
        <el-tooltip content="Search" effect="dark" placement="bottom">
          <HeaderSearch id="header-search" class="right-menu-item" />
        </el-tooltip>

        <error-log class="errLog-container right-menu-item hover-effect" />

        <el-tooltip content="FullScreen" effect="dark" placement="bottom">
          <screenfull id="screenfull" class="right-menu-item hover-effect" />
        </el-tooltip>

        <el-tooltip content="Size" effect="dark" placement="bottom">
          <size-select id="size-select" class="right-menu-item hover-effect" />
        </el-tooltip>

        <el-tooltip
          content="Doc Refs [eryajf]"
          effect="dark"
          placement="bottom"
        >
          <el-link
            style="font-size: 23px"
            icon="el-icon-document"
            class="right-menu-item"
            href="http://ldapdoc.eryajf.net"
            :underline="false"
            target="_blank"
          />
        </el-tooltip>
        <el-tooltip content="GitHub" effect="dark" placement="bottom">
          <el-link
            style="font-size: 23px"
            class="iconfont icon-github right-menu-item"
            href="https://github.com/bagechashu/micro-net-hub"
            :underline="false"
            target="_blank"
          />
        </el-tooltip>
      </template>

      <el-dropdown
        class="avatar-container right-menu-item hover-effect"
        trigger="click"
      >
        <div v-if="token" class="avatar-wrapper">
          <img :src="navavatar" class="user-avatar">
          <i class="el-icon-caret-bottom" />
        </div>
        <el-dropdown-menu v-if="token" slot="dropdown">
          <router-link to="/profile/index">
            <el-dropdown-item>个人中心</el-dropdown-item>
          </router-link>
          <el-dropdown-item divided @click.native="logout">
            <span style="display: block">退出登陆</span>
          </el-dropdown-item>
        </el-dropdown-menu>
        <div v-else>
          <el-button type="primary" size="mini" plain @click="updateLoginFormVisible(true)">登录</el-button>
        </div>
      </el-dropdown>
    </div>
    <Login :login-form-visible.sync="loginFormVisible" @emitUpdateLoginFormVisible="updateLoginFormVisible" />
  </div>
</template>

<script>
import { mapGetters } from "vuex";
import Breadcrumb from "@/components/Breadcrumb";
import Hamburger from "@/components/Hamburger";
import ErrorLog from "@/components/ErrorLog";
import Screenfull from "@/components/Screenfull";
import SizeSelect from "@/components/SizeSelect";
import HeaderSearch from "@/components/HeaderSearch";
import Login from "@/components/Login";
import "@/assets/iconfont/font/iconfont.css";

export default {
  components: {
    Breadcrumb,
    Hamburger,
    ErrorLog,
    Screenfull,
    SizeSelect,
    HeaderSearch,
    Login
  },
  data() {
    return {
      navavatar: "",
      loginFormVisible: false
    };
  },
  computed: {
    ...mapGetters(["sidebar", "avatar", "device", "token"])
  },
  created() {
    this.getAvator();
  },
  methods: {
    updateLoginFormVisible(value) {
      this.loginFormVisible = value;
    },
    toggleSideBar() {
      this.$store.dispatch("app/toggleSideBar");
    },
    async logout() {
      await this.$store.dispatch("user/logout");
      // this.$router.push(`/login?redirect=${this.$route.fullPath}`);
      this.$router.push(`/`);
    },
    getAvator() {
      this.navavatar = this.avatar
        ? this.avatar
        : "https://q1.qlogo.cn/g?b=qq&nk=10002&s=100";
    }
  }
};
</script>

<style lang="scss" scoped>
.head-github {
  cursor: pointer;
  font-size: 18px;
  vertical-align: middle;
}
.navbar {
  height: 50px;
  overflow: hidden;
  position: relative;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);

  .hamburger-container {
    line-height: 46px;
    height: 100%;
    float: left;
    cursor: pointer;
    transition: background 0.3s;
    -webkit-tap-highlight-color: transparent;

    &:hover {
      background: rgba(0, 0, 0, 0.025);
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
      font-size: 18px;
      color: #5a5e66;
      vertical-align: text-bottom;

      &.hover-effect {
        cursor: pointer;
        transition: background 0.3s;

        &:hover {
          background: rgba(0, 0, 0, 0.025);
        }
      }
    }

    .avatar-container {
      margin-right: 30px;

      .avatar-wrapper {
        margin-top: 5px;
        position: relative;

        .user-avatar {
          cursor: pointer;
          width: 30px;
          height: 30px;
          border-radius: 5px;
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
