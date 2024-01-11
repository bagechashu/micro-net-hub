<template>
  <el-card style="margin-bottom:20px;">
    <div slot="header" class="clearfix">
      <span>Totp QRcode</span>
    </div>

    <div class="user-profile">
      <div class="box-center">
        <!-- <div class="user-name text-center">{{ user.totp.secret }}</div> -->
        <QrCode :id="'QrCode'" :text="secret" />
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
          totp: {}
        };
      }
    }
  },
  data() {
    return {
      totp: this.user.totp,
      // FIXME: totp 格式 otp://
      secret: this.user.totp.secret
    };
  }
};
</script>

<style lang="scss" scoped>
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
