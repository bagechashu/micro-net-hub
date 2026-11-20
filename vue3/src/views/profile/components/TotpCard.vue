<template>
  <el-card class="settings-card">
    <template #header>
      <div class="settings-card__header">
        <span class="settings-card__icon"><IconShieldLock :size="20" stroke="1.75" /></span>
        <div>
          <div class="settings-card__title">{{ $t('profile.totpTitle') }}</div>
          <div class="settings-card__desc">{{ $t('profile.totpDesc') }}</div>
        </div>
      </div>
    </template>

    <!-- 第一步：输入当前 OTP 并确认重置 -->
    <div v-if="showDiscribe" class="totp-reset">
      <div class="totp-alert">
        <IconAlertTriangle :size="18" stroke="1.75" class="totp-alert__icon" />
        <div>
          <div class="totp-alert__title">{{ $t('totpNotice.notice') }}</div>
          <ul>
            <li>{{ $t('totpNotice.content[0]') }}</li>
            <li>{{ $t('totpNotice.content[1]') }}</li>
          </ul>
        </div>
      </div>

      <el-form
        ref="resetTotpSecretFormRef"
        class="totp-form"
        :model="resetTotpSecretFormData"
        :rules="resetTotpSecretFormRules"
        label-position="top"
        @submit.prevent
      >
        <el-form-item :label="$t('profile.changeTotp')" prop="totp">
          <div class="totp-form__row">
            <el-input
              v-model.trim="resetTotpSecretFormData.totp"
              class="totp-form__input"
              maxlength="6"
              inputmode="numeric"
              autocomplete="one-time-code"
              :placeholder="$t('profile.changeTotpTips')"
            />
            <el-popconfirm :title="$t('profile.totpConfirm')" width="260" @confirm="resetTotp">
              <template #reference>
                <el-button type="danger" plain :icon="Refresh" :loading="resetLoading">{{ $t('common.reset') }}</el-button>
              </template>
            </el-popconfirm>
          </div>
        </el-form-item>
      </el-form>
    </div>

    <!-- 第二步：展示新二维码（只显示一次） -->
    <div v-else class="totp-qr">
      <div class="totp-qr__code">
        <QrCode id="QrCode" :text="qrCodeStr" :width="168" />
      </div>
      <div class="totp-qr__text">
        <div class="totp-qr__title">
          <IconCircleCheck :size="18" stroke="1.75" />{{ $t('profile.totpScanTitle') }}
        </div>
        <p>{{ $t('profile.totpScanDesc') }}</p>
        <el-button type="primary" @click="finish">{{ $t('profile.done') }}</el-button>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { Refresh } from '@/icons/tabler';
import { IconShieldLock, IconAlertTriangle, IconCircleCheck } from '@tabler/icons-vue';
import QrCode from '@/components/Qrcode/Qrcode.vue';
import { resetTotpSecret } from '@/api/personnel/user';

const { t } = useI18n();

const resetTotpSecretFormRef = ref<FormInstance>();
const showDiscribe = ref(true);
const qrCodeStr = ref('');
const resetTotpSecretFormData = reactive({
  totp: ''
});
const resetTotpSecretFormRules: FormRules = {
  totp: [
    {
      required: true,
      message: t('profile.changeTotpTips'),
      trigger: 'blur'
    },
    {
      validator: (_rule, value, callback) => {
        if (!/^\d{6}$/.test(value)) {
          return callback(new Error(t('profile.changeTotpTipsValid')));
        }
        callback();
      },
      trigger: 'blur'
    }
  ]
};

const resetLoading = ref(false);

function finish() {
  qrCodeStr.value = '';
  showDiscribe.value = true;
  resetTotpSecretFormRef.value?.resetFields();
}

async function resetTotp() {
  if (!resetTotpSecretFormRef.value) return;
  const valid = await resetTotpSecretFormRef.value.validate().catch(() => false);
  if (!valid) {
    ElMessage({
      showClose: true,
      message: t('tips.formValidFailed'),
      type: 'warning'
    });
    return;
  }
  resetLoading.value = true;
  try {
    const { data } = await resetTotpSecret(resetTotpSecretFormData);
    qrCodeStr.value = data;
    showDiscribe.value = false;
  } catch {
    /* error already shown by request wrapper */
  } finally {
    resetLoading.value = false;
  }
}
</script>

<style lang="scss" scoped>
.totp-alert {
  display: flex;
  gap: 10px;
  padding: 12px 14px;
  margin-bottom: 18px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  background: var(--el-color-warning-light-9);
  border: 1px solid var(--el-color-warning-light-7);
  border-radius: 6px;

  ul {
    margin: 4px 0 0;
    padding-left: 18px;
  }
}

.totp-alert__icon {
  flex-shrink: 0;
  color: var(--el-color-warning);
}

.totp-alert__title {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.totp-form :deep(.el-form-item) {
  margin-bottom: 0;
}

.totp-form__row {
  display: flex;
  gap: 8px;
  width: 100%;
  max-width: 360px;
}

.totp-form__input {
  flex: 1;
  letter-spacing: 2px;
}

.totp-qr {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 24px;
}

.totp-qr__code {
  padding: 8px;
  line-height: 0;
  background: #fff;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
}

.totp-qr__text {
  flex: 1;
  min-width: 220px;

  p {
    margin: 6px 0 16px;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }
}

.totp-qr__title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;

  svg {
    color: var(--el-color-success);
  }
}
</style>
