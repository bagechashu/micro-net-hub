<template>
  <el-card class="profile-card">
    <template #header>
      <div class="clearfix">
        <span>{{ $t('profile.aboutme') }}</span>
      </div>
    </template>

    <div class="user-profile">
      <div class="box-center">
        <img :src="user.avatar" height="80" width="80" style="border-radius: 10px;">
      </div>
      <div class="box-center">
        <div class="user-name text-center">{{ user.name }}</div>
        <div class="user-name text-center">{{ user.mail }}</div>
        <div class="user-role text-center text-muted">{{ uppercaseFirst(user.role) }}</div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
export interface IProfileUser {
  name: string;
  mail: string;
  avatar: string;
  role: string;
  introduction?: string;
}

defineProps({
  user: {
    type: Object as () => IProfileUser,
    default: () => ({
      name: '',
      mail: '',
      avatar: '',
      role: ''
    })
  }
});

function uppercaseFirst(str?: string) {
  if (!str) return '';
  return str.charAt(0).toUpperCase() + str.slice(1);
}
</script>

<style lang="scss" scoped>
.profile-card {
  min-height: 18rem;
  height: 18rem;
}

.box-center {
  margin: 0 auto;
  display: table;
}

.text-center {
  text-align: center;
}

.text-muted {
  color: var(--el-text-color-secondary);
}

.user-profile {
  .user-name {
    font-weight: bold;
  }

  .box-center {
    padding-top: 10px;
  }

  .user-role {
    padding-top: 10px;
    font-weight: 400;
    font-size: 14px;
  }
}
</style>
