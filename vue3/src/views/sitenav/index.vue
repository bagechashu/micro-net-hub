<template>
  <div class="app-container sitenav">
    <Teleport to="#page-header-actions" defer>
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
    </Teleport>

    <!-- 公告：独占一行；没有公告时不渲染，导航内容直接贴着标题 -->
    <Notice class="sitenav-notice" />

    <NavSub v-if="data.length > 0" :data="data" />
    <el-empty v-else :description="$t('common.nodata')" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { Search, Refresh } from '@/icons/tabler';
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
// 搜索框通过 Teleport 渲染在页面标题行右侧，scoped 样式仍然生效
.sitenav-search {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sitenav-search__input {
  width: 320px;
}

.sitenav-notice {
  margin-bottom: 20px;
}

@media (max-width: 768px) {
  .sitenav-search,
  .sitenav-search__input {
    width: 100%;
  }
}
</style>
