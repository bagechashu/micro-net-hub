<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="small" :inline="true" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('operation-log.6111tvyrbi80')">
          <el-input v-model.trim="params.username" clearable :placeholder="$t('operation-log.6111tvyrbi80')" @keyup.enter="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('operation-log.ipaddr')">
          <el-input v-model.trim="params.ip" clearable :placeholder="$t('operation-log.ipaddr')" @keyup.enter="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('operation-log.6111tvyrbt00')">
          <el-input v-model.trim="params.path" clearable :placeholder="$t('operation-log.6111tvyrbt00')" @keyup.enter="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('operation-log.6111tvyrbw80')">
          <el-input v-model.trim="params.status" clearable :placeholder="$t('operation-log.6111tvyrbw80')" @keyup.enter="search" @clear="search" />
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" :icon="Search" type="primary" @click="search">{{ $t('operation-log.6111tvyrbyo0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :disabled="multipleSelection.length === 0" :loading="loading" :icon="Delete" type="danger" @click="batchDelete">{{ $t('operation-log.6111tvyrc0o0') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column show-overflow-tooltip sortable prop="username" :label="$t('operation-log.6111tvyrbi80')" width="120" />
        <el-table-column show-overflow-tooltip sortable prop="ip" :label="$t('operation-log.ipaddr')" width="120" />
        <el-table-column show-overflow-tooltip sortable prop="method" :label="$t('operation-log.6111tvyrc2s0')" width="100" />
        <el-table-column show-overflow-tooltip sortable prop="path" :label="$t('operation-log.6111tvyrbt00')" />
        <el-table-column show-overflow-tooltip sortable prop="status" :label="$t('operation-log.6111tvyrbw80')" width="100" align="center">
          <template #default="scope">
            <el-tag size="small" :type="statusTagType(scope.row.status)" disable-transitions>{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="startTime" :label="$t('operation-log.6111tvyrc4w0')" width="300" />
        <el-table-column show-overflow-tooltip sortable prop="timeCost" :label="$t('operation-log.6111tvyrc8s0')" width="130" align="center">
          <template #default="scope">
            <el-tag size="small" :type="timeCostTagType(scope.row.timeCost)" disable-transitions>{{ scope.row.timeCost }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="desc" :label="$t('operation-log.6111tvyrcaw0')" />
        <el-table-column fixed="right" :label="$t('operation-log.6111tvyrccw0')" align="center" width="80">
          <template #default="scope">
            <el-tooltip :content="$t('operation-log.6111tvyrcf40')" effect="dark" placement="top">
              <span>
                <el-popconfirm :title="$t('operation-log.6111tvyrchk0')" @confirm="singleDelete(scope.row.ID)">
                  <template #reference>
                    <el-button size="small" :icon="Delete" circle type="danger" />
                  </template>
                </el-popconfirm>
              </span>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        :current-page="params.pageNum"
        :page-size="params.pageSize"
        :total="total"
        :page-sizes="[1, 5, 10, 30]"
        layout="total, prev, pager, next, sizes"
        background
        style="margin-top: 10px;float:right;margin-bottom: 10px;"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Delete } from '@element-plus/icons-vue';
import { getOperationLogs, batchDeleteOperationLogByIds } from '@/api/log/operationLog';

defineOptions({ name: 'OperationLog' });

type TagType = 'success' | 'warning' | 'danger' | 'info' | 'primary';

const { t } = useI18n();

// 查询参数
const params = reactive({
  username: '',
  ip: '',
  path: '',
  status: '',
  pageNum: 1,
  pageSize: 10
});
// 表格数据
const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);
// 表格多选
const multipleSelection = ref<any[]>([]);

function statusTagType(val: number): TagType {
  if (val === 200) return 'success';
  if (val === 400) return 'warning';
  if (val === 401 || val === 403 || val === 500) return 'danger';
  return 'info';
}

function timeCostTagType(val: number): TagType {
  if (val <= 200) return 'success';
  if (val <= 1000) return 'primary';
  if (val <= 2000) return 'warning';
  return 'danger';
}

// 查询
function search() {
  params.pageNum = 1;
  getTableData();
}

// 获取表格数据
async function getTableData() {
  loading.value = true;
  try {
    const { data } = await getOperationLogs(params);
    tableData.value = data.logs;
    total.value = data.total;
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
}

function showDeleteSuccess() {
  ElMessage({
    showClose: true,
    message: t('operation-log.6111tvyrcjk0'),
    type: 'success'
  });
}

async function deleteByIds(operationLogIds: number[]) {
  loading.value = true;
  try {
    await batchDeleteOperationLogByIds({ operationLogIds });
    showDeleteSuccess();
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
  getTableData();
}

// 批量删除
async function batchDelete() {
  try {
    await ElMessageBox.confirm(t('tips.deleteWarning'), t('operation-log.6111tvyrcls0'), {
      confirmButtonText: t('operation-log.6111tvyrcnw0'),
      cancelButtonText: t('operation-log.6111tvyrcq40'),
      type: 'warning'
    });
  } catch {
    ElMessage({
      showClose: true,
      type: 'info',
      message: t('operation-log.6111tvyrcs40')
    });
    return;
  }
  await deleteByIds(multipleSelection.value.map(x => x.ID));
}

// 表格多选
function handleSelectionChange(val: any[]) {
  multipleSelection.value = val;
}

// 单个删除
function singleDelete(id: number) {
  return deleteByIds([id]);
}

// 分页
function handleSizeChange(val: number) {
  params.pageSize = val;
  getTableData();
}
function handleCurrentChange(val: number) {
  params.pageNum = val;
  getTableData();
}

onMounted(() => {
  getTableData();
});
</script>

<style scoped>
  .container-card{
    margin: 10px;
    margin-bottom: 100px;
  }

  .delete-popover{
    margin-left: 10px;
  }
</style>
