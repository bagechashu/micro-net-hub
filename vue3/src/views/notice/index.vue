<template>
  <div class="message-container">
    <div
      class="marquee"
      @mouseover="paused = true"
      @mouseleave="paused = false"
    >
      <div
        class="marquee-track"
        :style="{
          animationDuration: `${duration * copies}s`,
          animationPlayState: paused ? 'paused' : 'running',
          '--marquee-shift': `${-100 / copies}%`
        }"
      >
        <div v-for="n in copies" :key="n" class="marquee-copy" :aria-hidden="n > 1 ? 'true' : undefined">
          <span v-for="(item, i) in data" :key="item.ID">
            <span class="message" :style="styles[i]">
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
import { ref, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { getNotice } from '@/api/notice';

defineOptions({ name: 'NoticeBoard' });

const { t } = useI18n();

const duration = ref(15);
const paused = ref(false);
const copies = ref(4);
const data = ref<any[]>([]);
// Random font styles are generated once per fetch so repeated copies stay identical (seamless loop).
const styles = ref<Record<string, string>[]>([]);

function getRandomFontSize() {
  const minSize = 12;
  const maxSize = 20;
  return `${Math.floor(Math.random() * (maxSize - minSize + 1)) + minSize}px`;
}

function getRandomWeight() {
  return Math.random() > 0.5 ? 'bold' : 'normal';
}

async function fetchNotice() {
  try {
    const res = await getNotice();
    const list = Array.isArray(res.data) ? res.data : [];
    data.value = list;
    styles.value = list.map(() => ({ fontSize: getRandomFontSize(), fontWeight: getRandomWeight() }));
    duration.value = Math.max(5, Math.round(JSON.stringify(list).length / 30));
    copies.value = Math.max(2, 10 - list.length);
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
  flex-direction: column;
  height: fit-content;
  padding: 8px;
}

.marquee {
  overflow: hidden;
  width: 100%;
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
  align-items: center;
  padding: 2px 10px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.info-level {
  background-color: #66bb6a;
  color: white;
}

.warning-level {
  background-color: #ffa726;
  color: white;
}

.critical-level {
  background-color: #e53935;
  color: white;
}

.default-level {
  background-color: #999999;
  color: white;
}

.level {
  margin-left: 5px;
  padding: 1px 8px;
  border-radius: 4px;
  font-weight: bold;
}

.content {
  white-space: pre;
}
</style>
