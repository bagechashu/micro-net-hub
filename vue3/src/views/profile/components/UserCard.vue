<template>
  <el-card class="user-card">
    <div class="user-card__head">
      <img v-if="avatar" :src="avatar" class="user-card__avatar" alt="" @error="avatarFailed = true">
      <span v-else class="user-card__avatar user-card__avatar--initials">{{ initials }}</span>
      <div class="user-card__name">{{ displayName }}</div>
      <div class="user-card__sub">@{{ info?.username }}</div>
      <div v-if="info?.position" class="user-card__sub">{{ info.position }}</div>
      <div v-if="roles.length" class="user-card__roles">
        <span v-for="r in roles" :key="r" class="badge-soft">{{ r }}</span>
      </div>
    </div>

    <p v-if="info?.introduction" class="user-card__intro">{{ info.introduction }}</p>

    <dl class="user-card__list">
      <div v-for="item in details" :key="item.label" class="user-card__item">
        <dt><component :is="item.icon" :size="16" stroke="1.75" />{{ item.label }}</dt>
        <dd :class="{ 'is-empty': !item.value }">{{ item.value || $t('profile.notSet') }}</dd>
      </div>
    </dl>
  </el-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { IconMail, IconPhone, IconId, IconUserCircle, IconTag, IconPlugConnected } from '@tabler/icons-vue';
import useUserStore from '@/store/modules/user';

defineOptions({ name: 'UserCard' });

const { t } = useI18n();
const userStore = useUserStore();
const info = computed(() => userStore.info);
const avatarFailed = ref(false);

const str = (v: unknown) => (typeof v === 'string' ? v : '');

const displayName = computed(() => str(info.value?.nickname) || info.value?.username || userStore.name);
const avatar = computed(() => (avatarFailed.value ? '' : userStore.avatar));
// 没有头像时显示姓名首字（中文取最后一个字，英文取前两个字母）
const initials = computed(() => {
  const name = displayName.value || '?';
  return /[一-龥]/.test(name) ? name.slice(-1) : name.slice(0, 2).toUpperCase();
});
const roles = computed(() => userStore.roles || []);

const details = computed(() => [
  { label: t('profile.mail'), value: str(info.value?.mail), icon: IconMail },
  { label: t('profile.mobile'), value: str(info.value?.mobile), icon: IconPhone },
  { label: t('profile.jobNumber'), value: str(info.value?.jobNumber), icon: IconId },
  { label: t('profile.givenName'), value: str(info.value?.givenName), icon: IconTag },
  { label: t('profile.username'), value: str(info.value?.username), icon: IconUserCircle },
  { label: t('profile.source'), value: str(info.value?.source), icon: IconPlugConnected }
]);
</script>

<style lang="scss" scoped>
.user-card__head {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 8px 0 16px;
}

.user-card__avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  object-fit: cover;
  margin-bottom: 12px;
  box-shadow: 0 0 0 4px var(--el-color-primary-light-9);
}

.user-card__avatar--initials {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 600;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.user-card__name {
  font-size: 18px;
  font-weight: 600;
  line-height: 1.4;
}

.user-card__sub {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.user-card__roles {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 6px;
  margin-top: 12px;
}

.badge-soft {
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 4px;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.user-card__intro {
  margin: 0 0 16px;
  padding: 12px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
  border-radius: 6px;
}

.user-card__list {
  margin: 0;
  border-top: 1px solid var(--el-border-color-lighter);
}

.user-card__item {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 0;
  font-size: 13px;
  border-bottom: 1px solid var(--el-border-color-lighter);

  &:last-child {
    border-bottom: 0;
    padding-bottom: 0;
  }

  dt {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
    color: var(--el-text-color-secondary);
  }

  dd {
    margin: 0;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--el-text-color-primary);

    &.is-empty {
      color: var(--el-text-color-placeholder);
    }
  }
}
</style>
