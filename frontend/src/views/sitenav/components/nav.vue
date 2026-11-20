<template>
  <div class="site-grid">
    <el-popover
      v-for="(item, index) in sites"
      :key="item.ID ?? index"
      placement="top"
      :width="200"
      popper-class="site-actions-popper"
      trigger="hover"
      :show-after="500"
      :hide-after="100">
      <div class="site-actions">
        <el-button text :icon="Position" @click="jumpLink(item.link)">{{ $t('sitenav.visit') }}</el-button>
        <el-button text v-if="item.doc" :icon="Document" @click="jumpLink(item.doc)">{{ $t('sitenav.doc') }}</el-button>
        <el-button text :icon="CopyDocument" @click="copyLink(item.link)">{{ $t('sitenav.copyURL') }}</el-button>
        <el-button text :icon="isBookmarked(item.ID) ? IconStarOff : StarFilled" @click="onToggleBookmark(item)">
          {{ isBookmarked(item.ID) ? $t('sitenav.removeBookmark') : $t('sitenav.addBookmark') }}
        </el-button>
      </div>
      <template #reference>
        <div
          class="site-card"
          :class="{ 'is-bookmarked': isBookmarked(item.ID) }"
          role="link"
          tabindex="0"
          @click="jumpLink(item.link)"
          @keyup.enter="jumpLink(item.link)">
          <SiteIcon :value="item.icon" :name="item.name" :link="item.link" :size="40" />
          <div class="site-card__body">
            <div class="site-card__name">{{ item.name }}</div>
            <div class="site-card__desc">{{ item.desc || item.link }}</div>
            <div class="site-card__link">{{ item.link }}</div>
          </div>
          <!-- 已收藏标记 -->
          <IconStarFilled v-if="isBookmarked(item.ID)" class="site-card__star" :size="14" />
        </div>
      </template>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { Position, Document, CopyDocument, StarFilled } from '@/icons/tabler';
import { IconStarFilled, IconStarOff } from '@tabler/icons-vue';
import { useBookmarks } from '../bookmarks';
import SiteIcon from '@/components/SiteIcon/index.vue';

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

// 收藏只保存在当前浏览器（localStorage），收藏的站点显示在导航页最上方
const { isBookmarked, toggleBookmark } = useBookmarks();

function onToggleBookmark(item: { ID?: number; name?: string }) {
  const added = toggleBookmark(item.ID);
  ElMessage({
    message: added ? t('sitenav.bookmarkAdded', [item.name]) : t('sitenav.bookmarkRemoved', [item.name]),
    type: added ? 'success' : 'info'
  });
}
</script>

<style lang="scss" scoped>
.site-card.is-bookmarked {
  padding-right: 28px;
}

.site-card__star {
  position: absolute;
  top: 8px;
  right: 8px;
  color: var(--el-color-warning);
}

.site-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.site-card {
  position: relative;
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

// 悬停菜单：Tabler dropdown 风格，整行可点、悬停高亮
.site-actions {
  display: flex;
  flex-direction: column;
  gap: 2px;

  .el-button {
    width: 100%;
    height: 34px;
    margin: 0;
    padding: 0 10px;
    justify-content: flex-start;
    font-weight: 400;
    color: var(--el-text-color-regular);

    &:hover {
      color: var(--el-color-primary);
      background: var(--el-color-primary-light-9);
    }
  }
}
</style>

<style lang="scss">
.el-popover.site-actions-popper {
  padding: 6px;
}
</style>
