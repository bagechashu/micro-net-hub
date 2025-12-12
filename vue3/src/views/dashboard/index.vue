<template>
  <div class="dashboard-container">
    <component :is="currentRole" />
  </div>
</template>

<script>
import { defineComponent } from 'vue';
import { mapState } from 'pinia';
import adminDashboard from './admin';
import editorDashboard from './editor';
import userStore from '@/store/modules/user';

export default defineComponent({
  name: 'Dashboard',
  components: { adminDashboard, editorDashboard },
  data() {
    return {
      currentRole: 'adminDashboard'
    };
  },
  computed: {
    ...mapState(userStore, ['roles'])
  },
  created() {
    console.log('dashboard created');
    if (!this.roles.includes('admin')) {
      this.currentRole = 'editorDashboard';
    }
  }
});
</script>
