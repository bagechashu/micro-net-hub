<template>
  <!-- 站点图标：品牌 Logo / 通用图标 / favicon / 图片 / 首字头像，统一为圆角方块 -->
  <span
    class="site-icon"
    :class="{ 'is-dark-brand': darkBrand, 'is-image': !!imgSrc && !brand && !generic }"
    :style="{ width: `${size}px`, height: `${size}px`, '--site-icon-color': tint, borderRadius: `${Math.round(size / 5)}px` }"
  >
    <svg
      v-if="brand"
      class="site-icon__svg"
      viewBox="0 0 24 24"
      :width="glyph"
      :height="glyph"
      role="img"
      :aria-label="brand.title"
    >
      <path :d="brand.path" />
    </svg>
    <component :is="generic.component" v-else-if="generic" :size="glyph" stroke="1.75" />
    <img v-else-if="imgSrc" :src="imgSrc" :width="glyph" :height="glyph" alt="" loading="lazy" @error="imgFailed = true">
    <span v-else class="site-icon__initial" :style="{ fontSize: `${Math.round(size * 0.42)}px` }">{{ initial }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed, ref, shallowRef, watch } from 'vue';
import type { BrandIcon } from '@/icons/brands';
import {
  parseSiteIcon,
  genericIconMap,
  ICON_COLORS,
  faviconOf,
  initialOf,
  colorOfName,
  isDarkColor,
  isSafeImageUrl,
  loadBrands
} from './icon';

defineOptions({ name: 'SiteIcon' });

const props = withDefaults(
  defineProps<{
    /** icon 字段原值 */
    value?: string;
    /** 站点名称（首字头像用） */
    name?: string;
    /** 站点链接（favicon 用） */
    link?: string;
    size?: number;
  }>(),
  { value: '', name: '', link: '', size: 40 }
);

const parsed = computed(() => parseSiteIcon(props.value));
const glyph = computed(() => Math.round(props.size * 0.6));

// 品牌图标异步加载
const brand = shallowRef<BrandIcon | null>(null);
watch(
  parsed,
  async p => {
    brand.value = null;
    if (p.kind !== 'brand') return;
    const { brandIconMap } = await loadBrands();
    if (parseSiteIcon(props.value).kind === 'brand') brand.value = brandIconMap[p.slug] || null;
  },
  { immediate: true }
);

const generic = computed(() => (parsed.value.kind === 'tabler' ? genericIconMap[parsed.value.name] : undefined));

// 图片 / favicon：加载失败时退回首字头像
const imgFailed = ref(false);
watch(() => [props.value, props.link], () => (imgFailed.value = false));
const imgSrc = computed(() => {
  if (imgFailed.value) return '';
  const p = parsed.value;
  if (p.kind === 'image') return isSafeImageUrl(p.url) ? p.url : '';
  if (p.kind === 'favicon') return faviconOf(props.link);
  return '';
});

const initial = computed(() => initialOf(props.name));

// 图标颜色：品牌色 / 选定颜色 / 按名称固定的颜色
const tint = computed(() => {
  if (brand.value) return `#${brand.value.hex}`;
  if (parsed.value.kind === 'tabler') return ICON_COLORS[parsed.value.color];
  return colorOfName(props.name);
});
const darkBrand = computed(() => !!brand.value && isDarkColor(brand.value.hex));
</script>

<style lang="scss" scoped>
.site-icon {
  display: inline-flex;
  flex: none;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  color: var(--site-icon-color);
  background: color-mix(in srgb, var(--site-icon-color) 12%, transparent);

  img {
    object-fit: contain;
  }

  &.is-image {
    background: var(--el-fill-color-light);
  }
}

.site-icon__svg {
  fill: var(--site-icon-color);
}

.site-icon__initial {
  font-weight: 600;
  line-height: 1;
}

// 很暗的品牌色（如 GitHub 黑色）在暗色模式下反白，避免看不见
html.dark .site-icon.is-dark-brand {
  --site-icon-color: #e6e8eb;
}
</style>
