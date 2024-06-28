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
        <el-form
          ref="resetTotpSecretForm"
          size="small"
          :model="resetTotpSecretFormData"
          :rules="resetTotpSecretFormRules"
          label-width="auto"
        >
          <el-form-item label="OTP" prop="totp">
            <el-row type="flex" justify="space-between" align="middle">
              <el-col :xs="16" :sm="14" :md="16" :lg="16" :xl="16">
                <el-input
                  v-model.trim="resetTotpSecretFormData.totp"
                  placeholder="请输入OTP"
                />
              </el-col>

              <el-col :xs="8" :sm="10" :md="8" :lg="8" :xl="8">
                <el-popconfirm
                  title="确定重置吗？"
                  @confirm="resetTotpSecret"
                >
                  <el-button
                    slot="reference"
                    class="right"
                    size="mini"
                    icon="el-icon-refresh"
                    type="danger"
                  >重置
                  </el-button>
                </el-popconfirm>
              </el-col>
            </el-row>
          </el-form-item>
        </el-form>
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
      qrCodeStr: "",
      resetTotpSecretFormData: {
        totp: ""
      },
      resetTotpSecretFormRules: {
        totp: [
          {
            required: true,
            message: "请输入 OTP",
            trigger: "blur"
          },
          {
            validator: (rule, value, callback) => {
              if (!/^\d{6}$/.test(value)) {
                return callback(new Error("OTP 为6位数字码"));
              }
              callback();
            },
            trigger: "blur"
          }
        ]
      }
    };
  },
  methods: {
    async resetTotpSecret() {
      this.$refs["resetTotpSecretForm"].validate(async(valid) => {
        if (valid) {
          const { code, data } = await resetTotpSecret(
            this.resetTotpSecretFormData
          );
          if (code === 200) {
            this.qrCodeStr = data;
            // console.log(this.qrCodeStr)
            this.showDiscribe = false;
          }
        } else {
          this.$message({
            showClose: true,
            message: "重置 TOTP 秘钥表单校验失败",
            type: "warn"
          });
          return false;
        }
      });
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
