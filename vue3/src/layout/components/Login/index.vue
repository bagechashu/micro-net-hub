<!-- @format -->

<template>
  <div>
    <el-dialog v-model="childVisible" @open="onDialogOpen">
      <template #header>
        <div class="dialog-header">
          <div class="dialog-header-text">{{ $t('loginform.login') }}</div>
        </div>
      </template>

      <el-form ref="loginForm" :model="loginForm" :rules="loginRules" autocomplete="on" label-position="left" label-width="auto" class="dialog-body">
        <el-form-item prop="username">
          <template #label><IconUser :size="18" stroke="1.75" /></template>
          <el-input
            ref="username"
            v-model="loginForm.username"
            :placeholder="$t('loginform.username')"
            name="username"
            type="text"
            tabindex="1"
            autocomplete="on" />
        </el-form-item>

        <!-- 由 capsTooltip 控制显示（Element Plus 用 :visible；Element UI 的 v-model + manual 在这里无效，会退化成悬停即显示） -->
        <el-tooltip :visible="capsTooltip" :content="$t('loginform.capsLockOn')" placement="top">
          <el-form-item prop="password">
            <template #label>
              <IconLock :size="18" stroke="1.75" />
            </template>
            <el-input
              :key="passwordType"
              ref="password"
              v-model="loginForm.password"
              :type="passwordType"
              :placeholder="$t('loginform.password')"
              name="password"
              tabindex="2"
              autocomplete="on"
              @keydown="checkCapslock"
              @keyup="checkCapslock"
              @blur="capsTooltip = false"
              @keyup.enter="handleLogin" />
            <span class="show-pwd" @click="showPwd">
              <component :is="passwordType === 'password' ? 'IconEyeOff' : 'IconEye'" :size="18" stroke="1.75" />
            </span>
          </el-form-item>
        </el-tooltip>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <div class="forgetpass-btn" @click="changePass">{{ $t('loginform.forgetPassword') }}</div>
          <el-button :loading="loading" type="primary" size="default" @click.prevent="handleLogin">{{ $t('loginform.login') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { defineComponent } from 'vue';
import JSEncrypt from 'jsencrypt';
import userStore from '@/store/modules/user';
import { IconUser, IconLock, IconEye, IconEyeOff } from '@/icons/tabler';

export default defineComponent({
  name: 'NavbarLogin',
  components: { IconUser, IconLock, IconEye, IconEyeOff },
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue'],
  data() {
    const validatePassword = (rule, value, callback) => {
      if (!value || value.length < 6) {
        callback(new Error('The password can not be less than 6 digits'));
      } else {
        callback();
      }
    };
    return {
      loginForm: {
        username: '',
        password: '',
      },
      loginRules: {
        username: [{ required: true, trigger: 'blur' }],
        password: [{ required: true, trigger: 'blur', validator: validatePassword }],
      },
      passwordType: 'password',
      // use Vite public key env var if provided
      publicKey: import.meta.env.VITE_PUBLIC_KEY,
      capsTooltip: false,
      loading: false,
      redirect: undefined,
      otherQuery: {},
    };
  },
  computed: {
    childVisible: {
      get() {
        return this.modelValue;
      },
      set(v) {
        this.$emit('update:modelValue', v);
      },
    },
  },
  watch: {
    $route: {
      handler(route) {
        const query = route.query;
        if (query) {
          this.redirect = query.redirect;
          this.otherQuery = this.getOtherQuery(query);
        }
      },
      immediate: true,
    },
  },
  methods: {
    onDialogOpen() {
      this.$nextTick(() => {
        if (this.$refs.username && this.loginForm.username === '') {
          // @ts-expect-error - template ref typing
          this.$refs.username.focus();
        } else if (this.$refs.password) {
          // @ts-expect-error - template ref typing
          this.$refs.password.focus();
        }
      });
    },
    // 直接读取大写锁定状态：不受 Shift、数字键影响，按下 CapsLock 键切换时也能立即更新
    checkCapslock(e) {
      if (e && typeof e.getModifierState === 'function') {
        this.capsTooltip = e.getModifierState('CapsLock');
      }
    },
    showPwd() {
      if (this.passwordType === 'password') {
        this.passwordType = '';
      } else {
        this.passwordType = 'password';
      }
      this.$nextTick(() => {
        // @ts-expect-error - template ref typing
        if (this.$refs.password) this.$refs.password.focus();
      });
    },
    handleLogin() {
      // use Element Plus form validate via ref API
      // but keep compatibility with Options API validate callback style for simplicity
      // @ts-expect-error - template ref typing
      if (!this.$refs.loginForm) return;
      // call inner validate
      // @ts-expect-error - template ref typing
      this.$refs.loginForm.validate(async (valid) => {
        if (valid) {
          this.loading = true;
          let payload = { ...this.loginForm };
          try {
            if (!this.publicKey) {
              console.warn('VITE_PUBLIC_KEY is empty — password will be sent plaintext');
            }
            const encryptor = new JSEncrypt();
            if (this.publicKey) encryptor.setPublicKey(this.publicKey);
            const encPassword = this.publicKey ? encryptor.encrypt(this.loginForm.password) : null;
            if (encPassword) {
              payload = { username: this.loginForm.username, password: encPassword };
            }
          } catch (err) {
            // fallback to plain password
            console.warn('Password encryption failed, sending plain password', err);
          }

          userStore()
            .login(payload)
            .then(async () => {
              try {
                // 拉取用户信息并按菜单生成动态路由（只调 getInfo 会让路由守卫误以为菜单已加载）
                await userStore().resetAccessRouter();
              } catch (err) {
                console.warn('load user menus failed after login', err);
              }
              // 与 Vue2 版本一致：登录成功后进入个人中心（或 redirect 指定的页面）
              this.$router.push({ path: this.redirect || '/profile/index', query: this.otherQuery }).catch(() => {});
              this.loading = false;
              this.$emit('update:modelValue', false);
            })
            .catch(() => {
              this.loading = false;
            });
        } else {
          // validation failed
        }
      });
    },
    changePass() {
      this.$router.push({ path: '/changePass' });
    },
    getOtherQuery(query) {
      return Object.keys(query).reduce((acc, cur) => {
        if (cur !== 'redirect') acc[cur] = query[cur];
        return acc;
      }, {});
    },
  },
});
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
