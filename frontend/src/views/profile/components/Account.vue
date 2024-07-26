<template>
  <div>
    <el-card class="profile-card">
      <div slot="header" class="clearfix">
        <span>{{ $t('custom.profile.changePassword') }}</span>
      </div>

      <el-form
        ref="dialogForm"
        size="small"
        :model="dialogFormData"
        :rules="dialogFormRules"
        :label-width="labelWidth"
      >
        <el-form-item :label="$t('custom.loginform.oldpass')" prop="oldPassword">
          <el-input
            v-model.trim="dialogFormData.oldPassword"
            autocomplete="on"
            :type="passwordTypeOld"
            :placeholder="$t('custom.loginform.oldpassTips')"
          />
          <span class="show-pwd" @click="showPwdOld">
            <svg-icon
              :icon-class="passwordTypeOld === 'password' ? 'eye' : 'eye-open'"
            />
          </span>
        </el-form-item>

        <el-form-item :label="$t('custom.loginform.newpass')" prop="newPassword">
          <el-input
            v-model.trim="dialogFormData.newPassword"
            autocomplete="on"
            :type="passwordTypeNew"
            :placeholder="$t('custom.loginform.newpassTips')"
          />
          <span class="show-pwd" @click="showPwdNew">
            <svg-icon
              :icon-class="passwordTypeNew === 'password' ? 'eye' : 'eye-open'"
            />
          </span>
        </el-form-item>

        <el-form-item :label="$t('custom.loginform.confirmNewPass')" prop="confirmPassword">
          <el-input
            v-model.trim="dialogFormData.confirmPassword"
            autocomplete="on"
            :type="passwordTypeConfirm"
            :placeholder="$t('custom.loginform.confirmNewPassTips')"
          />
          <span class="show-pwd" @click="showPwdConfirm">
            <svg-icon
              :icon-class="
                passwordTypeConfirm === 'password' ? 'eye' : 'eye-open'
              "
            />
          </span>
        </el-form-item>

        <el-form-item>
          <el-button
            :loading="submitLoading"
            type="primary"
            @click="submitForm"
          >{{ $t('custom.common.confirm') }}</el-button>
          <el-button @click="cancelForm">{{ $t('custom.common.cancel') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { changePwd } from "@/api/system/user";
import { validatePassword } from "@/utils/validate";
import store from "@/store";
import JSEncrypt from "jsencrypt";
import { Message } from "element-ui";

export default {
  data() {
    const confirmPass = (rule, value, callback) => {
      if (value) {
        if (this.dialogFormData.newPassword !== value) {
          callback(new Error(this.$i18n.t("custom.loginform.confirmNewPassErr")));
        } else {
          callback();
        }
      } else {
        callback(new Error(this.$i18n.t("custom.loginform.confirmNewPassTips")));
      }
    };
    return {
      submitLoading: false,
      dialogFormData: {
        oldPassword: "",
        newPassword: "",
        confirmPassword: ""
      },
      dialogFormRules: {
        oldPassword: [
          { required: true, message: this.$i18n.t("custom.loginform.oldpassTips"), trigger: "blur" },
          {
            min: 6,
            max: 30,
            message: "长度在 6 到 30 个字符",
            trigger: "blur"
          }
        ],
        newPassword: [
          { required: true, validator: validatePassword, trigger: "blur" }
        ],
        confirmPassword: [
          { required: true, validator: confirmPass, trigger: "blur" }
        ]
      },
      publicKey: process.env.VUE_APP_PUBLIC_KEY,
      passwordTypeOld: "password",
      passwordTypeNew: "password",
      passwordTypeConfirm: "password"
    };
  },
  computed: {
    labelWidth() {
      return this.$i18n.locale === "zh" ? "80px" : "200px";
    }
  },
  methods: {
    submitForm() {
      this.$refs["dialogForm"].validate(async(valid) => {
        if (valid) {
          this.dialogFormDataCopy = { ...this.dialogFormData };

          // 密码RSA加密处理
          const encryptor = new JSEncrypt();
          // 设置公钥
          encryptor.setPublicKey(this.publicKey);
          // 加密密码
          const oldPasswd = encryptor.encrypt(this.dialogFormData.oldPassword);
          const newPasswd = encryptor.encrypt(this.dialogFormData.newPassword);
          const confirmPasswd = encryptor.encrypt(
            this.dialogFormData.confirmPassword
          );
          this.dialogFormDataCopy.oldPassword = oldPasswd;
          this.dialogFormDataCopy.newPassword = newPasswd;
          this.dialogFormDataCopy.confirmPassword = confirmPasswd;

          this.submitLoading = true;
          const { code } = await changePwd(this.dialogFormDataCopy);
          if (code === 200 || code === 0) {
            Message({
              showClose: true,
              message: this.$i18n.t("custom.loginform.changePasswordSuccess"),
              type: "success"
            });
            this.resetForm();
            // 重新登录
            setTimeout(() => {
              store.dispatch("user/logout").then(() => {
                location.reload(); // 为了重新实例化vue-router对象 避免bug
              });
            }, 1500);
          } else {
            Message({
              showClose: true,
              message: this.$i18n.t("custom.loginform.changePasswordErr"),
              type: "error"
            });
          }
          this.submitLoading = false;
        } else {
          this.$message({
            showClose: true,
            message: this.$i18n.t("custom.loginform.changePasswordValidErr"),
            type: "warn"
          });
          return false;
        }
      });
    },
    cancelForm() {
      this.resetForm();
    },
    resetForm() {
      this.$refs["dialogForm"].resetFields();
      this.dialogFormData = {
        oldPassword: "",
        newPassword: "",
        confirmPassword: ""
      };
    },
    showPwdOld() {
      if (this.passwordTypeOld === "password") {
        this.passwordTypeOld = "";
      } else {
        this.passwordTypeOld = "password";
      }
    },
    showPwdNew() {
      if (this.passwordTypeNew === "password") {
        this.passwordTypeNew = "";
      } else {
        this.passwordTypeNew = "password";
      }
    },
    showPwdConfirm() {
      if (this.passwordTypeConfirm === "password") {
        this.passwordTypeConfirm = "";
      } else {
        this.passwordTypeConfirm = "password";
      }
    }
  }
};
</script>

<style scoped>
.profile-card {
  min-height: 18rem;
  height: 18rem;
}
.show-pwd {
  position: absolute;
  right: 10px;
  top: 3px;
  font-size: 16px;
  color: #889aa4;
  cursor: pointer;
  user-select: none;
}
</style>
