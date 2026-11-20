<template>
  <div class="changepass-container">
    <el-card shadow="always" style="margin-top: 5rem">
      <template #header>
        <div>
          <b>{{ $t('changePassword.6111lkey47k0') }}</b>
        </div>
      </template>
      <el-form :model="form" size="default" label-width="auto">
        <el-form-item required :label="$t('changePassword.6111lkey4lw0')">
          <div class="input-container">
            <el-input v-model="form.mail" :placeholder="$t('changePassword.6111lkey4r00')" />
            <el-button :loading="vCodeLoading" type="primary" @click="sendVerificationCode">{{ $t('changePassword.6111lkey4tw0') }}</el-button>
          </div>
        </el-form-item>
        <el-form-item required :label="$t('changePassword.6111lkey4wg0')">
          <el-input v-model="form.code" :placeholder="$t('changePassword.6111lkey4zc0')" />
        </el-form-item>
        <el-form-item class="reset-item">
          <el-button :loading="resetPassLoading" type="primary" @click="resetPass">{{ $t('changePassword.6111lkey47k0') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { emailPass, sendCode } from '@/api/system/user';

defineOptions({ name: 'ChangePass' });

const { t } = useI18n();
const router = useRouter();

const vCodeLoading = ref(false);
const resetPassLoading = ref(false);
const form = reactive({
  mail: '',
  code: ''
});

// 发送邮箱验证码
async function sendVerificationCode() {
  vCodeLoading.value = true;
  try {
    await sendCode({ mail: form.mail });
    ElMessage({
      showClose: true,
      message: t('changePassword.otpSendTips'),
      type: 'success'
    });
    // 60s 后才允许再次发送
    setTimeout(() => {
      vCodeLoading.value = false;
    }, 60000);
  } catch {
    vCodeLoading.value = false;
  }
}

// 重置密码
async function resetPass() {
  resetPassLoading.value = true;
  try {
    await emailPass(form);
    ElMessage({
      showClose: true,
      message: t('changePassword.resetSuccessTips'),
      type: 'success'
    });
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
.changepass-container {
  display: flex;
  justify-content: center;
}

.input-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.input-container .el-input {
  flex: 1;
  margin-right: 10px;
}

.reset-item :deep(.el-form-item__content) {
  justify-content: flex-end;
}
</style>
