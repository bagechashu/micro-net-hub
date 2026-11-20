<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        ref="dnsZoneFormRef"
        :inline="true"
        size="small"
        :model="dnsZoneForm"
        :rules="dnsZoneFormRules"
      >
        <el-form-item label="Zone:" prop="name">
          <el-input v-model.trim="dnsZoneForm.name" placeholder="eg: example.com" />
        </el-form-item>
        <el-form-item>
          <el-button size="small" :loading="loading" type="primary" @click="addZone()">
            {{ $t('dns.6112hr284k00') }} Zone
          </el-button>
        </el-form-item>
      </el-form>
      <el-tabs
        v-if="dnsData.length > 0"
        v-model="dnsZoneActiveTab"
        type="border-card"
        closable
        @tab-remove="deleteZone"
      >
        <el-tab-pane v-for="item in dnsData" :key="item.ID" :label="item.name" :name="item.name">
          <el-form size="small" :inline="true" class="demo-form-inline">
            <el-form-item>
              <el-button :loading="loading" :icon="Plus" type="warning" @click="addRecord">
                {{ $t('dns.6112hr2851s0') }}
              </el-button>
            </el-form-item>
            <el-form-item>
              <el-button
                :disabled="multipleSelection.length === 0"
                :loading="loading"
                :icon="Delete"
                type="danger"
                @click="batchDeleteRecords"
              >
                {{ $t('dns.6112hr2856c0') }}
              </el-button>
            </el-form-item>
          </el-form>

          <el-table
            v-loading="loading"
            :data="item.records"
            border
            stripe
            size="small"
            style="width: 100%"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="55" align="center" />
            <el-table-column show-overflow-tooltip sortable prop="type" label="Type" />
            <el-table-column show-overflow-tooltip sortable prop="host" label="Host" />
            <el-table-column show-overflow-tooltip sortable prop="value" label="Value" />
            <el-table-column show-overflow-tooltip sortable prop="ttl" label="TTL" />
            <el-table-column show-overflow-tooltip sortable prop="creator" :label="$t('dns.6112hr285980')" />
            <el-table-column fixed="right" :label="$t('dns.6112hr285c40')" align="center" width="120">
              <template #default="scope">
                <el-tooltip :content="$t('dns.6112hr285ew0')" effect="dark" placement="top">
                  <el-button size="small" :icon="Edit" circle type="primary" @click="updateRecord(scope.row)" />
                </el-tooltip>
                <el-tooltip class="delete-popover" :content="$t('dns.6112hr285hs0')" effect="dark" placement="top">
                  <template #default>
                    <el-popconfirm :title="$t('dns.6112hr285kk0')" @confirm="deleteRecord(scope.row.ID)">
                      <template #reference>
                        <el-button size="small" :icon="Delete" circle type="danger" />
                      </template>
                    </el-popconfirm>
                  </template>
                </el-tooltip>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
      <el-empty v-else :description="$t('common.nodata')" />

      <el-dialog v-model="dnsRecordFormVisible" :title="dnsRecordFormTitle" @closed="resetRecordFormFields">
        <el-form
          ref="dnsRecordFormRef"
          size="small"
          :model="dnsRecordForm"
          :rules="dnsRecordFormRules"
          label-width="auto"
        >
          <el-form-item label="Zone" prop="zone_id">
            <el-select v-model="dnsRecordForm.zone_id" placeholder="choose one">
              <el-option v-for="item in zoneOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="Type" prop="type">
            <el-select v-model="dnsRecordForm.type" filterable default-first-option placeholder="choose one">
              <el-option v-for="item in typeOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="Host" prop="host">
            <el-input v-model.trim="dnsRecordForm.host" placeholder="eg: www/@" />
          </el-form-item>
          <el-form-item label="value" prop="value">
            <el-input v-model.trim="dnsRecordForm.value" placeholder="value" />
          </el-form-item>
          <el-form-item label="TTL" prop="ttl">
            <el-input v-model.number="dnsRecordForm.ttl" :placeholder="$t('dns.6112hr285n40')" />
          </el-form-item>
        </el-form>
        <template #footer>
          <div class="dialog-footer">
            <el-button size="small" @click="dnsRecordFormCancel()">{{ $t('dns.6112hr285q00') }}</el-button>
            <el-button size="small" :loading="loading" type="primary" @click="dnsRecordFormSubmit()">
              {{ $t('dns.6112hr285so0') }}
            </el-button>
          </div>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance } from 'element-plus';
