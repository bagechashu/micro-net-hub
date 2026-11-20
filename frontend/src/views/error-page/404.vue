<template>
  <ErrorPage
    code="404"
    :icon="IconMapSearch"
    :title="$t('errorPage.notFoundTitle')"
    :desc="$t('errorPage.notFoundDesc')"
  />
</template>

<script setup lang="ts">
import { IconMapSearch } from '@tabler/icons-vue';
import ErrorPage from './ErrorPage.vue';

defineOptions({ name: 'Page404' });
</script>
