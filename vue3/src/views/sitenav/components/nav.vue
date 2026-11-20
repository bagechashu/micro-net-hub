<template>
  <div class="site-grid">
    <el-popover
      v-for="(item, index) in sites"
      :key="item.ID ?? index"
      placement="top"
      :width="180"
      trigger="hover"
      :show-after="500"
      :hide-after="100">
      <div class="site-actions">
        <el-button size="small" :icon="Position" @click="jumpLink(item.link)">{{ $t('sitenav.visit') }}</el-button>
        <el-button v-if="item.doc" size="small" :icon="Document" @click="jumpLink(item.doc)">{{ $t('sitenav.doc') }}</el-button>
        <el-button size="small" :icon="CopyDocument" @click="copyLink(item.link)">{{ $t('sitenav.copyURL') }}</el-button>
        <el-button size="small" :icon="StarFilled" @click="addBookmarks">{{ $t('sitenav.addBookmark') }}</el-button>
      </div>
      <template #reference>
        <div class="site-card" role="link" tabindex="0" @click="jumpLink(item.link)" @keyup.enter="jumpLink(item.link)">
          <span class="site-card__avatar">
            <img v-if="item.icon" :src="item.icon" loading="lazy" alt="">
            <span v-else>{{ (item.name || '?').slice(0, 1).toUpperCase() }}</span>
          </span>
          <div class="site-card__body">
            <div class="site-card__name">{{ item.name }}</div>
            <div class="site-card__desc">{{ item.desc || item.link }}</div>
            <div class="site-card__link">{{ item.link }}</div>
          </div>
        </div>
      </template>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { Position, Document, CopyDocument, StarFilled } from '@/icons/tabler';

defineProps({
  sites: {
    type: Array as () => any[],
    default: () => []
  }
});

const { t } = useI18n();

// Only http(s) links may be opened; stored links such as `javascript:` are rejected.
function safeHttpUrl(link: unknown): string | null {
  if (typeof link !== 'string' || !link.trim()) return null;
  try {
    const url = new URL(link.trim());
    return url.protocol === 'http:' || url.protocol === 'https:' ? url.href : null;
  } catch {
    return null;
  }
}

function jumpLink(link: unknown) {
  const url = safeHttpUrl(link);
  if (!url) {
    ElMessage.warning(t('sitenav.invalidLink'));
    return;
  }
  window.open(url, '_blank', 'noopener,noreferrer');
}

async function copyLink(link: string) {
  try {
    await navigator.clipboard.writeText(link);
    ElMessage({ message: t('tips.copySuccess'), type: 'success' });
  } catch (error) {
    ElMessage({ message: t('tips.copyFailed', [String(error)]), type: 'error' });
  }
}

function addBookmarks() {
  ElMessage({ message: t('tips.addBookmarkInfo'), type: 'info' });
}
</script>

<style lang="scss" scoped>
.site-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.site-card {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding: 14px 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--app-card-radius);
  background: var(--el-bg-color);
  box-shadow: var(--app-shadow);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s, transform 0.15s;

  &:hover,
  &:focus-visible {
    border-color: var(--el-color-primary-light-5);
    box-shadow: var(--app-shadow-hover);
    transform: translateY(-1px);
    outline: none;
  }
}

.site-card__avatar {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--el-fill-color-light);
  color: var(--el-color-primary);
  font-weight: 600;
  overflow: hidden;

  img {
    width: 28px;
    height: 28px;
    object-fit: contain;
  }
}

.site-card__body {
  min-width: 0;
  line-height: 1.4;
}

.site-card__name {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.site-card__name,
.site-card__desc,
.site-card__link {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.site-card__desc {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.site-card__link {
  font-size: 12px;
  color: var(--app-muted-text);
}

.site-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;

  .el-button {
    width: 100%;
    margin: 0;
    justify-content: flex-start;
  }
}
</style>
