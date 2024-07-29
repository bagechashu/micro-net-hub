<template>
  <div class="changepass-container">
    <el-card shadow="always" style="margin-top: 5rem">
      <div slot="header">
        <b>{{ $t('changePassword.6111lkey47k0') }}</b>
      </div>
      <el-form ref="form" :model="form" size="medium" label-width="auto">
        <el-form-item required :label="$t('changePassword.6111lkey4lw0')">
          <div class="input-container">
            <el-input v-model="form.mail" :placeholder="$t('changePassword.6111lkey4r00')" />
            <el-button
              :loading="vCodeLoading"
              type="primary"
              @click="sendVerificationCode"
            >{{ $t('changePassword.6111lkey4tw0') }}</el-button>
          </div>
        </el-form-item>
        <el-form-item required :label="$t('changePassword.6111lkey4wg0')">
          <el-input v-model="form.code" :placeholder="$t('changePassword.6111lkey4zc0')" />
        </el-form-item>
        <el-form-item class="reset-item">
          <el-button
            :loading="resetPassLoading"
            type="primary"
            @click="resetPass"
          >{{ $t('changePassword.6111lkey47k0') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { emailPass, sendCode } from "@/api/system/user";
import { Message } from "element-ui";

export default {
  name: "ChangePass",
  data() {
    return {
      vCodeLoading: false,
      resetPassLoading: false,
      // 查询参数
      form: {
        mail: "",
        code: ""
      }
    };
  },
  methods: {
    // 发送邮箱验证码
    async sendVerificationCode() {
      this.vCodeLoading = true;
      try {
        await sendCode({ mail: this.form.mail }).then((res) => {
          if (res.code === 200) {
            Message({
              showClose: true,
              message: this.$t("changePassword.otpSendTips"),
              type: "success"
            });
            // 重新登录
            setTimeout(() => {
              this.vCodeLoading = false;
            }, 60000);
          } else {
            this.vCodeLoading = false;
          }
        });
      } catch {
        (err) => {
          this.vCodeLoading = false;
          Message({
            showClose: true,
            message: err,
            type: "error"
          });
        };
      }
    },
    // 重置密码
    async resetPass() {
      this.resetPassLoading = true;
      try {
        await emailPass(this.form).then((res) => {
          if (res.code === 200) {
            Message({
              showClose: true,
              message: this.$t("changePassword.resetSuccessTips"),
              type: "success"
            });
            // 重新登录
            setTimeout(() => {
              this.$router.replace({ path: "/" });
            }, 3000);
            this.resetPassLoading = false;
          } else {
            this.resetPassLoading = false;
          }
        });
      } catch {
        (err) => {
          this.resetPassLoading = false;
          Message({
            showClose: true,
            message: err,
            type: "error"
          });
        };
      }
    }
  }
};
</script>

<style scoped lang="scss">
.changepass-container {
  display: flex;
  justify-content: center;
}

.input-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.input-container .el-input {
  flex: 1;
  margin-right: 10px;
}

// .code-item .el-input {
//   width: 20rem;
// }

.reset-item {
  text-align: right;
}
</style>
