<template>
  <div class="nav-groups">
    <section v-for="item in data" :id="item.title" :key="item.ID ?? item.title" class="nav-group">
      <template v-if="item.sites">
        <header class="nav-group__header">
          <h3 class="nav-group__title">
            <IconStarFilled v-if="item.bookmark" :size="16" class="nav-group__star" />
            {{ item.title }}
          </h3>
          <span class="nav-group__meta">
            <span v-if="item.bookmark" class="nav-group__hint">{{ $t('sitenav.bookmarksHint') }}</span>
            <span class="nav-group__count">{{ $t('sitenav.totolInfo', [item.sites.length]) }}</span>
          </span>
        </header>
        <Nav :sites="item.sites" />
      </template>
    </section>
  </div>
</template>

<script setup lang="ts">
import { IconStarFilled } from '@tabler/icons-vue';
import Nav from './nav.vue';

defineProps({
  data: {
    type: Array as () => any[],
    default: () => []
  }
});
</script>

<style lang="scss" scoped>
.nav-group + .nav-group {
  margin-top: 28px;
}

.nav-group__header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 12px;
}

.nav-group__title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.nav-group__star {
  color: var(--el-color-warning);
}

.nav-group__meta {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.nav-group__hint {
  font-size: 12px;
  color: var(--app-muted-text);
}

.nav-group__count {
  font-size: 12px;
  font-weight: 500;
  color: var(--app-muted-text);
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--el-fill-color-light);
}
</style>
