<template>
  <el-card class="profile-card">
    <template #header>
      <div class="clearfix">
        <span>{{ $t('profile.changeTotp') }}</span>
      </div>
    </template>
    <div class="box-center">
      <div v-show="showDiscribe">
        <div class="warning-message">
          <p><b>{{ $t('totpNotice.notice') }}</b></p>
          <p>{{ $t('totpNotice.content[0]') }}</p>
          <p>{{ $t('totpNotice.content[1]') }}</p>
        </div>
        <el-form
          ref="resetTotpSecretFormRef"
          size="small"
          :model="resetTotpSecretFormData"
          :rules="resetTotpSecretFormRules"
          label-width="auto"
        >
          <el-form-item label="OTP" prop="totp">
            <el-row justify="space-between" align="middle" style="width: 100%;">
              <el-col :xs="16" :sm="14" :md="16" :lg="16" :xl="16">
                <el-input
                  v-model.trim="resetTotpSecretFormData.totp"
                  :placeholder="$t('profile.changeTotpTips')"
                />
              </el-col>

              <el-col :xs="8" :sm="10" :md="8" :lg="8" :xl="8">
                <el-popconfirm
                  :title="$t('common.areyousure')"
                  @confirm="resetTotp"
                >
                  <template #reference>
                    <el-button
                      class="right"
                      size="small"
                      :icon="Refresh"
                      type="danger"
                    >{{ $t('common.reset') }}
                    </el-button>
                  </template>
                </el-popconfirm>
              </el-col>
            </el-row>
          </el-form-item>
        </el-form>
      </div>
      <QrCode id="QrCode" class="mt-30" :text="qrCodeStr" />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { Refresh } from '@element-plus/icons-vue';
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
  try {
    const { data } = await resetTotpSecret(resetTotpSecretFormData);
    qrCodeStr.value = data;
    showDiscribe.value = false;
  } catch {
    /* error already shown by request wrapper */
  }
}
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
  background-color: var(--el-color-danger-light-9);
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
