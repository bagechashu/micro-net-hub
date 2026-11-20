<template>
  <el-row :gutter="16" class="profile-page">
    <el-col :xs="24" :md="8" :lg="7">
      <UserCard />
    </el-col>
    <el-col :xs="24" :md="16" :lg="17">
      <div class="profile-stack">
        <Account />
        <TotpCard />
      </div>
    </el-col>
  </el-row>
</template>

<script setup lang="ts">
import UserCard from './components/UserCard.vue';
import TotpCard from './components/TotpCard.vue';
import Account from './components/Account.vue';

defineOptions({ name: 'Profile' });
</script>

<style scoped>
.profile-page {
  row-gap: 16px;
}

.profile-stack {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
