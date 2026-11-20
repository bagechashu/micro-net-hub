<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form :inline="true" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('fieldRelation.610zu0w6ymo0')">
          <el-input
            v-model.trim="params.flag"
            clearable
            :placeholder="$t('fieldRelation.610zu0w6z440')"
            @keyup.enter="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" :icon="Search" @click="search">{{ $t('fieldRelation.610zu0w6z7w0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" :icon="Plus" type="primary" @click="create">{{ $t('fieldRelation.610zu0w6zag0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            :icon="Delete"
            type="danger"
            @click="batchDelete"
          >{{ $t('fieldRelation.610zu0w6zdw0') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="loading"
        :default-expand-all="true"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        row-key="ID"
        :data="infoTableData"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="48" align="center" />
        <el-table-column show-overflow-tooltip sortable prop="ID" :label="$t('fieldRelation.610zu0w6zgo0')" width="90" />
        <el-table-column show-overflow-tooltip sortable prop="CreatedAt" :label="$t('fieldRelation.610zu0w6ziw0')" width="180" />
        <el-table-column show-overflow-tooltip sortable prop="Flag" :label="$t('fieldRelation.610zu0w6ymo0')" min-width="190" />
        <el-table-column show-overflow-tooltip sortable :label="$t('fieldRelation.610zu0w6zm80')" min-width="240">
          <template #default="props">
            <span>{{ props.row.Attributes }}</span>
          </template>
        </el-table-column>
        <el-table-column fixed="right" :label="$t('fieldRelation.610zu0w6zos0')" align="left" width="160">
          <template #default="scope">
            <div class="row-actions">
              <el-button link type="primary" :icon="Edit" @click="update(scope.row)">{{ $t('common.edit') }}</el-button>
              <el-popconfirm :title="$t('fieldRelation.610zu0w6zwk0')" @confirm="singleDelete(scope.row.ID)">
                <template #reference>
                  <el-button link type="danger" :icon="Delete">{{ $t('common.delete') }}</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 新增 -->
      <el-dialog v-model="createDialogVisible" :title="dialogFormTitle">
        <div class="components-container">
          <aside>
            {{ $t('fieldRelation.610zu0w6zyo0') }}
            <a href="http://ldapdoc.eryajf.net/pages/84953d/" target="_blank">{{ $t('fieldRelation.610zu0w70140') }}</a>
          </aside>
        </div>
        <el-form ref="createFormRef" :model="dialogFormData" :rules="dialogFormRules" label-position="top">
          <el-form-item :label="$t('fieldRelation.610zu0w703g0')">
            <el-checkbox-group v-model="fieldRelationChecked">
              <el-checkbox-button
                v-for="type in fieldRelationTypes"
                :key="type"
                :value="type"
                @change="fieldRelationCheck(type)"
              >
                {{ type }}
              </el-checkbox-button>
            </el-checkbox-group>
          </el-form-item>

          <template v-if="isUserRelation">
            <el-form-item :label="$t('fieldRelation.610zu0w708g0')">
              <el-select v-model="userVal" :placeholder="$t('fieldRelation.610zu0w70ao0')">
                <el-option v-for="item in userOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
            <el-form-item v-for="f in userFields" :key="f.prop" :label="$t(f.label)" :prop="f.prop">
              <el-input v-model.trim="dialogFormData[f.prop]" :placeholder="$t(f.label)" />
            </el-form-item>
          </template>
          <template v-else-if="isGroupRelation">
            <el-form-item :label="$t('fieldRelation.610zu0w708g0')">
              <el-select v-model="groupVal" :placeholder="$t('fieldRelation.610zu0w70ao0')">
                <el-option v-for="item in groupOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
            <el-form-item v-for="f in groupFields" :key="f.prop" :label="$t(f.label)" :prop="f.prop">
              <el-input v-model.trim="dialogFormData[f.prop]" :placeholder="$t(f.label)" />
            </el-form-item>
          </template>
          <template v-else>
            <el-form-item><b>↑ {{ $t('fieldRelation.610zu0w71ck0') }}</b></el-form-item>
          </template>
        </el-form>
        <template #footer>
          <el-button @click="cancelForm()">{{ $t('fieldRelation.610zu0w71ec0') }}</el-button>
          <el-button :loading="submitLoading" type="primary" @click="submitForm()">{{ $t('fieldRelation.610zu0w71gc0') }}</el-button>
        </template>
      </el-dialog>

      <!-- 编辑 -->
      <el-dialog v-model="updateDialogVisible" :title="dialogFormTitle">
        <div class="components-container">
          <aside>
            {{ $t('fieldRelation.610zu0w6zyo0') }}
            <a href="http://ldapdoc.eryajf.net/pages/84953d/" target="_blank">{{ $t('fieldRelation.610zu0w70140') }}</a>
          </aside>
        </div>
        <el-form ref="updateFormRef" :model="dialogFormData" :rules="dialogFormRules" label-position="top">
          <template v-if="isUserRelation">
            <el-form-item :label="$t('fieldRelation.610zu0w703g0')">
              <el-button type="primary">{{ $t('fieldRelation.610zu0w70600') }}</el-button>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w708g0')">
              <el-select v-model="userVal" :placeholder="$t('fieldRelation.610zu0w70ao0')">
                <el-option v-for="item in userOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
            <el-form-item v-for="f in userFields" :key="f.prop" :label="$t(f.label)" :prop="f.prop">
              <el-input v-model.trim="dialogFormData[f.prop]" :placeholder="$t(f.label)" />
            </el-form-item>
          </template>
          <template v-else>
            <el-form-item :label="$t('fieldRelation.610zu0w703g0')">
              <el-button type="primary">{{ $t('fieldRelation.610zu0w712o0') }}</el-button>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w708g0')">
              <el-select v-model="groupVal" :placeholder="$t('fieldRelation.610zu0w70ao0')">
                <el-option v-for="item in groupOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
            <el-form-item v-for="f in groupFields" :key="f.prop" :label="$t(f.label)" :prop="f.prop">
              <el-input v-model.trim="dialogFormData[f.prop]" :placeholder="$t(f.label)" />
            </el-form-item>
          </template>
        </el-form>
        <template #footer>
          <el-button @click="cancelForm()">{{ $t('fieldRelation.610zu0w71ec0') }}</el-button>
          <el-button :loading="submitLoading" type="primary" @click="submitForm()">{{ $t('fieldRelation.610zu0w71gc0') }}</el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { Search, Plus, Delete, Edit } from '@/icons/tabler';
import { relationList, relationAdd, relationUp, relationDel } from '@/api/personnel/fieldRelation';

defineOptions({ name: 'FieldRelation' });

const { t } = useI18n();

const groupOptions = computed(() => [
  { label: t('fieldRelation.610zu0w71i00'), value: 'feishu_group' },
  { label: t('fieldRelation.610zu0w71jo0'), value: 'dingtalk_group' },
  { label: t('fieldRelation.610zu0w71lg0'), value: 'wecom_group' }
]);
const userOptions = computed(() => [
  { label: t('fieldRelation.610zu0w71i00'), value: 'feishu_user' },
  { label: t('fieldRelation.610zu0w71jo0'), value: 'dingtalk_user' },
  { label: t('fieldRelation.610zu0w71lg0'), value: 'wecom_user' }
]);

// 用户字段表单项
const userFields = [
  { prop: 'username', label: 'fieldRelation.610zu0w70cs0' }, // 用户名(通常为用户名拼音) name_pinyin
  { prop: 'nickname', label: 'fieldRelation.610zu0w70fc0' }, // 中文名字 name
  { prop: 'givenName', label: 'fieldRelation.610zu0w70ik0' }, // 花名 name
  { prop: 'mail', label: 'fieldRelation.610zu0w70kk0' }, // 邮箱 email
  { prop: 'jobNumber', label: 'fieldRelation.610zu0w70nc0' }, // 工号 job_number
  { prop: 'mobile', label: 'fieldRelation.610zu0w70p80' }, // 手机号 mobile
  { prop: 'avatar', label: 'fieldRelation.610zu0w70qw0' }, // 头像 avatar
  { prop: 'postalAddress', label: 'fieldRelation.610zu0w70sw0' }, // 地址 work_place
  { prop: 'position', label: 'fieldRelation.610zu0w70uo0' }, // 职位 title
  { prop: 'sourceUserId', label: 'fieldRelation.610zu0w70ws0' }, // 源用户ID userid
  { prop: 'sourceUnionId', label: 'fieldRelation.610zu0w70ys0' }, // 源用户唯一ID unionid
  { prop: 'introduction', label: 'fieldRelation.610zu0w710s0' } // 说明 remark
];
// 分组字段表单项
const groupFields = [
  { prop: 'groupName', label: 'fieldRelation.610zu0w714g0' }, // 分组名称
  { prop: 'sourceDeptParentId', label: 'fieldRelation.610zu0w716c0' }, // 父部门ID
  { prop: 'sourceDeptId', label: 'fieldRelation.610zu0w71840' }, // 部门ID
  { prop: 'remark', label: 'fieldRelation.610zu0w71as0' } // 分组描述
];

const userVal = ref('');
const groupVal = ref('');
const updateId = ref<number | string>('');
const fieldRelationChecked = ref<string[]>(['userFieldRelation']); // 新增数据默认选中
const fieldRelationTypes = ['userFieldRelation', 'groupFieldRelation'];

const isUserRelation = computed(() => fieldRelationChecked.value.length === 1 && fieldRelationChecked.value[0] === 'userFieldRelation');
const isGroupRelation = computed(() => fieldRelationChecked.value.length === 1 && fieldRelationChecked.value[0] === 'groupFieldRelation');

// 查询参数
const params = reactive({
  flag: '',
  pageNum: 1,
  pageSize: 1000 // 后台限制最大单次获取1000条
});
// 表格数据
const tableData = ref<any[]>([]);
const infoTableData = ref<any[]>([]);
const loading = ref(false);
// 表格多选
const multipleSelection = ref<any[]>([]);

// dialog对话框
const createDialogVisible = ref(false); // 新增
const updateDialogVisible = ref(false); // 编辑
const submitLoading = ref(false);
const dialogFormTitle = ref('');
const dialogType = ref('');
const dialogFormData = ref<Record<string, any>>({});
const createFormRef = ref<FormInstance>();
const updateFormRef = ref<FormInstance>();

const lengthRule = (min: number, max: number, trigger: string | string[] = 'blur', msg: [number, number] = [min, max]) => ({
  min,
  max,
  message: t('valid.length', msg),
  trigger
});

const dialogFormRules = computed<FormRules>(() => ({
  sourceDeptParentId: [{ required: true, message: t('fieldRelation.610zu0w71n40'), trigger: 'blur' }, lengthRule(1, 50)],
  sourceDeptId: [{ required: true, message: t('fieldRelation.610zu0w71ow0'), trigger: 'blur' }, lengthRule(1, 50)],
  username: [{ required: true, message: t('fieldRelation.610zu0w71r40'), trigger: 'blur' }, lengthRule(1, 50)],
  givenName: [{ required: true, message: t('fieldRelation.610zu0w71to0'), trigger: 'blur' }, lengthRule(1, 50)],
  avatar: [{ required: true, message: t('fieldRelation.610zu0w71to0'), trigger: 'blur' }, lengthRule(1, 50)],
  postalAddress: [{ required: true, message: t('fieldRelation.610zu0w71to0'), trigger: 'blur' }, lengthRule(1, 50)],
  position: [{ required: true, message: t('fieldRelation.610zu0w71to0'), trigger: 'blur' }, lengthRule(1, 50)],
  sourceUserId: [{ required: true, message: t('fieldRelation.610zu0w71to0'), trigger: 'blur' }, lengthRule(1, 50, ['blur', 'change'])],
  sourceUnionId: [{ required: true, message: t('fieldRelation.610zu0w71to0'), trigger: 'blur' }, lengthRule(1, 50, ['blur', 'change'])],
  groupName: [{ required: true, message: t('fieldRelation.610zu0w71w00'), trigger: 'blur' }, lengthRule(1, 50)],
  remark: [{ required: true, message: t('fieldRelation.610zu0w71xs0'), trigger: 'blur' }, lengthRule(1, 50)],
  mail: [{ required: true, message: t('fieldRelation.610zu0w71zg0'), trigger: 'blur' }, lengthRule(1, 50)],
  jobNumber: [{ required: false, message: t('fieldRelation.610zu0w722w0'), trigger: 'blur' }, lengthRule(0, 20, 'blur', [2, 20])],
  nickname: [{ required: true, message: t('fieldRelation.610zu0w724o0'), trigger: 'blur' }, lengthRule(2, 20)],
  mobile: [{ required: false, message: t('fieldRelation.610zu0w726g0'), trigger: 'blur' }],
  introduction: [{ required: false, message: t('fieldRelation.610zu0w710s0'), trigger: 'blur' }, lengthRule(0, 100)]
}));

function fieldRelationCheck(type: string) {
  fieldRelationChecked.value = fieldRelationChecked.value.includes(type) ? [type] : [];
}

// 页面数据过滤: 不符合条件的节点被移除, 其符合条件的子节点被提升上来
function deal(nodes: any[] | undefined, predicate: (node: any) => boolean): any[] {
  if (!(nodes && nodes.length)) {
    return [];
  }
  const newChildren: any[] = [];
  for (const node of nodes) {
    if (predicate(node)) {
      newChildren.push(node);
      node.children = deal(node.children, predicate);
    } else {
      newChildren.push(...deal(node.children, predicate));
    }
  }
  return newChildren;
}

// 查询 (前端过滤)
function search() {
  const copy = JSON.parse(JSON.stringify(tableData.value));
  infoTableData.value = deal(copy, node => String(node.Flag ?? '').includes(params.flag));
}

// 获取表格数据
async function getTableData() {
  loading.value = true;
  try {
    const { data } = await relationList({ pageNum: params.pageNum, pageSize: params.pageSize });
    tableData.value = data || [];
    infoTableData.value = JSON.parse(JSON.stringify(tableData.value));
    if (params.flag) {
      search();
    }
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
}

// 新增
function create() {
  fieldRelationChecked.value = ['userFieldRelation'];
  userVal.value = '';
  groupVal.value = '';
  dialogFormData.value = {};
  dialogFormTitle.value = t('fieldRelation.610zu0w6zag0');
  dialogType.value = 'create';
  createDialogVisible.value = true;
  createFormRef.value?.clearValidate();
}

// 修改
function update(row: any) {
  const typeDialog = String(row.Flag || '').split('_')[1];
  const attrs = row.Attributes || {};
  updateId.value = row.ID;

  if (typeDialog === 'user') {
    fieldRelationChecked.value = ['userFieldRelation'];
    userVal.value = row.Flag;
    const data: Record<string, any> = {};
    userFields.forEach(f => { data[f.prop] = attrs[f.prop]; });
    dialogFormData.value = data;
  } else {
    fieldRelationChecked.value = ['groupFieldRelation'];
    groupVal.value = row.Flag;
    const data: Record<string, any> = {};
    groupFields.forEach(f => { data[f.prop] = attrs[f.prop]; });
    dialogFormData.value = data;
  }

  dialogFormTitle.value = t('fieldRelation.610zu0w72880');
  dialogType.value = 'update';
  updateDialogVisible.value = true;
  updateFormRef.value?.clearValidate();
}

// 提交表单
async function submitForm() {
  let flag: string;
  if (fieldRelationChecked.value[0] === 'userFieldRelation') {
    if (userVal.value === '') {
      ElMessage({ message: t('fieldRelation.610zu0w72a40'), type: 'warning' });
      return;
    }
    flag = userVal.value;
  } else {
    if (groupVal.value === '') {
      ElMessage({ message: t('fieldRelation.610zu0w72a40'), type: 'warning' });
      return;
    }
    flag = groupVal.value;
  }
  const attributes = dialogFormData.value;

  const formRef = dialogType.value === 'create' ? createFormRef.value : updateFormRef.value;
  if (!formRef) return;
  const valid = await formRef.validate().catch(() => false);
  if (!valid) {
    ElMessage({ showClose: true, message: t('fieldRelation.610zu0w72dg0'), type: 'warning' });
    return;
  }

  submitLoading.value = true;
  try {
    if (dialogType.value === 'create') {
      await relationAdd({ flag, attributes });
    } else {
      await relationUp({ id: updateId.value, flag, attributes });
    }
  } catch {
    // error already shown
    return;
  } finally {
    submitLoading.value = false;
  }
  resetForm();
  getTableData();
  ElMessage({ showClose: true, message: t('fieldRelation.610zu0w72bs0'), type: 'success' });
}

function cancelForm() {
  resetForm();
}

function resetForm() {
  createDialogVisible.value = false;
  updateDialogVisible.value = false;
  createFormRef.value?.resetFields();
  updateFormRef.value?.resetFields();
  dialogFormData.value = {};
}

// 批量删除
async function batchDelete() {
  try {
    await ElMessageBox.confirm(t('fieldRelation.batchDeleteTips'), t('fieldRelation.610zu0w72fk0'), {
      confirmButtonText: t('fieldRelation.610zu0w71gc0'),
      cancelButtonText: t('fieldRelation.610zu0w71ec0'),
      type: 'warning'
    });
  } catch {
    ElMessage({ showClose: true, type: 'info', message: t('fieldRelation.610zu0w72jk0') });
    return;
  }
  loading.value = true;
  try {
    await relationDel({ fieldRelationIds: multipleSelection.value.map(x => x.ID) });
    ElMessage({ showClose: true, message: t('fieldRelation.610zu0w72hg0'), type: 'success' });
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
  getTableData();
}

// 单个删除
async function singleDelete(id: number) {
  loading.value = true;
  try {
    await relationDel({ fieldRelationIds: [id] });
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
  getTableData();
}

// 表格多选
function handleSelectionChange(val: any[]) {
  multipleSelection.value = val;
}

onMounted(() => {
  getTableData();
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
