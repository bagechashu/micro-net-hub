<template>
  <div class="app-container sitenav">
    <div class="sitenav-toolbar">
      <div class="sitenav-search">
        <el-input
          v-model="search"
          :placeholder="$t('sitenav.searchTips')"
          :prefix-icon="Search"
          clearable
          class="sitenav-search__input"
          @keyup.enter="searchData"
          @clear="searchStatus && resetSearch()" />
        <el-button type="primary" @click="searchData">{{ $t('common.search') }}</el-button>
        <el-button v-show="searchStatus" :icon="Refresh" @click="resetSearch">{{ $t('common.reset') }}</el-button>
      </div>
      <div class="sitenav-notice"><Notice /></div>
    </div>

    <NavSub v-if="data.length > 0" :data="data" />
    <el-empty v-else :description="$t('common.nodata')" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { Search, Refresh } from '@element-plus/icons-vue';
import NavSub from '@/views/sitenav/components/sub.vue';
import Notice from '@/views/notice/index.vue';
import { getNav } from '@/api/sitenav';

defineOptions({ name: 'SiteNav' });

const { t } = useI18n();

const search = ref('');
const searchStatus = ref(false);
const data = ref<any[]>([]);
let sourceData: any[] = [];

async function getNavJson() {
  try {
    const res = await getNav();
    data.value = Array.isArray(res.data) ? res.data : [];
  } catch {
    // error already shown by request wrapper
  }
}

function searchData() {
  const keyword = (search.value || '').trim().toLowerCase();
  if (!keyword) {
    ElMessage({ message: t('valid.pleaseInput'), type: 'error' });
    return;
  }

  if (!searchStatus.value) {
    sourceData = JSON.parse(JSON.stringify(data.value));
  }
  searchStatus.value = true;

  let searchNum = 0;
  data.value = (JSON.parse(JSON.stringify(sourceData)) as any[])
    .map(group => {
      const sites = (group.sites || []).filter((site: any) =>
        String(site.name ?? '').toLowerCase().includes(keyword) ||
        String(site.link ?? '').toLowerCase().includes(keyword)
      );
      searchNum += sites.length;
      return { ...group, sites };
    })
    .filter(group => group.sites.length > 0);

  if (searchNum === 0) {
    ElMessage({ message: t('tips.notFoundAndRetry'), type: 'error' });
  } else {
    ElMessage({ message: t('tips.foundSome', [searchNum]), type: 'success' });
  }
}

function resetSearch() {
  searchStatus.value = false;
  search.value = '';
  data.value = JSON.parse(JSON.stringify(sourceData));
}

getNavJson();
</script>

<style lang="scss" scoped>
.sitenav-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.sitenav-search {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sitenav-search__input {
  width: 320px;
}

.sitenav-notice {
  flex: 1;
  min-width: 240px;
}

@media (max-width: 768px) {
  .sitenav-search,
  .sitenav-search__input {
    width: 100%;
  }
}
</style>