import { Plus, Delete, Edit } from '@element-plus/icons-vue';
import {
  getDnsRecords,
  createDnsZone,
  batchDeleteDnsZoneByIds,
  createDnsRecord,
  updateDnsRecord,
  batchDeleteDnsRecordByIds
} from '@/api/dns';

defineOptions({ name: 'DnsManager' });

const { t } = useI18n();

// Ported from Vue2 utils/validate.js validateSecondLevelDomain (not present in vue3 utils/validate.ts).
function validateSecondLevelDomain(_rule: unknown, value: unknown, callback: (e?: Error) => void) {
  if (!value || typeof value !== 'string' || value.trim() === '') {
    return callback(new Error(t('valid.notAllowEmpty')));
  }
  const domainRegex = /^[a-zA-Z0-9-_]+\.[a-zA-Z]{2,}$/;
  if (!domainRegex.test(value)) {
    callback(new Error(t('valid.invalidDnsZone')));
  } else {
    callback();
  }
}

const loading = ref(false);
const dnsData = ref<any[]>([]);

const dnsZoneFormRef = ref<FormInstance>();
const dnsZoneForm = reactive({ name: '' });
const dnsZoneFormRules = {
  name: [
    { required: true, message: t('dns.inputZoneNameTips'), trigger: 'blur' },
    { min: 4, max: 20, message: t('valid.length', [4, 20]), trigger: 'blur' },
    { required: true, validator: validateSecondLevelDomain, trigger: 'blur' }
  ]
};
const dnsZoneActiveTab = ref<string | number>('');

const dnsRecordFormRef = ref<FormInstance>();
const dnsRecordFormTitle = ref('');
const dnsRecordFormType = ref('');
const dnsRecordFormVisible = ref(false);
const emptyRecordForm = () => ({
  id: '' as number | string,
  zone_id: '' as number | string | undefined,
  type: '',
  host: '',
  value: '',
  ttl: '' as number | string
});
const dnsRecordForm = ref(emptyRecordForm());
const dnsRecordFormRules = {
  zone_id: [{ required: true, message: t('dns.6112hr285v80'), trigger: 'blur' }],
  type: [{ required: true, message: t('dns.6112hr285xs0'), trigger: 'blur' }],
  host: [
    { required: true, message: t('dns.inputHostNameTips'), trigger: 'blur' },
    { min: 1, max: 30, message: t('valid.length', [1, 30]), trigger: 'blur' }
  ],
  value: [
    { required: true, message: t('dns.6112hr2860o0'), trigger: 'change' },
    { min: 1, max: 100, message: t('valid.length', [1, 100]), trigger: 'blur' }
  ],
  ttl: [
    { required: true, message: t('dns.inputTtlTips'), trigger: 'blur' },
    { type: 'number' as const, message: 'must be number', trigger: 'blur' }
  ]
};
const typeOptions = ['A', 'CNAME', 'TXT'];
const zoneOptions = ref<{ id: number; name: string }[]>([]);
const multipleSelection = ref<any[]>([]);

