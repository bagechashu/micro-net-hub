<template>
  <div>
    <el-card class="profile-card">
      <template #header>
        <div class="clearfix">
          <span>{{ $t('profile.changePassword') }}</span>
        </div>
      </template>

      <el-form
        ref="dialogFormRef"
        size="small"
        :model="dialogFormData"
        :rules="dialogFormRules"
        :label-width="labelWidth"
      >
        <el-form-item :label="$t('loginform.oldpass')" prop="oldPassword">
          <el-input
            v-model.trim="dialogFormData.oldPassword"
            autocomplete="on"
            :type="passwordTypeOld"
            :placeholder="$t('loginform.oldpassTips')"
          />
          <span class="show-pwd" @click="passwordTypeOld = togglePwd(passwordTypeOld)">
            <svg-icon :icon-class="passwordTypeOld === 'password' ? 'eye' : 'eye-open'" />
          </span>
        </el-form-item>

        <el-form-item :label="$t('loginform.newpass')" prop="newPassword">
          <el-input
            v-model.trim="dialogFormData.newPassword"
            autocomplete="on"
            :type="passwordTypeNew"
            :placeholder="$t('loginform.newpassTips')"
          />
          <span class="show-pwd" @click="passwordTypeNew = togglePwd(passwordTypeNew)">
            <svg-icon :icon-class="passwordTypeNew === 'password' ? 'eye' : 'eye-open'" />
          </span>
        </el-form-item>

        <el-form-item :label="$t('loginform.confirmNewPass')" prop="confirmPassword">
          <el-input
            v-model.trim="dialogFormData.confirmPassword"
            autocomplete="on"
            :type="passwordTypeConfirm"
            :placeholder="$t('loginform.confirmNewPassTips')"
          />
          <span class="show-pwd" @click="passwordTypeConfirm = togglePwd(passwordTypeConfirm)">
            <svg-icon :icon-class="passwordTypeConfirm === 'password' ? 'eye' : 'eye-open'" />
          </span>
        </el-form-item>

        <el-form-item>
          <el-button :loading="submitLoading" type="primary" @click="submitForm">{{ $t('common.confirm') }}</el-button>
          <el-button @click="resetForm">{{ $t('common.cancel') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import JSEncrypt from 'jsencrypt';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { changePwd } from '@/api/system/user';
import useUserStore from '@/store/modules/user';

const { t, locale } = useI18n();

const publicKey = import.meta.env.VITE_PUBLIC_KEY;

const dialogFormRef = ref<FormInstance>();
const submitLoading = ref(false);
const passwordTypeOld = ref('password');
const passwordTypeNew = ref('password');
const passwordTypeConfirm = ref('password');
const dialogFormData = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
});

const labelWidth = computed(() => (locale.value === 'zh' ? '80px' : '200px'));

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

  // 密码RSA加密处理
  const encryptor = new JSEncrypt();
  encryptor.setPublicKey(publicKey);
  const payload = {
    oldPassword: encryptor.encrypt(dialogFormData.oldPassword),
    newPassword: encryptor.encrypt(dialogFormData.newPassword),
    confirmPassword: encryptor.encrypt(dialogFormData.confirmPassword)
  };

  submitLoading.value = true;
  try {
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
  color: var(--el-text-color-secondary);
  cursor: pointer;
  user-select: none;
}
</style>
