<template>
  <el-card class="profile-card">
    <div slot="header" class="clearfix">
      <span>Totp QRcode</span>
    </div>

    <div class="user-profile">
      <div class="box-center">
        <!-- <div class="user-name text-center">{{ user.totp.secret }}</div> -->
        <QrCode :id="'QrCode'" :text="formattedSecret" />
      </div>
    </div>
  </el-card>
</template>

<script>
import QrCode from "@/components/Qrcode/Qrcode.vue";
export default {
  components: { QrCode },
  props: {
    user: {
      type: Object,
      default: () => {
        return {
          totp: {},
          introduction: ""
        };
      }
    }
  },
  data() {
    return {
      totp: this.user.totp
    };
  },
  computed: {
    // 使用计算属性来生成最终的secret字符串
    formattedSecret() {
      const trimmedIntro = this.formatIntroduction();
      // eg: otpauth://totp/presightdefault_pvpnuser001?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
      return `otpauth://totp/${trimmedIntro}_${this.user.name}?secret=${this.user.totp.secret}`;
    }
  },
  methods: {
    formatIntroduction() {
      let intro = this.user.introduction.replace(/\s/g, ""); // 移除所有空格和制表符
      intro = intro.substring(0, 20); // 只获取前20个字符
      return intro;
    }
  }
};
</script>

<style lang="scss" scoped>
.profile-card {
  min-height: 18rem;
  height: 18rem;
};
.box-center {
  margin: 0 auto;
  display: table;
}

.text-muted {
  color: #777;
}

.user-profile {
  .user-name {
    font-weight: bold;
  }

  .box-center {
    padding-top: 10px;
  }

  .user-role {
    padding-top: 10px;
    font-weight: 400;
    font-size: 14px;
  }

  .box-social {
    padding-top: 30px;

    .el-table {
      border-top: 1px solid #dfe6ec;
    }
  }

  .user-follow {
    padding-top: 20px;
  }
}

.user-bio {
  margin-top: 20px;
  color: #606266;

  span {
    padding-left: 4px;
  }

  .user-bio-section {
    font-size: 14px;
    padding: 15px 0;

    .user-bio-section-header {
      border-bottom: 1px solid #dfe6ec;
      padding-bottom: 10px;
      margin-bottom: 10px;
      font-weight: bold;
    }
  }
}
</style>
