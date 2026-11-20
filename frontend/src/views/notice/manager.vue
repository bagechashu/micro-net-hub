<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form :inline="true">
        <el-form-item>
          <el-button :loading="loading" :icon="Plus" type="primary" @click="addNotice">
            {{ $t('notice.6112860xxb80') }}
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            :icon="Delete"
            type="danger"
            @click="batchDeleteNotices"
          >
            {{ $t('notice.6112860xxug0') }}
          </el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-if="noticeData.length > 0"
        v-loading="loading"
        :data="noticeData"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="48" align="center" />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="level"
          :label="$t('notice.6112860xxyo0')"
          width="100"
          align="center"
        >
          <template #default="scope">
            <span :style="{ color: getLevelColor(scope.row.level), fontWeight: 'bold' }">
              {{ getLevelText(scope.row.level) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="content" :label="$t('notice.6112860xy1c0')" min-width="300" />
        <el-table-column show-overflow-tooltip sortable prop="creator" :label="$t('notice.6112860xy3w0')" width="120" />
        <el-table-column fixed="right" :label="$t('notice.6112860xy700')" align="left" width="160">
          <template #default="scope">
            <div class="row-actions">
              <el-button link type="primary" :icon="Edit" @click="updateNoticeRow(scope.row)">{{ $t('common.edit') }}</el-button>
              <el-popconfirm :title="$t('notice.6112860xyfk0')" @confirm="deleteNotice(scope.row.ID)">
                <template #reference>
                  <el-button link type="danger" :icon="Delete">{{ $t('common.delete') }}</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else :description="$t('common.nodata')" />

      <el-dialog v-model="noticeFormVisible" :title="noticeFormTitle" @closed="resetNoticeFormFields">
        <el-form ref="noticeFormRef" :model="noticeForm" :rules="noticeFormRules" label-position="top">
          <el-form-item :label="$t('notice.6112860xxyo0')" prop="level">
            <el-select v-model="noticeForm.level" :placeholder="$t('notice.6112860xyjw0')">
              <el-option v-for="item in levelOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('notice.6112860xy1c0')" prop="content">
            <el-input
              v-model="noticeForm.content"
              type="textarea"
              :placeholder="$t('notice.6112860xy1c0')"
              show-word-limit
              maxlength="100"
              @blur="noticeForm.content = (noticeForm.content || '').trim()"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <div class="dialog-footer">
            <el-button @click="noticeFormCancel()">{{ $t('notice.6112860xymw0') }}</el-button>
            <el-button :loading="loading" type="primary" @click="noticeFormSubmit()">
              {{ $t('notice.6112860xyq00') }}
            </el-button>
          </div>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance } from 'element-plus';
import { Plus, Delete, Edit } from '@/icons/tabler';
import { getNotice, createNotice, updateNotice, batchDeleteNoticeByIds } from '@/api/notice';

defineOptions({ name: 'NoticeManager' });

const { t } = useI18n();

const loading = ref(false);
const noticeData = ref<any[]>([]);

const noticeFormRef = ref<FormInstance>();
const noticeFormTitle = ref('');
const noticeFormType = ref('');
const noticeFormVisible = ref(false);
const emptyNoticeForm = () => ({ id: undefined as number | undefined, level: 1, content: '' });
const noticeForm = ref(emptyNoticeForm());
const noticeFormRules = {
  level: [{ required: true, message: t('notice.6112860xyus0'), trigger: 'blur' }],
  content: [
    { required: true, message: t('notice.6112860xyxk0'), trigger: 'blur' },
    { min: 10, max: 100, message: t('valid.length', [10, 100]), trigger: 'blur' }
  ]
};
const levelOptions = computed(() => [
  { name: t('notice.6112860xz0c0'), id: 1 },
  { name: t('notice.6112860xz380'), id: 2 },
  { name: t('notice.6112860xz5w0'), id: 3 },
  { name: t('notice.6112860xz8s0'), id: 4 }
]);
const multipleSelection = ref<any[]>([]);

async function getData() {
  loading.value = true;
  try {
    const { data } = await getNotice();
    noticeData.value = data || [];
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
}

function getLevelText(level: number) {
  const option = levelOptions.value.find(o => o.id === level);
  return option ? option.name : t('notice.6112860xzbk0');
}

function getLevelColor(level: number) {
  switch (level) {
    case 1:
      return '#999999'; // 一般
    case 2:
      return '#66bb6a'; // 普通
    case 3:
      return '#ffa726'; // 重要
    case 4:
      return '#e53935'; // 紧急
    default:
      return '#999999';
  }
}

function successMessage(res?: { msg?: string }) {
  ElMessage({ showClose: true, message: res?.msg || t('notice.6112860xzr40'), type: 'success' });
}

function cancelledMessage() {
  ElMessage({ showClose: true, type: 'info', message: t('notice.6112860xzn00') });
}

function addNotice() {
  noticeFormTitle.value = t('notice.6112860xze80');
  noticeFormType.value = 'add';
  noticeFormVisible.value = true;
}

function updateNoticeRow(row: any) {
  noticeForm.value = { id: row.ID, level: row.level, content: row.content };
  noticeFormTitle.value = t('notice.6112860xzhs0');
  noticeFormType.value = 'update';
  noticeFormVisible.value = true;
}

async function deleteNotice(id: number) {
  loading.value = true;
  try {
    successMessage(await batchDeleteNoticeByIds({ ids: [id] }));
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

async function batchDeleteNotices() {
  try {
    await ElMessageBox.confirm(t('tips.deleteWarning'), t('notice.6112860xzkg0'), {
      confirmButtonText: t('notice.6112860xyq00'),
      cancelButtonText: t('notice.6112860xymw0'),
      type: 'warning'
    });
  } catch {
    cancelledMessage();
    return;
  }
  loading.value = true;
  try {
    successMessage(await batchDeleteNoticeByIds({ ids: multipleSelection.value.map(x => x.ID) }));
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

async function noticeFormSubmit() {
  const valid = await noticeFormRef.value?.validate().catch(() => false);
  if (!valid) {
    ElMessage({ showClose: true, message: t('notice.6112860xzp00'), type: 'warning' });
    return;
  }
  loading.value = true;
  try {
    const payload = { ...noticeForm.value };
    if (noticeFormType.value === 'add') {
      delete payload.id;
      successMessage(await createNotice(payload));
    } else if (noticeFormType.value === 'update') {
      successMessage(await updateNotice(payload));
    }
  } catch {
    loading.value = false;
    return;
  }
  loading.value = false;
  noticeFormReset();
  getData();
}

function noticeFormCancel() {
  noticeFormReset();
}

function resetNoticeFormFields() {
  noticeFormRef.value?.resetFields();
  noticeForm.value = emptyNoticeForm();
}

function noticeFormReset() {
  noticeFormVisible.value = false;
  resetNoticeFormFields();
}

function handleSelectionChange(val: any[]) {
  multipleSelection.value = val;
}

onMounted(() => {
  getData();
});
</script>

<style scoped>
.container-card {
  margin: 10px;
  margin-bottom: 100px;
}

.delete-popover {
  margin-left: 10px;
}
</style>
