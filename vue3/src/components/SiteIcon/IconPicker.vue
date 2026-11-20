<template>
  <div class="icon-picker">
    <el-popover v-model:visible="open" placement="bottom-start" :width="480" :fallback-placements="['top-start', 'bottom-start', 'right']" trigger="click" popper-class="icon-picker__popper">
      <template #reference>
        <button type="button" class="icon-picker__trigger">
          <SiteIcon :value="modelValue" :name="name" :link="link" :size="36" />
          <span class="icon-picker__summary">
            <span class="icon-picker__kind">{{ summary.kind }}</span>
            <span class="icon-picker__detail">{{ summary.detail }}</span>
          </span>
          <IconChevronDown :size="16" stroke="1.75" class="icon-picker__arrow" />
        </button>
      </template>

      <el-tabs v-model="tab" class="icon-picker__tabs">
        <!-- 品牌 Logo -->
        <el-tab-pane :label="$t('sitenav.iconPicker.brand')" name="brand">
          <el-input v-model.trim="brandQuery" :prefix-icon="Search" clearable :placeholder="$t('sitenav.iconPicker.searchBrand')" />
          <div v-loading="!brands.length" class="icon-picker__grid">
            <el-tooltip v-for="b in filteredBrands" :key="b.slug" :content="b.title" placement="top" :show-after="300">
              <button
                type="button"
                class="icon-picker__cell"
                :class="{ 'is-active': modelValue === `si:${b.slug}` }"
                @click="pick(`si:${b.slug}`)"
              >
                <SiteIcon :value="`si:${b.slug}`" :size="36" />
              </button>
            </el-tooltip>
            <div v-if="brands.length && !filteredBrands.length" class="icon-picker__empty">
              {{ $t('sitenav.iconPicker.noBrand') }}
            </div>
          </div>
        </el-tab-pane>

        <!-- 通用图标 + 颜色 -->
        <el-tab-pane :label="$t('sitenav.iconPicker.generic')" name="generic">
          <el-input v-model.trim="genericQuery" :prefix-icon="Search" clearable :placeholder="$t('sitenav.iconPicker.searchGeneric')" />
          <div class="icon-picker__colors">
            <button
              v-for="(hex, key) in ICON_COLORS"
              :key="key"
              type="button"
              class="icon-picker__swatch"
              :class="{ 'is-active': color === key }"
              :style="{ background: hex }"
              :title="key"
              @click="setColor(key)"
            />
          </div>
          <div class="icon-picker__grid">
            <el-tooltip v-for="g in filteredGeneric" :key="g.name" :content="g.name" placement="top" :show-after="300">
              <button
                type="button"
                class="icon-picker__cell"
                :class="{ 'is-active': current.kind === 'tabler' && current.name === g.name }"
                @click="pick(formatSiteIcon({ kind: 'tabler', name: g.name, color }))"
              >
                <SiteIcon :value="formatSiteIcon({ kind: 'tabler', name: g.name, color })" :size="36" />
              </button>
            </el-tooltip>
          </div>
        </el-tab-pane>

        <!-- 网站自己的 favicon -->
        <el-tab-pane :label="$t('sitenav.iconPicker.favicon')" name="favicon">
          <div class="icon-picker__option">
            <SiteIcon value="favicon" :name="name" :link="link" :size="48" />
            <div>
              <p>{{ $t('sitenav.iconPicker.faviconDesc') }}</p>
              <p v-if="faviconUrl" class="icon-picker__muted">{{ faviconUrl }}</p>
              <p v-else class="icon-picker__muted">{{ $t('sitenav.iconPicker.faviconNeedLink') }}</p>
              <el-button type="primary" plain @click="pick('favicon')">{{ $t('sitenav.iconPicker.use') }}</el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- 图片地址（http/https） -->
        <el-tab-pane :label="$t('sitenav.iconPicker.image')" name="image">
          <div class="icon-picker__url">
            <el-input v-model.trim="imageUrl" clearable :placeholder="$t('sitenav.iconPicker.imagePlaceholder')" />
            <el-button type="primary" plain :disabled="!imageUrlValid" @click="pick(imageUrl)">{{ $t('sitenav.iconPicker.use') }}</el-button>
          </div>
          <p class="icon-picker__muted">{{ $t('sitenav.iconPicker.imageDesc') }}</p>
        </el-tab-pane>
      </el-tabs>

      <div class="icon-picker__footer">
        <el-button link @click="pick('')">{{ $t('sitenav.iconPicker.clear') }}</el-button>
        <span class="icon-picker__muted">{{ $t('sitenav.iconPicker.clearDesc') }}</span>
      </div>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, shallowRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { IconChevronDown } from '@tabler/icons-vue';
import { Search } from '@/icons/tabler';
import type { BrandIcon } from '@/icons/brands';
import SiteIcon from './index.vue';
import {
  ICON_COLORS,
  DEFAULT_ICON_COLOR,
  genericIcons,
  genericIconMap,
  parseSiteIcon,
  formatSiteIcon,
  faviconOf,
  isSafeImageUrl,
  loadBrands
} from './icon';

defineOptions({ name: 'IconPicker' });

const props = withDefaults(
  defineProps<{
    modelValue?: string;
    name?: string;
    link?: string;
  }>(),
  { modelValue: '', name: '', link: '' }
);
const emit = defineEmits<{ 'update:modelValue': [value: string] }>();

const { t } = useI18n();
const open = ref(false);
const tab = ref('brand');
const current = computed(() => parseSiteIcon(props.modelValue));

