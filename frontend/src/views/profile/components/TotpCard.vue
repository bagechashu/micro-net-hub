<template>
  <el-card class="profile-card">
    <div slot="header" class="clearfix">
      <span>重置 Totp 秘钥</span>
    </div>
    <div class="box-center">
      <div :style="{ display: showDiscribe ? 'block' : 'none' }">
        <div class="warning-message">
          <p><b>注意:</b></p>
          <p>重置后, 之前的TOTP秘钥将失效.</p>
          <p>二维码只显示一次, 刷新/切换页面后将消失.</p>
        </div>
        <el-popconfirm title="确定重置吗？" @onConfirm="resetTotpSecret">
          <el-button
            slot="reference"
            class="right"
            size="mini"
            icon="el-icon-refresh"
            type="danger"
          >重置
          </el-button>
        </el-popconfirm>
      </div>
      <!-- <div class="user-name">{{ qrCodeStr }}</div> -->
      <QrCode :id="'QrCode'" class="mt-30" :text="qrCodeStr" />
    </div>
  </el-card>
</template>

<script>
import QrCode from "@/components/Qrcode/Qrcode.vue";
import { resetTotpSecret } from "@/api/personnel/user";

export default {
  components: { QrCode },
  data() {
    return {
      showDiscribe: true,
      qrCodeStr: ""
    };
  },
  methods: {
    async resetTotpSecret() {
      try {
        const { data } = await resetTotpSecret();
        this.qrCodeStr = data;
        // console.log(this.qrCodeStr)
        this.showDiscribe = false;
      } catch (e) {
        console.log(e);
      }
    }
  }
};
</script>

<style lang="scss" scoped>
.profile-card {
  min-height: 18rem;
  height: 18rem;
}
.box-center {
  margin: auto;
  display: table;
}
.right {
  float: right;
}
.mt-30 {
  margin-top: 30px;
}
.warning-message {
  background-color: #fef0f0;
  border: 1px solid #f86c6b;
  border-radius: 5px;
  padding: 10px;
  margin: 0 0 20px 0;
  font-size: smaller;
  p {
    margin: 10px 0;
  }

  b {
    color: #f86c6b;
  }
}
</style>