async function getData() {
  loading.value = true;
  try {
    const { data } = await getDnsRecords();
    dnsData.value = data || [];
    if ((dnsZoneActiveTab.value === '0' || dnsZoneActiveTab.value === '') && dnsData.value.length > 0) {
      dnsZoneActiveTab.value = dnsData.value[0].name;
    }
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
}

function getZoneOptions() {
  zoneOptions.value = dnsData.value.map(item => ({ id: item.ID, name: item.name }));
}

function successMessage() {
  ElMessage({ showClose: true, message: t('dns.6112hr286i40'), type: 'success' });
}

function validateFailedMessage() {
  ElMessage({ showClose: true, message: t('dns.6112hr2864k0'), type: 'warning' });
}

function cancelledMessage() {
  ElMessage({ showClose: true, type: 'info', message: t('dns.6112hr2869s0') });
}

async function addZone() {
  const valid = await dnsZoneFormRef.value?.validate().catch(() => false);
  if (!valid) {
    validateFailedMessage();
    return;
  }
  loading.value = true;
  try {
    await createDnsZone({ ...dnsZoneForm });
    successMessage();
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

function getZoneIDFromTabname(tabname: string | number) {
  return dnsData.value.find(item => item.name === tabname)?.ID;
}

async function deleteZone(tabname: string | number) {
  const dnsZoneId = getZoneIDFromTabname(tabname);
  try {
    await ElMessageBox.confirm(t('dns.deleteZoneTips'), t('dns.6112hr286740'), {
      confirmButtonText: t('dns.6112hr285so0'),
      cancelButtonText: t('dns.6112hr285q00'),
      type: 'warning'
    });
  } catch {
    cancelledMessage();
    return;
  }
  loading.value = true;
  try {
    await batchDeleteDnsZoneByIds({ ids: [dnsZoneId] });
    successMessage();
    dnsZoneActiveTab.value = '0';
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

function addRecord() {
  getZoneOptions();
  dnsRecordForm.value.zone_id = getZoneIDFromTabname(dnsZoneActiveTab.value);
  dnsRecordFormTitle.value = t('dns.6112hr286cc0');
  dnsRecordFormType.value = 'add';
  dnsRecordFormVisible.value = true;
}

function updateRecord(row: any) {
  getZoneOptions();
  dnsRecordForm.value = {
    id: row.ID,
    zone_id: row.zone_id,
    type: row.type,
    host: row.host,
    value: row.value,
    ttl: row.ttl
  };
  dnsRecordFormTitle.value = t('dns.6112hr286f40');
  dnsRecordFormType.value = 'update';
  dnsRecordFormVisible.value = true;
}

async function deleteRecord(id: number) {
  loading.value = true;
  try {
    await batchDeleteDnsRecordByIds({ ids: [id] });
    successMessage();
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

async function batchDeleteRecords() {
  try {
    await ElMessageBox.confirm(t('tips.deleteWarning'), t('dns.6112hr286740'), {
      confirmButtonText: t('dns.6112hr285so0'),
      cancelButtonText: t('dns.6112hr285q00'),
      type: 'warning'
    });
  } catch {
    cancelledMessage();
    return;
  }
  loading.value = true;
  try {
    await batchDeleteDnsRecordByIds({ ids: multipleSelection.value.map(x => x.ID) });
    successMessage();
  } catch {
    /* already shown */
  } finally {
    loading.value = false;
  }
  getData();
}

async function dnsRecordFormSubmit() {
  const valid = await dnsRecordFormRef.value?.validate().catch(() => false);
  if (!valid) {
    validateFailedMessage();
    return;
  }
  loading.value = true;
  try {
    if (dnsRecordFormType.value === 'add') {
      await createDnsRecord({ ...dnsRecordForm.value });
    } else if (dnsRecordFormType.value === 'update') {
      await updateDnsRecord({ ...dnsRecordForm.value });
    }
    successMessage();
  } catch {
    loading.value = false;
    return;
  }
  loading.value = false;
  dnsRecordFormReset();
  getData();
}

function dnsRecordFormCancel() {
  dnsRecordFormReset();
}

function resetRecordFormFields() {
  dnsRecordFormRef.value?.resetFields();
  dnsRecordForm.value = emptyRecordForm();
}

function dnsRecordFormReset() {
  dnsRecordFormVisible.value = false;
  resetRecordFormFields();
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
