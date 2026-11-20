<template>
  <div>
    <div v-for="(item, index) in data" :id="item.title" :key="index">
      <el-card v-if="item.sites" class="node-card" shadow="hover" :body-style="{ padding: '5px' }">
        <template #header>
          <div class="clearfix card-header">
            <span style="float: left; padding: 3px 0">{{ item.title }}<a :href="'#' + item.title" /> </span>
            <span style="float: right; padding: 3px 0">{{ $t('sitenav.totolInfo', [item.sites.length]) }}</span>
          </div>
        </template>
        <Nav :sites="item.sites" />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import Nav from './nav.vue';

defineProps({
  data: {
    type: Array as () => any[],
    default: () => []
  }
});
</script>

<style scoped>
.card-header {
  font-size: 14px;
}
:deep(.el-card__header) {
  padding: 8px;
}
.node-card {
  margin-top: 1rem;
}
.clearfix:before,
.clearfix:after {
  display: table;
  content: "";
}
.clearfix:after {
  clear: both;
}
</style>
