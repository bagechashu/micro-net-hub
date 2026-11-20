<template>
  <ErrorPage
    code="401"
    tone="warning"
    :icon="IconLock"
    :title="$t('errorPage.forbiddenTitle')"
    :desc="$t('errorPage.forbiddenDesc')"
  />
</template>

<script setup lang="ts">
import { IconLock } from '@tabler/icons-vue';
import ErrorPage from './ErrorPage.vue';

defineOptions({ name: 'Page401' });
</script>
