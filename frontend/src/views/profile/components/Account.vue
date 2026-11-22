<template>
  <el-card class="settings-card">
    <template #header>
      <div class="settings-card__header">
        <span class="settings-card__icon"><IconKey :size="20" stroke="1.75" /></span>
        <div>
          <div class="settings-card__title">{{ $t('profile.changePassword') }}</div>
          <div class="settings-card__desc">{{ $t('profile.changePasswordDesc') }}</div>
        </div>
      </div>
    </template>

    <div class="pwd-body">
      <el-form
        ref="dialogFormRef"
        class="pwd-form"
        :model="dialogFormData"
        :rules="dialogFormRules"
        label-position="top"
      >
            <el-form-item :label="$t('loginform.oldpass')" prop="oldPassword">
              <el-tooltip :visible="capsField === 'old'" :content="$t('loginform.capsLockOn')" placement="top">
                <el-input
                  v-model.trim="dialogFormData.oldPassword"
                  autocomplete="on"
                  :type="passwordTypeOld"
                  :placeholder="$t('loginform.oldpassTips')"
                  @keydown="checkCapslock($event, 'old')"
                  @keyup="checkCapslock($event, 'old')"
                  @blur="clearCapslock('old')"
                >
                  <template #suffix>
                    <span class="show-pwd" @click="passwordTypeOld = togglePwd(passwordTypeOld)">
                      <component :is="passwordTypeOld === 'password' ? IconEyeOff : IconEye" :size="18" stroke="1.75" />
                    </span>
                  </template>
                </el-input>
              </el-tooltip>
            </el-form-item>
        <div class="pwd-form__row">
            <el-form-item :label="$t('loginform.newpass')" prop="newPassword">
              <el-tooltip :visible="capsField === 'new'" :content="$t('loginform.capsLockOn')" placement="top">
                <el-input
                  v-model.trim="dialogFormData.newPassword"
                  autocomplete="on"
                  :type="passwordTypeNew"
                  :placeholder="$t('loginform.newpassTips')"
                  @keydown="checkCapslock($event, 'new')"
                  @keyup="checkCapslock($event, 'new')"
                  @blur="clearCapslock('new')"
                >
                  <template #suffix>
                    <span class="show-pwd" @click="passwordTypeNew = togglePwd(passwordTypeNew)">
                      <component :is="passwordTypeNew === 'password' ? IconEyeOff : IconEye" :size="18" stroke="1.75" />
                    </span>
                  </template>
                </el-input>
              </el-tooltip>
            </el-form-item>
            <el-form-item :label="$t('loginform.confirmNewPass')" prop="confirmPassword">
              <el-tooltip :visible="capsField === 'confirm'" :content="$t('loginform.capsLockOn')" placement="top">
                <el-input
                  v-model.trim="dialogFormData.confirmPassword"
                  autocomplete="on"
                  :type="passwordTypeConfirm"
                  :placeholder="$t('loginform.confirmNewPassTips')"
                  @keydown="checkCapslock($event, 'confirm')"
                  @keyup="checkCapslock($event, 'confirm')"
                  @blur="clearCapslock('confirm')"
                >
                  <template #suffix>
                    <span class="show-pwd" @click="passwordTypeConfirm = togglePwd(passwordTypeConfirm)">
                      <component :is="passwordTypeConfirm === 'password' ? IconEyeOff : IconEye" :size="18" stroke="1.75" />
                    </span>
                  </template>
                </el-input>
              </el-tooltip>
            </el-form-item>
        </div>
      </el-form>

      <aside class="pwd-rules">
        <div class="pwd-rules__title">{{ $t('profile.pwdRuleTitle') }}</div>
        <ul>
          <li v-for="(rule, i) in pwdRules" :key="i" :class="{ 'is-ok': rule.ok }">
            <component :is="rule.ok ? IconCircleCheck : IconCircle" :size="16" stroke="1.75" />
            <span>{{ rule.text }}</span>
          </li>
        </ul>
      </aside>
    </div>

    <template #footer>
      <div class="settings-card__footer">
        <el-button @click="resetForm">{{ $t('common.cancel') }}</el-button>
        <el-button :loading="submitLoading" type="primary" @click="submitForm">{{ $t('profile.changePassword') }}</el-button>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { encryptPassword } from '@/utils/rsa';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { changePwd } from '@/api/system/user';
import useUserStore from '@/store/modules/user';
import { IconEye, IconEyeOff } from '@/icons/tabler';
import { IconKey, IconCircle, IconCircleCheck } from '@tabler/icons-vue';

const { t } = useI18n();


const dialogFormRef = ref<FormInstance>();
const submitLoading = ref(false);
const passwordTypeOld = ref('password');
const passwordTypeNew = ref('password');
const passwordTypeConfirm = ref('password');
// 当前开启了大写锁定、需要显示提示的口令框（与登录框做法一致：直接读取 CapsLock 状态）
const capsField = ref<'' | 'old' | 'new' | 'confirm'>('');

