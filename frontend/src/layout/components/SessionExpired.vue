<template>
  <!-- 登录过期提示（Tabler modal-status 风格）：顶部警示色条 + 图标 + 说明 + 两个操作 -->
  <el-dialog
    :model-value="app.sessionExpired"
    width="400px"
    class="session-expired"
    :show-close="false"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
    append-to-body
  >
    <div class="session-expired__body">
      <span class="session-expired__icon"><IconClockExclamation :size="30" stroke="1.75" /></span>
      <h3 class="session-expired__title">{{ $t('session.expiredTitle') }}</h3>
      <p class="session-expired__desc">{{ $t('session.expiredDesc') }}</p>
    </div>
    <template #footer>
      <div class="session-expired__actions">
        <el-button :loading="busy === 'home'" :disabled="!!busy" @click="backHome">{{ $t('session.backHome') }}</el-button>
        <el-button type="primary" :loading="busy === 'login'" :disabled="!!busy" @click="relogin">{{ $t('session.relogin') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { IconClockExclamation } from '@tabler/icons-vue';
import useAppStore from '@/store/modules/app';
import useUserStore from '@/store/modules/user';

defineOptions({ name: 'SessionExpired' });

const app = useAppStore();
const user = useUserStore();
const router = useRouter();
const busy = ref<'' | 'home' | 'login'>('');

// 清掉本地登录状态（令牌已失效，后端登出失败也没关系）
async function clearSession() {
  await user.logout().catch(() => undefined);
}

async function backHome() {
  busy.value = 'home';
  await clearSession();
  app.sessionExpired = false;
  busy.value = '';
  router.push('/');
}

// 重新登录：回到导航首页并打开登录框，登录成功后回到刚才的页面
async function relogin() {
  busy.value = 'login';
  const back = router.currentRoute.value.fullPath;
  await clearSession();
  app.sessionExpired = false;
  busy.value = '';
  const query = back && back !== '/' && !back.startsWith('/sitenav') ? { redirect: back } : undefined;
  await router.push({ path: '/sitenav', query });
  app.loginDialogVisible = true;
}
</script>

<style lang="scss">
// 弹框挂在 body 上，样式不能用 scoped
.el-dialog.session-expired {
  border-top: 3px solid var(--el-color-warning);

  .el-dialog__header {
    display: none;
  }

  .el-dialog__body {
    padding: 28px 28px 8px;
  }

  .el-dialog__footer {
    padding: 16px 28px 24px;
    background: transparent;
    border-top: 0;
  }
}

.session-expired__body {
  text-align: center;
}

.session-expired__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  margin-bottom: 14px;
  color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
  border-radius: 50%;
}

.session-expired__title {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.session-expired__desc {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.session-expired__actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  width: 100%;

  .el-button + .el-button {
    margin-left: 0;
  }
}
</style>
