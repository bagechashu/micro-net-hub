<template>
  <div v-if="data.length" class="message-container">
    <IconSpeakerphone class="message-icon" :size="18" stroke="1.75" />
    <div
      ref="marqueeRef"
      class="marquee"
      @mouseover="paused = true"
      @mouseleave="paused = false"
    >
      <div
        class="marquee-track"
        :style="{
          animationDuration: `${duration}s`,
          animationPlayState: paused ? 'paused' : 'running',
          '--marquee-shift': `${-100 / copies}%`
        }"
      >
        <div v-for="n in copies" :key="n" ref="copyRefs" class="marquee-copy" :aria-hidden="n > 1 ? 'true' : undefined">
          <span v-for="item in data" :key="item.ID">
            <span class="message">
              <span :class="['level', getLevelColorClass(item.level)]">{{ getLevelText(item.level) }}:</span>
              <span class="content">{{ item.content }}</span>
            </span>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref, onBeforeUnmount } from 'vue';
import { IconSpeakerphone } from '@tabler/icons-vue';
import { useI18n } from 'vue-i18n';
import { getNotice } from '@/api/notice';

defineOptions({ name: 'NoticeBoard' });

const { t } = useI18n();

// 滚动速度（像素/秒）
const SPEED = 60;

const duration = ref(15);
const paused = ref(false);
const copies = ref(2);
const data = ref<any[]>([]);
const marqueeRef = ref<HTMLElement>();
const copyRefs = ref<HTMLElement[]>([]);

// 动画每次平移一份内容的宽度：按实际宽度换算时长，保证匀速；份数要足够铺满一行，循环才连贯
async function updateMarquee() {
  copies.value = 2;
  await nextTick();
  const copyWidth = copyRefs.value[0]?.offsetWidth || 0;
  const boxWidth = marqueeRef.value?.offsetWidth || 0;
  if (!copyWidth) return;
  copies.value = Math.max(2, Math.ceil(boxWidth / copyWidth) + 1);
  duration.value = Math.max(4, copyWidth / SPEED);
}

async function fetchNotice() {
  try {
    const res = await getNotice();
    const list = Array.isArray(res.data) ? res.data : [];
    data.value = list;
    if (list.length) await updateMarquee();
  } catch {
    // error already shown by request wrapper
  }
}

function getLevelText(level: number) {
  switch (level) {
    case 1:
      return t('notice.6112860xz0c0');
    case 2:
      return t('notice.6112860xz380');
    case 3:
      return t('notice.6112860xz5w0');
    case 4:
      return t('notice.6112860xz8s0');
    default:
      return t('notice.61125ywdxyk0');
  }
}

function getLevelColorClass(level: number) {
  switch (level) {
    case 2:
      return 'info-level';
    case 3:
      return 'warning-level';
    case 4:
      return 'critical-level';
    default:
      return 'default-level';
  }
}

fetchNotice();
const timer = setInterval(fetchNotice, 60 * 60 * 1000); // refresh every 60 minutes

onBeforeUnmount(() => {
  clearInterval(timer);
});
</script>

<style lang="css" scoped>
.message-container {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 40px;
  padding: 0 14px;
  border: 1px solid var(--el-border-color-light);
  border-radius: var(--app-card-radius);
  background: var(--el-bg-color);
  box-shadow: var(--app-shadow);
  font-size: 13px;
}

.message-icon {
  flex: none;
  color: var(--el-color-primary);
}

.marquee {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.marquee-track {
  display: flex;
  width: max-content;
  animation-name: marquee-scroll;
  animation-timing-function: linear;
  animation-iteration-count: infinite;
}

.marquee-copy {
  flex-shrink: 0;
  white-space: nowrap;
}

@keyframes marquee-scroll {
  from {
    transform: translateX(0);
  }
  to {
    transform: translateX(var(--marquee-shift));
  }
}

@media (prefers-reduced-motion: reduce) {
  .marquee-track {
    animation-play-state: paused !important;
  }
}

.message {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-right: 32px;
  color: var(--el-text-color-regular);
}

/* Tabler 风格的浅色徽标 */
.level {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.info-level {
  background-color: var(--el-color-success-light-9);
  color: var(--el-color-success);
}

.warning-level {
  background-color: var(--el-color-warning-light-9);
  color: var(--el-color-warning);
}

.critical-level {
  background-color: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}

.default-level {
  background-color: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
}

.content {
  white-space: pre;
}
</style>