// 打开时定位到当前值所在的页；还没有有效图标时，按站点名称预先搜索品牌（如 jenkins）
watch(open, async v => {
  if (!v) return;
  const c = current.value;
  const validImage = c.kind === 'image' && isSafeImageUrl(c.url);
  tab.value = c.kind === 'tabler' ? 'generic' : c.kind === 'favicon' ? 'favicon' : validImage ? 'image' : 'brand';
  imageUrl.value = validImage ? (c as { url: string }).url : '';
  if (c.kind === 'tabler') color.value = c.color;
  const m = await loadBrands();
  brands.value = m.brandIcons;
  if (tab.value === 'brand' && c.kind !== 'brand' && !brandQuery.value) {
    const q = props.name.trim().toLowerCase();
    if (q && m.brandIcons.some(b => b.slug.includes(q) || b.title.toLowerCase().includes(q))) brandQuery.value = q;
  }
});

function pick(value: string) {
  emit('update:modelValue', value);
  open.value = false;
}

// 品牌
const brands = shallowRef<BrandIcon[]>([]);
const brandQuery = ref('');
const filteredBrands = computed(() => {
  const q = brandQuery.value.toLowerCase();
  return q ? brands.value.filter(b => b.title.toLowerCase().includes(q) || b.slug.includes(q)) : brands.value;
});

// 通用图标
const genericQuery = ref('');
const color = ref(DEFAULT_ICON_COLOR);
const filteredGeneric = computed(() => {
  const q = genericQuery.value.toLowerCase();
  return q ? genericIcons.filter(g => g.name.includes(q) || g.keywords.toLowerCase().includes(q)) : genericIcons;
});
// 已选通用图标时改颜色直接生效
function setColor(key: string) {
  color.value = key;
  if (current.value.kind === 'tabler') {
    emit('update:modelValue', formatSiteIcon({ kind: 'tabler', name: current.value.name, color: key }));
  }
}

// favicon / 图片
const faviconUrl = computed(() => faviconOf(props.link));
const imageUrl = ref('');
const imageUrlValid = computed(() => isSafeImageUrl(imageUrl.value) && imageUrl.value.length <= 100);

// 触发器上的文字说明
const brandTitle = ref('');
watch(
  current,
  async c => {
    brandTitle.value = '';
    if (c.kind === 'brand') {
      const { brandIconMap } = await loadBrands();
      brandTitle.value = brandIconMap[c.slug]?.title || c.slug;
    }
  },
  { immediate: true }
);
const summary = computed(() => {
  const c = current.value;
  switch (c.kind) {
    case 'brand':
      return { kind: t('sitenav.iconPicker.brand'), detail: brandTitle.value || c.slug };
    case 'tabler':
      return { kind: t('sitenav.iconPicker.generic'), detail: genericIconMap[c.name] ? c.name : `${c.name} ?` };
    case 'favicon':
      return { kind: t('sitenav.iconPicker.favicon'), detail: faviconUrl.value || t('sitenav.iconPicker.faviconNeedLink') };
    case 'image':
      return { kind: t('sitenav.iconPicker.image'), detail: c.url };
    default:
      return { kind: t('sitenav.iconPicker.initial'), detail: t('sitenav.iconPicker.initialDesc') };
  }
});
</script>

<style lang="scss" scoped>
.icon-picker {
  width: 100%;
}

.icon-picker__trigger {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 6px 12px 6px 6px;
  font: inherit;
  text-align: left;
  color: inherit;
  cursor: pointer;
  background: var(--el-fill-color-blank);
  border: 0;
  border-radius: var(--el-border-radius-base);
  box-shadow: 0 0 0 1px var(--el-border-color) inset;
  transition: box-shadow 0.2s;

  &:hover {
    box-shadow: 0 0 0 1px var(--el-border-color-hover) inset;
  }

  &:focus-visible {
    outline: none;
    box-shadow: 0 0 0 1px var(--el-color-primary) inset;
  }
}

.icon-picker__summary {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  line-height: 1.4;
}

.icon-picker__kind {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.icon-picker__detail {
  overflow: hidden;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.icon-picker__arrow {
  flex: none;
  color: var(--el-text-color-secondary);
}

.icon-picker__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(44px, 1fr));
  gap: 6px;
  min-height: 60px;
  max-height: 200px;
  margin-top: 10px;
  overflow-y: auto;
}

.icon-picker__cell {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 44px;
  padding: 0;
  cursor: pointer;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 8px;

  &:hover {
    background: var(--el-fill-color-light);
  }

  &.is-active {
    border-color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
  }
}

.icon-picker__empty {
  grid-column: 1 / -1;
  padding: 16px 0;
  font-size: 13px;
  text-align: center;
  color: var(--el-text-color-secondary);
}

.icon-picker__colors {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

.icon-picker__swatch {
  width: 20px;
  height: 20px;
  padding: 0;
  cursor: pointer;
  border: 2px solid transparent;
  border-radius: 50%;
  box-shadow: 0 0 0 1px var(--el-border-color-lighter);

  &.is-active {
    border-color: var(--el-bg-color);
    box-shadow: 0 0 0 2px var(--el-text-color-primary);
  }
}

.icon-picker__option {
  display: flex;
  gap: 16px;
  padding: 8px 0;

  p {
    margin: 0 0 8px;
    font-size: 13px;
  }
}

.icon-picker__url {
  display: flex;
  gap: 8px;
}


.icon-picker__muted {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
}

.icon-picker__footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 10px;
  margin-top: 10px;
  border-top: 1px solid var(--el-border-color-lighter);

  .icon-picker__muted {
    margin: 0;
  }
}
</style>

<style lang="scss">
.el-popover.icon-picker__popper {
  padding: 12px 14px;

  .el-tabs__header {
    margin-bottom: 12px;
  }
}
</style>
