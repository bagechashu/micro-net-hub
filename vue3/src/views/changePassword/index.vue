<template>
  <div class="changepass">
    <el-card class="changepass__card">
      <div class="changepass__head">
        <span class="changepass__icon"><IconLockQuestion :size="28" stroke="1.75" /></span>
        <h3 class="changepass__title">{{ $t('changePassword.6111lkey47k0') }}</h3>
        <p class="changepass__desc">{{ $t('changePassword.desc') }}</p>
      </div>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent>
        <el-form-item :label="$t('changePassword.6111lkey4lw0')" prop="mail">
          <div class="changepass__mail">
            <el-input v-model.trim="form.mail" type="email" autocomplete="email" :placeholder="$t('changePassword.6111lkey4r00')">
              <template #prefix><IconMail :size="16" stroke="1.75" /></template>
            </el-input>
            <el-button :loading="vCodeLoading" :disabled="countdown > 0" @click="sendVerificationCode">
              {{ countdown > 0 ? `${countdown}s` : $t('changePassword.6111lkey4tw0') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="$t('changePassword.6111lkey4wg0')" prop="code">
          <el-input
            v-model.trim="form.code"
            autocomplete="one-time-code"
            :placeholder="$t('changePassword.6111lkey4zc0')"
            @keyup.enter="resetPass"
          >
            <template #prefix><IconShieldCheck :size="16" stroke="1.75" /></template>
          </el-input>
        </el-form-item>
        <el-button class="changepass__submit" type="primary" :loading="resetPassLoading" @click="resetPass">
          {{ $t('changePassword.6111lkey47k0') }}
        </el-button>
      </el-form>
    </el-card>

    <router-link to="/" class="changepass__back">
      <IconArrowLeft :size="16" stroke="1.75" />{{ $t('changePassword.backHome') }}
    </router-link>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { IconArrowLeft, IconLockQuestion, IconMail, IconShieldCheck } from '@tabler/icons-vue';
import { emailPass, sendCode } from '@/api/system/user';

defineOptions({ name: 'ChangePass' });

const { t } = useI18n();
const router = useRouter();

const formRef = ref<FormInstance>();
const vCodeLoading = ref(false);
const resetPassLoading = ref(false);
const countdown = ref(0);
let timer: ReturnType<typeof setInterval> | undefined;

const form = reactive({
  mail: '',
  code: ''
});

const rules: FormRules = {
  mail: [
    { required: true, message: t('changePassword.6111lkey4r00'), trigger: 'blur' },
    { type: 'email', message: t('changePassword.6111lkey4r00'), trigger: 'blur' }
  ],
  code: [{ required: true, message: t('changePassword.6111lkey4zc0'), trigger: 'blur' }]
};

// 60s 内不允许再次发送
function startCountdown() {
  countdown.value = 60;
  clearInterval(timer);
  timer = setInterval(() => {
    countdown.value -= 1;
    if (countdown.value <= 0) clearInterval(timer);
  }, 1000);
}

onBeforeUnmount(() => clearInterval(timer));

// 发送邮箱验证码
async function sendVerificationCode() {
  const ok = await formRef.value?.validateField('mail').then(() => true).catch(() => false);
  if (!ok) return;
  vCodeLoading.value = true;
  try {
    await sendCode({ mail: form.mail });
    ElMessage({ showClose: true, message: t('changePassword.otpSendTips'), type: 'success' });
    startCountdown();
  } catch {
    /* error already shown by request wrapper */
  } finally {
    vCodeLoading.value = false;
  }
}

// 重置密码
async function resetPass() {
  const ok = await formRef.value?.validate().then(() => true).catch(() => false);
  if (!ok) return;
  resetPassLoading.value = true;
  try {
    await emailPass(form);
    ElMessage({ showClose: true, message: t('changePassword.resetSuccessTips'), type: 'success' });
    // 返回首页
    setTimeout(() => {
      router.replace({ path: '/' });
    }, 3000);
  } catch {
    /* error already shown by request wrapper */
  } finally {
    resetPassLoading.value = false;
  }
}
</script>

<style scoped lang="scss">
.changepass {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0 8px;
}

.changepass__card {
  width: 100%;
  max-width: 440px;

  :deep(.el-card__body) {
    padding: 32px;
  }
}

.changepass__head {
  margin-bottom: 24px;
  text-align: center;
}

.changepass__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  margin-bottom: 12px;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  border-radius: 50%;
}

.changepass__title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 600;
}

.changepass__desc {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.changepass__mail {
  display: flex;
  gap: 8px;
  width: 100%;

  .el-input {
    flex: 1;
  }

  .el-button {
    min-width: 108px;
  }
}

.changepass__submit {
  width: 100%;
  margin-top: 6px;
}

.changepass__back {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-top: 16px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  text-decoration: none;

  &:hover {
    color: var(--el-color-primary);
  }
}
</style>
