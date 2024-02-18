<template>
  <div class="changepass-container">
    <el-card shadow="always" style="margin-top: 5rem">
      <div slot="header">
        <b>重置密码</b>
      </div>
      <el-form ref="form" :model="form" size="medium" label-width="auto">
        <el-form-item required label="邮箱">
          <div class="input-container">
            <el-input v-model="form.mail" placeholder="请输入个人邮箱" />
            <el-button
              :loading="vCodeLoading"
              type="primary"
              @click="sendVerificationCode"
            >发送验证码</el-button>
          </div>
        </el-form-item>
        <el-form-item required label="验证码" class="code-item">
          <el-input v-model="form.code" placeholder="请输入验证码" />
        </el-form-item>
        <el-form-item class="reset-item">
          <el-button
            :loading="resetPassLoading"
            type="primary"
            @click="resetPass"
          >重置密码</el-button>
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
              message: "验证码已发送, 60s后再次发送",
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
              message: "操作成功, 3s后返回首页",
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

.code-item .el-input {
  width: 20rem;
}

.reset-item {
  text-align: right;
}
</style>
