<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        :inline="true"
        :model="params"
        class="demo-form-inline"
      >
        <el-form-item :label="$t('group.610yx3auj8w0')">
          <el-input
            v-model.trim="params.groupName"
            style="width: 160px"
            clearable
            :placeholder="$t('group.610yx3auj8w0')"
            @keyup.enter="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item :label="$t('group.610yx3auk3k0')">
          <el-input
            v-model.trim="params.remark"
            style="width: 160px"
            clearable
            :placeholder="$t('group.610yx3auk3k0')"
            @keyup.enter="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item :label="$t('group.610yx3auk780')">
          <el-select
            v-model="params.syncState"
            style="width: 160px"
            clearable
            :placeholder="$t('group.610yx3auk780')"
            @change="search"
            @clear="search"
          >
            <el-option :label="$t('group.610yx3auk9s0')" value="1" />
            <el-option :label="$t('group.610yx3aukc00')" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            :icon="Search"
            @click="search"
          >{{ $t('group.610yx3auke40') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            :icon="Plus"
            type="primary"
            @click="create"
          >{{ $t('group.610yx3aukig0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            :icon="Delete"
            type="danger"
            @click="batchDelete"
          >{{ $t('group.610yx3aukks0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            :icon="Upload"
            type="success"
            @click="batchSync"
          >{{ $t('group.610yx3aukmw0') }}</el-button>
        </el-form-item>
        <br>
        <el-form-item>
          <el-button
            :loading="loading"
            :icon="Download"
            @click="runDeptSync(syncOpenLdapDeptsApi)"
          >{{ $t('group.610yx3aukos0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            :icon="Download"
            @click="runDeptSync(syncDingTalkDeptsApi)"
          >{{ $t('group.610yx3aukr40') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            :icon="Download"
            @click="runDeptSync(syncFeiShuDeptsApi)"
          >{{ $t('group.610yx3aukvw0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            :icon="Download"
            @click="runDeptSync(syncWeComDeptsApi)"
          >{{ $t('group.610yx3aukyc0') }}</el-button>
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
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="groupName"
          :label="$t('group.610yx3auj8w0')"
          min-width="180"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="groupType"
          :label="$t('group.610yx3aul140')"
          width="190"
        >
          <template #default="scope">
            {{ scope.row.groupType === 'ou' ? 'organizationalUnit' : 'groupOfUniqueNames' }}
          </template>
        </el-table-column>
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="groupDn"
          label="DN"
          min-width="300"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="remark"
          :label="$t('group.610yx3auk3k0')"
          min-width="160"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="CreatedAt"
          :label="$t('group.610yx3aul380')"
          width="180"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="UpdatedAt"
          :label="$t('group.610yx3aul5c0')"
          width="180"
        />
        <el-table-column fixed="right" :label="$t('group.610yx3aul7g0')" align="left" width="320">
          <template #default="scope">
            <div class="row-actions">
              <el-button v-if="scope.row.groupType != 'ou' && scope.row.groupName != 'root'" link type="primary" :icon="Users" @click="handleGetUserOfGroup(scope.row)">{{ $t('group.members') }}</el-button>
              <el-button link type="primary" :icon="Edit" @click="update(scope.row)">{{ $t('common.edit') }}</el-button>
              <el-popconfirm :title="$t('group.610yx3aulgo0')" @confirm="singleDelete(scope.row.ID)">
                <template #reference>
                  <el-button link type="danger" :icon="Delete">{{ $t('common.delete') }}</el-button>
                </template>
              </el-popconfirm>
              <el-popconfirm v-if="scope.row.syncState === 2" :title="$t('group.610yx3aulks0')" @confirm="singleSync(scope.row.ID)">
                <template #reference>
                  <el-button link type="primary" :icon="Upload">{{ $t('group.610yx3aulis0') }}</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <!-- 新增 -->
      <el-dialog v-model="dialogGroupAddFormVisible" :title="dialogFormTitle" @closed="resetForm">
        <el-form
          ref="addFormRef"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-position="top"
        >
          <el-form-item :label="$t('group.610yx3auj8w0')" prop="groupName">
            <el-input
              v-model.trim="dialogFormData.groupName"
              :placeholder="$t('group.610yx3auj8w0')"
            />
          </el-form-item>
          <el-form-item :label="$t('group.610yx3aulns0')" prop="groupType">
            <el-select
              v-model="dialogFormData.groupType"
              :placeholder="$t('group.610yx3aulpo0')"
              style="width: 100%"
            >
              <el-option label="ou [organizationalUnit]" value="ou" />
              <el-option label="cn [groupOfUniqueNames]" value="cn" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('group.610yx3aulrk0')" prop="parentId">
            <TreeSelect
              v-model="dialogFormData.parentId"
              :options="treeselectData"
              :normalizer="normalizer"
              :placeholder="$t('group.610yx3aultc0')"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item :label="$t('group.610yx3auk3k0')" prop="remark">
            <el-input
              v-model.trim="dialogFormData.remark"
              type="textarea"
              :placeholder="$t('group.610yx3auk3k0')"
              :autosize="{ minRows: 3, maxRows: 6 }"
              show-word-limit
              maxlength="100"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="cancelForm()">{{ $t('group.610yx3aulv00') }}</el-button>
          <el-button
            :loading="submitLoading"
            type="primary"
            @click="submitForm()"
          >{{ $t('group.610yx3aulws0') }}</el-button>
        </template>
      </el-dialog>
      <!-- 编辑 -->
      <el-dialog v-model="dialogGroupEditFormVisible" :title="dialogFormTitle" @closed="resetForm">
        <el-form
          ref="editFormRef"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-position="top"
        >
          <el-form-item :label="$t('group.610yx3auj8w0')" prop="groupName">
            <el-input
              v-model.trim="dialogFormData.groupName"
              :disabled="true"
              :placeholder="$t('group.610yx3auj8w0')"
            />
          </el-form-item>
          <el-form-item :label="$t('group.610yx3auk3k0')" prop="remark">
            <el-input
              v-model.trim="dialogFormData.remark"
              type="textarea"
              :placeholder="$t('group.610yx3auk3k0')"
              :autosize="{ minRows: 3, maxRows: 6 }"
              show-word-limit
              maxlength="100"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="cancelForm()">{{ $t('group.610yx3aulv00') }}</el-button>
          <el-button
            :loading="submitLoading"
            type="primary"
            @click="submitForm()"
          >{{ $t('group.610yx3aulws0') }}</el-button>
        </template>
      </el-dialog>
      <!-- 组用户管理 -->
      <el-dialog v-model="dialogUsersMgrVisible" :title="dialogUsersMgrTitle" width="860px">
        <!-- 添加成员：远程搜索不在组内的用户，可多选 -->
        <div class="members-add">
          <el-select
            v-model="groupAddUserList"
            class="members-add__select"
            multiple
            filterable
            remote
            reserve-keyword
            collapse-tags
            collapse-tags-tooltip
            :max-collapse-tags="3"
            :placeholder="$t('group.addMembersTips')"
            :remote-method="getUsersNoInGroupData"
            :loading="submitLoading"
          >
            <template #prefix><el-icon><Plus /></el-icon></template>
            <el-option
              v-for="item in groupAddUserListOption"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-button
            type="primary"
            :disabled="groupAddUserList.length === 0"
            :loading="submitLoading"
            :icon="Plus"
            @click="handleGroupAddUser()"
          >{{ $t('group.610yx3aum1w0') }}</el-button>
        </div>

        <!-- 当前成员：筛选 + 批量移除 -->
        <div class="members-toolbar">
          <el-input
            v-model.trim="userOfGroupParams.nickname"
            class="members-toolbar__search"
            clearable
            :prefix-icon="Search"
            :placeholder="$t('group.610yx3aulyk0')"
            @keyup.enter="getUsersOfGroupData()"
            @clear="getUsersOfGroupData()"
          />
          <el-button :loading="submitLoading" @click="getUsersOfGroupData()">{{ $t('group.610yx3aum0c0') }}</el-button>
          <el-button
            class="members-toolbar__remove"
            type="danger"
            plain
            :disabled="userOfGroupListMultipleSelection.length === 0"
            :loading="submitLoading"
            :icon="Delete"
            @click="handleGroupDelUser()"
          >{{ $t('group.610yx3aum3o0') }}<template v-if="userOfGroupListMultipleSelection.length"> ({{ userOfGroupListMultipleSelection.length }})</template></el-button>
        </div>

        <el-table
          :data="userOfGroupList"
          height="360"
          style="width: 100%"
          :default-sort="{ prop: 'userName', order: 'ascending' }"
          @selection-change="handleUserOfGroupListSelectionChange"
        >
          <el-table-column type="selection" width="48" />
          <el-table-column type="expand" width="44">
            <template #default="props">
              <RowDetail
                :items="[
                  { label: $t('group.610yx3aum7g0'), value: props.row.mail },
                  { label: $t('group.610yx3aum5k0'), value: props.row.mobile }
                ]"
              />
            </template>
          </el-table-column>
          <el-table-column prop="userName" :label="$t('group.610yx3aum940')" min-width="140" sortable />
          <el-table-column prop="nickName" :label="$t('group.610yx3aumb40')" min-width="120" />
          <el-table-column prop="introduction" :label="$t('group.610yx3aumcw0')" min-width="160" />
        </el-table>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { Search, Plus, Delete, Upload, Download, Edit, Users } from '@/icons/tabler';
import TreeSelect from '@zanmato/vue3-treeselect';
import RowDetail from '@/components/RowDetail/index.vue';
import '@zanmato/vue3-treeselect/dist/vue3-treeselect.min.css';
import {
  getGroupTree,
  groupAdd,
  groupUpdate,
  groupDel,
  syncDingTalkDeptsApi,
  syncWeComDeptsApi,
  syncFeiShuDeptsApi,
  syncOpenLdapDeptsApi,
  syncSqlGroups,
  userInGroup,
  userNoInGroup,
  groupAddUser,
  groupDelUser
} from '@/api/personnel/group';

defineOptions({ name: 'Group' });

const { t } = useI18n();

// 查询参数
const params = reactive<Record<string, any>>({
  groupName: undefined,
  remark: undefined,
  syncState: undefined,
  pageNum: 1,
  pageSize: 1000 // 后台限制最大单次获取1000条
});

// 表格数据
const groupTree = ref<any[]>([]);
const infoTableData = ref<any[]>([]);
const loading = ref(false);
// 上级目录数据
const treeselectData = ref<any[]>([]);

// dialog对话框
const addFormRef = ref<FormInstance>();
const editFormRef = ref<FormInstance>();
const dialogGroupAddFormVisible = ref(false);
const dialogGroupEditFormVisible = ref(false);
const submitLoading = ref(false);
const dialogFormTitle = ref('');
const dialogType = ref('');

const emptyFormData = () => ({
  ID: '' as any,
  groupName: '',
  parentId: 0 as any,
  syncState: 1,
  groupType: '',
  remark: ''
});
const dialogFormData = ref(emptyFormData());

const validateName = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (value === '') {
    return callback(new Error(t('valid.notAllowEmpty')));
  }
  if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
    return callback(new Error(t('valid.InvalidName')));
  }
  callback();
};

const dialogFormRules: FormRules = {
  groupName: [
    { required: true, message: t('group.610yx3aumeo0'), trigger: 'blur' },
    { min: 2, max: 30, message: t('valid.length', [2, 30]), trigger: 'blur' },
    { required: true, validator: validateName, trigger: 'blur' }
  ],
  groupType: [
    { required: true, message: t('group.610yx3aumgg0'), trigger: 'blur' },
    { min: 1, max: 50, message: t('group.selectGroupTypeTips'), trigger: 'blur' }
  ],
  parentId: [
    { required: true, message: t('group.610yx3aumi40'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value >= 0) {
          callback();
        } else {
          callback(new Error(t('group.610yx3aumjw0')));
        }
      }
    }
  ],
  remark: [
    { required: false, message: t('group.610yx3aumlk0'), trigger: 'blur' },
    { min: 0, max: 100, message: t('valid.length', [0, 100]), trigger: 'blur' }
  ]
};

// 表格多选
const multipleSelection = ref<any[]>([]);

// 组用户管理
const dialogUsersMgrVisible = ref(false);
const dialogUsersMgrTitle = ref('');
const userOfGroupParams = reactive({ groupId: '' as any, nickname: '' });
const userOfGroupList = ref<any[]>([]);
const userOfGroupListMultipleSelection = ref<any[]>([]);
const groupAddUserParams = reactive({ groupId: '' as any, nickname: '' });
const groupAddUserListOption = ref<{ label: string; value: any }[]>([]);
const groupAddUserList = ref<any[]>([]);

const showSuccess = () => {
  ElMessage({ showClose: true, message: t('group.610yx3aumtg0'), type: 'success' });
};

const confirmAction = async (message: string, cancelMessage: string) => {
  try {
    await ElMessageBox.confirm(message, t('group.610yx3aumrs0'), {
      confirmButtonText: t('group.610yx3aulws0'),
      cancelButtonText: t('group.610yx3aulv00'),
      type: 'warning'
    });
    return true;
  } catch {
    ElMessage({ showClose: true, type: 'info', message: cancelMessage });
    return false;
  }
};

// 页面数据过滤
const deal = (nodes: any[], predicate: (node: any) => boolean): any[] => {
  if (!(nodes && nodes.length)) {
    return [];
  }
  const newChildren: any[] = [];
  for (const node of nodes) {
    if (predicate(node)) {
      // 节点符合条件，直接加入新的节点集
      newChildren.push(node);
      node.children = deal(node.children, predicate);
    } else {
      // 当前节点不符合条件，递归过滤子节点，把符合条件的子节点提升上来
      newChildren.push(...deal(node.children, predicate));
    }
  }
  return newChildren;
};

// 查询
const search = () => {
  const tree = JSON.parse(JSON.stringify(groupTree.value));
  infoTableData.value = deal(
    tree,
    (node) =>
      node.groupName.includes(params.groupName) ||
      node.remark.includes(params.remark) ||
      node.syncState.toString().includes(params.syncState)
  );
};

// 获取表格数据
const getGroupTreeData = async () => {
  loading.value = true;
  try {
    const { data } = await getGroupTree(params);
    groupTree.value = data;
    infoTableData.value = JSON.parse(JSON.stringify(data));
    treeselectData.value = [{ ID: 0, groupName: t('group.610yx3aumo00'), children: data }];
  } catch {
    // 错误已提示
  } finally {
    loading.value = false;
  }
};

// 用户管理
const handleGetUserOfGroup = (row: any) => {
  dialogUsersMgrVisible.value = true;
  dialogUsersMgrTitle.value = t('group.group', [row.groupName]);
  userOfGroupParams.groupId = row.ID;
  userOfGroupParams.nickname = '';
  groupAddUserParams.groupId = row.ID;
  groupAddUserParams.nickname = '';
  getUsersOfGroupData();
};

// 获取组内用户数据
const getUsersOfGroupData = async () => {
  submitLoading.value = true;
  try {
    const { data } = await userInGroup(userOfGroupParams);
    userOfGroupList.value = data.userList;
  } catch {
    // 错误已提示
  } finally {
    submitLoading.value = false;
  }
};

const handleUserOfGroupListSelectionChange = (val: any[]) => {
  userOfGroupListMultipleSelection.value = val;
};

// 删除组用户
const handleGroupDelUser = async () => {
  if (!(await confirmAction(t('group.610yx3aumps0'), t('group.610yx3aumvc0')))) return;
  loading.value = true;
  try {
    await groupDelUser({
      groupId: Number(userOfGroupParams.groupId),
      userIds: userOfGroupListMultipleSelection.value.map((x) => x.userId)
    });
    showSuccess();
  } catch {
    // 错误已提示
  } finally {
    loading.value = false;
  }
  getUsersOfGroupData();
};

// 获取非组内用户数据
const getUsersNoInGroupData = async (query: string) => {
  submitLoading.value = true;
  groupAddUserParams.nickname = query;
  try {
    const { data } = await userNoInGroup(groupAddUserParams);
    groupAddUserListOption.value = (data.userList || []).map((item: any) => ({
      label: item.userName,
      value: item.userId
    }));
  } catch {
    // 错误已提示
  } finally {
    submitLoading.value = false;
  }
};

// 添加用户到组
const handleGroupAddUser = async () => {
  if (!(await confirmAction(t('group.610yx3aumx00'), t('group.610yx3aumvc0')))) return;
  loading.value = true;
  try {
    await groupAddUser({
      groupId: Number(groupAddUserParams.groupId),
      userIds: groupAddUserList.value
    });
    showSuccess();
  } catch {
    // 错误已提示
  } finally {
    loading.value = false;
  }
  userOfGroupParams.nickname = '';
  groupAddUserList.value = [];
  groupAddUserParams.nickname = '';
  getUsersOfGroupData();
};

// 新增
const create = () => {
  dialogFormTitle.value = t('group.610yx3aumyw0');
  dialogType.value = 'create';
  dialogGroupAddFormVisible.value = true;
};

// 修改
const update = (row: any) => {
  dialogFormData.value.ID = row.ID;
  dialogFormData.value.groupName = row.groupName;
  dialogFormData.value.remark = row.remark;
  dialogFormTitle.value = t('group.610yx3aun0k0');
  dialogType.value = 'update';
  dialogGroupEditFormVisible.value = true;
};

const resetForm = () => {
  dialogGroupEditFormVisible.value = false;
  dialogGroupAddFormVisible.value = false;
  addFormRef.value?.resetFields();
  editFormRef.value?.resetFields();
  dialogFormData.value = emptyFormData();
};

// 提交表单
const submitForm = async () => {
  const formRef = dialogType.value === 'create' ? addFormRef.value : editFormRef.value;
  if (!formRef) return;
  const valid = await formRef.validate().catch(() => false);
  if (!valid) {
    ElMessage({ showClose: true, message: t('group.610yx3aun2c0'), type: 'warning' });
    return;
  }
  submitLoading.value = true;
  try {
    if (dialogType.value === 'create') {
      await groupAdd(dialogFormData.value);
    } else {
      await groupUpdate(dialogFormData.value);
    }
    showSuccess();
  } catch {
    // 错误已提示
  } finally {
    submitLoading.value = false;
  }
  resetForm();
  getGroupTreeData();
};

const cancelForm = () => {
  resetForm();
};

// 批量删除
const batchDelete = async () => {
  if (!(await confirmAction(t('group.batchDeleteTips'), t('group.610yx3aumvc0')))) return;
  loading.value = true;
  try {
    await groupDel({ groupIds: multipleSelection.value.map((x) => x.ID) });
    showSuccess();
  } catch {
    // 错误已提示
  } finally {
    loading.value = false;
  }
  getGroupTreeData();
};

// 批量同步
const batchSync = async () => {
  if (!(await confirmAction(t('group.betchSyncLdapTips'), t('group.610yx3aun440')))) return;
  loading.value = true;
  try {
    await syncSqlGroups({ groupIds: multipleSelection.value.map((x) => x.ID) });
    showSuccess();
  } catch {
    // 错误已提示
  } finally {
    loading.value = false;
  }
  getGroupTreeData();
};

const handleSelectionChange = (val: any[]) => {
  multipleSelection.value = val;
};

// 单个删除
const singleDelete = async (id: number) => {
  loading.value = true;
  try {
    await groupDel({ groupIds: [id] });
    showSuccess();
  } catch {
    // 错误已提示
  } finally {
    loading.value = false;
  }
  getGroupTreeData();
};

// 单个同步
const singleSync = async (id: number) => {
  loading.value = true;
  try {
    await syncSqlGroups({ groupIds: [id] });
    showSuccess();
  } catch {
    // 错误已提示
  } finally {
    loading.value = false;
  }
  getGroupTreeData();
};

// treeselect
const normalizer = (node: any) => ({
  id: node.ID,
  label: node.groupName,
  children: node.children
});

// 同步 OpenLDAP / 钉钉 / 飞书 / 企业微信 部门
const runDeptSync = async (api: (data?: unknown) => Promise<any>) => {
  loading.value = true;
  try {
    await api();
  } catch {
    // 错误已提示
    return;
  } finally {
    loading.value = false;
  }
  showSuccess();
  getGroupTreeData();
};

onMounted(() => {
  getGroupTreeData();
});
</script>

<style scoped>
.members-add,
.members-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.members-add {
  padding-bottom: 16px;
  margin-bottom: 16px;
  border-bottom: 1px dashed var(--el-border-color);
}

.members-add__select {
  flex: 1;
  min-width: 240px;
}

.members-toolbar__search {
  width: 240px;
}

.members-toolbar__remove {
  margin-left: auto !important;
}

.container-card {
  margin: 10px;
  margin-bottom: 100px;
}

.delete-popover {
  margin-left: 10px;
}

</style>
