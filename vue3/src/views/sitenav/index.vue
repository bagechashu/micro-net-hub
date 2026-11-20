<template>
  <div class="app-container">
    <el-row>
      <el-col :xs="24" :sm="12" :md="10" :lg="10" :xl="6">
        <div class="header-bar">
          <el-input
            v-model="search"
            :placeholder="$t('sitenav.searchTips')"
            class="search"
            @keyup.enter="searchData"
          />
          <span class="search-text">
            <el-button type="primary" :icon="Search" @click="searchData">{{ $t('common.search') }}</el-button>
          </span>
          <el-button v-show="searchStatus" type="success" :icon="Refresh" @click="resetSearch">
            {{ $t('common.reset') }}
          </el-button>
        </div>
      </el-col>
      <el-col :xs="24" :sm="10" :md="13" :lg="13" :xl="17"><Notice /></el-col>
    </el-row>

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

<style lang="css" scoped>
.header-bar {
  display: flex;
  align-items: center;
  background: transparent;
}

.search {
  width: 300px;
}

@media screen and (max-width: 768px) {
  .search {
    width: auto;
  }

  .search-text {
    margin: 0 3px;
  }
}
</style>
