<template>
  <!-- 表格展开行的详情：固定列数，按顺序交错排列；每项“标签 | 值”横排 -->
  <dl class="row-detail" :class="{ 'row-detail--wide': columns > 2 }" :style="{ '--row-detail-columns': columns }">
    <div v-for="item in items" :key="item.label" class="row-detail__item" :class="{ 'is-wide': item.wide }">
      <dt>{{ item.label }}</dt>
      <dd :class="{ 'is-empty': isEmpty(item.value) }">
        <slot :name="item.slot || ''" :item="item">{{ isEmpty(item.value) ? '—' : item.value }}</slot>
      </dd>
    </div>
  </dl>
</template>

<script setup lang="ts">
export interface RowDetailItem {
  label: string;
  value?: unknown;
  /** 占满整行（较长的内容，如 DN、简介） */
  wide?: boolean;
  /** 自定义渲染的插槽名 */
  slot?: string;
}

defineOptions({ name: 'RowDetail' });

withDefaults(defineProps<{
  items: RowDetailItem[];
  /** 列数，默认 2 */
  columns?: number;
}>(), { columns: 2 });

function isEmpty(v: unknown) {
  return v === undefined || v === null || v === '';
}
</script>

<style lang="scss" scoped>
// 固定两列，按顺序左右交错排列；每项“标签 | 值”横排，标签列等宽对齐
.row-detail {
  display: grid;
  grid-template-columns: repeat(var(--row-detail-columns, 2), minmax(0, 1fr));
  gap: 10px 40px;
  margin: 0;
  padding: 16px 20px 16px 104px;
}

.row-detail__item {
  display: grid;
  grid-template-columns: 112px minmax(0, 1fr);
  gap: 12px;
  align-items: baseline;
  min-width: 0;

  &.is-wide {
    grid-column: 1 / -1;
  }

  dt {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  dd {
    margin: 0;
    font-size: 13px;
    color: var(--el-text-color-primary);
    word-break: break-all;
    white-space: pre-line;

    &.is-empty {
      color: var(--el-text-color-placeholder);
    }
  }
}

// 3 列及以上：放宽最大宽度；屏幕不够宽时退回 2 列
.row-detail {
  max-width: 960px;

  &.row-detail--wide {
    max-width: 1280px;
  }
}

@media (max-width: 1200px) {
  .row-detail.row-detail--wide {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .row-detail {
    grid-template-columns: 1fr;
    padding-left: 20px;
  }
}
</style>
