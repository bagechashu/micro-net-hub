<template>
  <div class="app-container">
    <div v-if="user">
      <el-row :gutter="20">
        <el-col :span="6" :xs="24">
          <UserCard :user="user" />
        </el-col>
        <el-col :span="8" :xs="24">
          <TotpCard />
        </el-col>
        <el-col :span="10" :xs="24">
          <Account />
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import useUserStore from '@/store/modules/user';
import UserCard from './components/UserCard.vue';
import TotpCard from './components/TotpCard.vue';
import Account from './components/Account.vue';

defineOptions({ name: 'Profile' });

const userStore = useUserStore();

const user = computed(() => ({
  name: userStore.name,
  role: (userStore.roles || []).join(' | '),
  mail: userStore.mail,
  avatar: userStore.avatar ? userStore.avatar : 'https://q1.qlogo.cn/g?b=qq&nk=10002&s=100',
  introduction: userStore.introduction
}));
</script>
