<template>
  <div>
    <el-row :gutter="0">
      <el-col
        v-for="(item, index) in sites"
        :key="index"
        :xs="12"
        :sm="8"
        :md="6"
        :lg="4"
        :xl="4"
        class="site-card"
      >
        <div @click="jumpLink(item.link)">
          <el-popover placement="top" :width="200" trigger="hover" :show-after="500" :hide-after="100">
            <el-button-group class="vertical">
              <el-button size="small" :icon="Position" @click="jumpLink(item.link)">
                {{ $t('sitenav.visit') }}
              </el-button>
              <el-button v-if="item.doc" size="small" :icon="Document" @click="jumpLink(item.doc)">
                {{ $t('sitenav.doc') }}
              </el-button>
              <el-button size="small" :icon="CopyDocument" class="clip-btn" @click="copyLink(item.link)">
                {{ $t('sitenav.copyURL') }}
              </el-button>
              <el-button size="small" :icon="StarFilled" @click="addBookmarks">
                {{ $t('sitenav.addBookmark') }}
              </el-button>
            </el-button-group>
            <template #reference>
              <el-card shadow="hover" :body-style="{ padding: '8px' }">
                <span><img :src="item.icon" loading="lazy" class="icon" alt=""></span>
                <span>{{ item.name }}</span>
                <p />
                <div class="desc">{{ item.desc }}</div>
                <div class="desclink">{{ item.link }}</div>
              </el-card>
            </template>
          </el-popover>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { Position, Document, CopyDocument, StarFilled } from '@element-plus/icons-vue';

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

<style lang="css" scoped>
.site-card {
  margin: 5px;
}
.el-button-group.vertical {
  display: flex;
  flex-direction: column;
}

.el-button-group.vertical .el-button + .el-button {
  margin-left: 0;
}

.el-button-group.vertical .el-button {
  width: 100%;
  margin-left: 0;
}

.icon {
  width: 30px;
  height: 30px;
  vertical-align: middle;
  border-radius: 50%;
  pointer-events: none;
  margin-right: 1rem;
}

.desc {
  padding-top: 5px;
  border-top: 1px solid var(--el-border-color-lighter);
  margin-top: 5px;
  font-size: 0.6rem;
  color: var(--app-muted-text);
}
.desclink {
  margin-top: 5px;
  font-size: 0.6rem;
  width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--app-muted-text);
}
</style>
