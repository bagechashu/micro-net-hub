<template>
  <header class="app-header">
    <div class="app-container-xl app-header__inner">
      <router-link to="/" class="app-header__brand">
        <img :src="logo" alt="Micro Net Hub" class="app-header__logo">
        <span class="app-header__title">Micro Net Hub</span>
      </router-link>

      <div class="app-header__actions">
        <el-tooltip :content="$t('theme.tooltip')" placement="bottom">
          <theme-switch class="icon-btn" />
        </el-tooltip>

        <el-tooltip content="Language" placement="bottom">
          <language-select class="icon-btn" />
        </el-tooltip>

        <el-tooltip content="Doc Refs [eryajf]" placement="bottom">
          <a class="icon-btn" href="http://ldapdoc.eryajf.net" target="_blank" rel="noopener">
            <el-icon :size="18"><Document /></el-icon>
          </a>
        </el-tooltip>

        <el-dropdown v-if="user.token" trigger="click" class="app-header__user">
          <div class="user-block">
            <img :src="user.avatar || defaultAvatar" class="user-block__avatar" alt="">
            <div class="user-block__text">
              <div class="user-block__name">{{ user.name }}</div>
              <div class="user-block__role">{{ (user.roles || []).join(' | ') }}</div>
            </div>
            <el-icon class="user-block__caret"><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <router-link to="/profile/index">
                <el-dropdown-item :icon="User">{{ $t('profile.profile') }}</el-dropdown-item>
              </router-link>
              <el-dropdown-item divided :icon="SwitchButton" @click="logout">
                {{ $t('loginform.logout') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button v-else type="primary" class="app-header__login" @click="loginVisible = true">
          {{ $t('loginform.login') }}
        </el-button>
      </div>
    </div>
    <Login v-model="loginVisible" />
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { ArrowDown, Document, SwitchButton, User } from '@element-plus/icons-vue';
import useUserStore from '@/store/modules/user';
import ThemeSwitch from '@/components/ThemeSwitch/index.vue';
import LanguageSelect from '@/components/LanguageSelect/index.vue';
import Login from '@/layout/components/Login/index.vue';
import logo from '@/assets/sidebar-logo/micro-net-hub.png';

defineOptions({ name: 'AppHeader' });

const user = useUserStore();
const router = useRouter();
const loginVisible = ref(false);
const defaultAvatar = 'https://q1.qlogo.cn/g?b=qq&nk=10002&s=100';

async function logout() {
  await user.logout();
  router.push('/');
}
</script>

<style lang="scss" scoped>
.app-header {
  background: var(--app-header-bg);
  border-bottom: 1px solid var(--app-header-border);
}

.app-header__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
}

.app-header__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--el-text-color-primary);
  text-decoration: none;
}

.app-header__logo {
  height: 32px;
  border-radius: 6px;
}

.app-header__title {
  font-size: 18px;
  font-weight: 600;
  letter-spacing: -0.01em;
}

.app-header__actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

// 统一的圆形图标按钮（主题、语言、文档）
.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  color: var(--el-text-color-regular);
  cursor: pointer;
  transition: background-color 0.15s;

  &:hover {
    background: var(--app-icon-btn-hover);
    color: var(--el-text-color-primary);
  }

  :deep(.svg-icon),
  :deep(.language-icon) {
    width: 18px;
    height: 18px;
  }
}

.app-header__login {
  margin-left: 8px;
}

.app-header__user {
  margin-left: 8px;
}

.user-block {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px;
  border-radius: 6px;
  cursor: pointer;

  &:hover {
    background: var(--app-icon-btn-hover);
  }
}

.user-block__avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  box-shadow: 0 0 0 1px var(--el-border-color-light);
}

.user-block__text {
  line-height: 1.25;
}

.user-block__name {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.user-block__role {
  font-size: 12px;
  color: var(--app-muted-text);
}

.user-block__caret {
  color: var(--app-muted-text);
}

@media (max-width: 768px) {
  .app-header__title,
  .user-block__text {
    display: none;
  }
}
</style>
