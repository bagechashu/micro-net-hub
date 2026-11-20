<template>
  <!-- https://juejin.cn/post/7049281448305491975 -->
  <div>
    <el-dialog :visible.sync="childVisible" @open="onDialogOpen">
      <div slot="title" class="dialog-header">
        <div class="dialog-header-text">Login</div>
      </div>
      <el-form
        ref="loginForm"
        :model="loginForm"
        :rules="loginRules"
        autocomplete="on"
        label-position="left"
        label-width="auto"
        class="dialog-body"
      >
        <!-- label-position="left" label-width="auto" let label of item at same line -->
        <el-form-item prop="username">
          <svg-icon slot="label" icon-class="user" />
          <el-input
            ref="username"
            v-model="loginForm.username"
            :placeholder="$t('loginform.username')"
            name="username"
            type="text"
            tabindex="1"
            autocomplete="on"
          />
        </el-form-item>
        <el-tooltip
          v-model="capsTooltip"
          content="Caps lock is On"
          placement="top"
          manual
        >
          <el-form-item prop="password">
            <svg-icon slot="label" icon-class="password" />
            <el-input
              :key="passwordType"
              ref="password"
              v-model="loginForm.password"
              :type="passwordType"
              :placeholder="$t('loginform.password')"
              name="password"
              tabindex="2"
              autocomplete="on"
              @keyup.native="checkCapslock"
              @blur="capsTooltip = false"
              @keyup.enter.native="handleLogin"
            />
            <span class="show-pwd" @click="showPwd">
              <svg-icon
                :icon-class="passwordType === 'password' ? 'eye' : 'eye-open'"
              />
            </span>
          </el-form-item>
        </el-tooltip>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <div class="forgetpass-btn" @click="changePass">{{ $t('loginform.forgetPassword') }}</div>
        <el-button
          :loading="loading"
          type="primary"
          size="medium"
          @click.native.prevent="handleLogin"
        >
          {{ $t('loginform.login') }}
        </el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import JSEncrypt from "jsencrypt";

export default {
  name: "Login",
  props: {
    loginFormVisible: {
      type: Boolean,
      default: false
    }
  },
  data() {
    const validatePassword = (rule, value, callback) => {
      if (value.length < 6) {
        callback(new Error("The password can not be less than 6 digits"));
      } else {
        callback();
      }
    };
    return {
      loginForm: {
        username: "",
        password: ""
      },
      loginRules: {
        username: [{ required: true, trigger: "blur" }],
        password: [
          { required: true, trigger: "blur", validator: validatePassword }
        ]
      },
      passwordType: "password",
      publicKey: process.env.VUE_APP_PUBLIC_KEY,
      capsTooltip: false,
      loading: false,
      redirect: undefined,
      otherQuery: {}
    };
  },
  computed: {
    childVisible: {
      get() {
        return this.loginFormVisible;
      },
      set(v) {
        this.$emit("emitUpdateLoginFormVisible", v);
      }
    }
  },
  watch: {
    $route: {
      handler: function(route) {
        const query = route.query;
        if (query) {
          this.redirect = query.redirect;
          this.otherQuery = this.getOtherQuery(query);
        }
      },
      immediate: true
    }
  },
  methods: {
    onDialogOpen() {
      this.$nextTick(() => {
        if (this.$refs.username && this.loginForm.username === "") {
          this.$refs.username.focus();
        } else if (this.$refs.password) {
          this.$refs.password.focus();
        }
      });
    },
    checkCapslock(e) {
      const { key } = e;
      this.capsTooltip = key && key.length === 1 && key >= "A" && key <= "Z";
    },
    showPwd() {
      if (this.passwordType === "password") {
        this.passwordType = "";
      } else {
        this.passwordType = "password";
      }
      if (this.$refs.password) {
        this.$nextTick(() => {
          this.$refs.password.focus();
        });
      }
    },
    handleLogin() {
      this.$refs.loginForm.validate((valid) => {
        if (valid) {
          this.loading = true;
          // 密码RSA加密处理
          const encryptor = new JSEncrypt();
          // 设置公钥
          encryptor.setPublicKey(this.publicKey);
          // 加密密码
          const encPassword = encryptor.encrypt(this.loginForm.password);
          const encLoginForm = {
            username: this.loginForm.username,
            password: encPassword
          };
          this.$store
            .dispatch("user/login", encLoginForm)
            .then(() => {
              this.$router.push({
                // 登录成功后, 切换到个人页
                path: this.redirect || "/profile/index",
                query: this.otherQuery
              });
              this.loading = false;
              this.$emit("emitUpdateLoginFormVisible", false);
            })
            .catch(() => {
              this.loading = false;
            });
        } else {
          return false;
        }
      });
    },
    changePass() {
      // window.location.href='/changePass'
      this.$router.push({ path: "/changePass" });
    },
    getOtherQuery(query) {
      return Object.keys(query).reduce((acc, cur) => {
        if (cur !== "redirect") {
          acc[cur] = query[cur];
        }
        return acc;
      }, {});
    }
  }
};
</script>

<style lang="scss" scoped>
.dialog-header {
  padding: 0;
  margin: 0;
}
.dialog-header-text {
  font-size: large;
  font-weight: bold;
}
.dialog-body {
  padding: 0;
  margin: 0;
}
.dialog-footer {
  padding: 0, 0, 0, 0;
  display: flex;
  justify-content: space-between;
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
.forgetpass-btn {
  margin-left: 10px;
  font-size: 16px;
  color: #889aa4;
  cursor: pointer;
  user-select: none;
}
</style>