function checkCapslock(e: Event, field: 'old' | 'new' | 'confirm') {
  const ke = e as KeyboardEvent;
  if (typeof ke.getModifierState !== 'function') return;
  if (ke.getModifierState('CapsLock')) {
    capsField.value = field;
  } else if (capsField.value === field) {
    capsField.value = '';
  }
}

function clearCapslock(field: 'old' | 'new' | 'confirm') {
  if (capsField.value === field) capsField.value = '';
}
const dialogFormData = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
});

// 右侧要求清单：随输入实时打勾
const pwdRules = computed(() => {
  const v = dialogFormData.newPassword;
  return [
    { text: t('profile.pwdRules[0]'), ok: v.length >= 8 },
    { text: t('profile.pwdRules[1]'), ok: checkPasswordComplexity(v) >= 3 },
    { text: t('profile.pwdRules[2]'), ok: !!v && v !== dialogFormData.oldPassword }
  ];
});

function checkPasswordComplexity(value: string) {
  let matches = 0;
  if (/[a-z]/.test(value)) matches++;
  if (/[A-Z]/.test(value)) matches++;
  if (/\d/.test(value)) matches++;
  if (/[!@#$%^&*,.?_-]/.test(value)) matches++;
  return matches;
}

// same rules as Vue2 utils/validate.js validatePassword
const validatePassword = (_rule, value, callback) => {
  if (!value) {
    return callback(new Error(t('loginform.newpassTips')));
  }
  if (value.length < 8) {
    return callback(new Error(t('loginform.newpassValidErrLenMustThan8')));
  }
  if (checkPasswordComplexity(value) < 3) {
    return callback(new Error(t('loginform.newpassValidErrMustComplex')));
  }
  callback();
};

const confirmPass = (_rule, value, callback) => {
  if (value) {
    if (dialogFormData.newPassword !== value) {
      callback(new Error(t('loginform.confirmNewPassErr')));
    } else {
      callback();
    }
  } else {
    callback(new Error(t('loginform.confirmNewPassTips')));
  }
};

const dialogFormRules: FormRules = {
  oldPassword: [
    { required: true, message: t('loginform.oldpassTips'), trigger: 'blur' },
    { min: 6, max: 30, message: t('valid.length', [6, 30]), trigger: 'blur' }
  ],
  newPassword: [{ required: true, validator: validatePassword, trigger: 'blur' }],
  confirmPassword: [{ required: true, validator: confirmPass, trigger: 'blur' }]
};

function togglePwd(type: string) {
  return type === 'password' ? '' : 'password';
}

async function submitForm() {
  if (!dialogFormRef.value) return;
  const valid = await dialogFormRef.value.validate().catch(() => false);
  if (!valid) {
    ElMessage({
      showClose: true,
      message: t('tips.formValidFailed'),
      type: 'warning'
    });
    return;
  }

  submitLoading.value = true;
  try {
    // 口令用后端当前公钥加密后提交
    let payload;
    try {
      payload = {
        oldPassword: await encryptPassword(dialogFormData.oldPassword),
        newPassword: await encryptPassword(dialogFormData.newPassword),
        confirmPassword: await encryptPassword(dialogFormData.confirmPassword)
      };
    } catch {
      ElMessage({ showClose: true, message: t('loginform.encryptFailed'), type: 'error' });
      return;
    }
    await changePwd(payload);
    ElMessage({
      showClose: true,
      message: t('loginform.changePasswordSuccess'),
      type: 'success'
    });
    resetForm();
    // 重新登录
    setTimeout(async () => {
      try {
        await useUserStore().logout();
      } finally {
        location.reload(); // 为了重新实例化vue-router对象 避免bug
      }
    }, 1500);
  } catch {
    /* error already shown by request wrapper */
  } finally {
    submitLoading.value = false;
  }
}

function resetForm() {
  dialogFormRef.value?.resetFields();
  dialogFormData.oldPassword = '';
  dialogFormData.newPassword = '';
  dialogFormData.confirmPassword = '';
}
</script>

<style lang="scss" scoped>
.pwd-body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 240px;
  gap: 24px;
}

.pwd-form__row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.pwd-form :deep(.el-form-item:last-child),
.pwd-form__row :deep(.el-form-item) {
  margin-bottom: 0;
}

.pwd-form :deep(.el-form-item) {
  margin-bottom: 18px;
}

.pwd-form__row :deep(.el-form-item) {
  margin-bottom: 0;
}

.show-pwd {
  display: inline-flex;
  align-items: center;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  user-select: none;

  &:hover {
    color: var(--el-text-color-primary);
  }
}

.pwd-rules {
  padding: 14px 16px;
  font-size: 13px;
  background: var(--el-fill-color-light);
  border-radius: 6px;

  ul {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    margin-top: 8px;
    color: var(--el-text-color-secondary);

    svg {
      flex-shrink: 0;
      margin-top: 1px;
    }

    &.is-ok {
      color: var(--el-color-success);
    }
  }
}

.pwd-rules__title {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

@media (max-width: 992px) {
  .pwd-body {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 576px) {
  .pwd-form__row {
    grid-template-columns: 1fr;
    gap: 18px;
  }
}
</style>
