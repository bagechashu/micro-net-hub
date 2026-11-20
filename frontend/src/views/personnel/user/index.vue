<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form :inline="true" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('user.610wkzjikzs0')">
          <el-input
            v-model.trim="params.username"
            style="width: 160px"
            clearable
            :placeholder="$t('user.610wkzjikzs0')"
            @keyup.enter="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item :label="$t('user.610wkzjill40')">
          <el-input
            v-model.trim="params.nickname"
            style="width: 160px"
            clearable
            :placeholder="$t('user.610wkzjill40')"
            @keyup.enter="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item :label="$t('common.status')">
          <el-select
            v-model="params.status"
            style="width: 160px"
            clearable
            :placeholder="$t('common.status')"
            @change="search"
            @clear="search"
          >
            <el-option :label="$t('common.enabled')" value="1" />
            <el-option :label="$t('common.disabled')" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('user.610wkzjilwo0')">
          <el-select
            v-model="params.syncState"
            style="width: 160px"
            clearable
            :placeholder="$t('user.610wkzjilwo0')"
            @change="search"
            @clear="search"
          >
            <el-option :label="$t('user.610wkzjilz00')" value="1" />
            <el-option :label="$t('user.610wkzjim1c0')" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" :icon="Search" @click="search">{{ $t('common.query') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" :icon="Plus" type="primary" @click="create">{{ $t('common.add') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            :icon="Delete"
            type="danger"
            @click="batchDelete"
          >{{ $t('user.610wkzjim8k0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            :icon="Upload"
            type="success"
            @click="batchSync"
          >{{ $t('user.610wkzjimb40') }}</el-button>
        </el-form-item>
        <br>
        <el-form-item>
          <el-button :loading="loading" :icon="Download" @click="runSync(syncOpenLdapUsersApi)">{{ $t('user.610wkzjimd80') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" :icon="Download" @click="runSync(syncDingTalkUsersApi)">{{ $t('user.610wkzjimfk0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" :icon="Download" @click="runSync(syncFeiShuUsersApi)">{{ $t('user.610wkzjimho0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" :icon="Download" @click="runSync(syncWeComUsersApi)">{{ $t('user.610wkzjimk00') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="loading"
        :data="tableData"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="48" align="center" />
        <!-- 展开列不写表头文字（窄列写字会把表头撑高），点箭头查看详情 -->
        <el-table-column type="expand" width="44">
          <template #default="props">
            <RowDetail
              :columns="3"
              :items="[
                { label: 'User DN', value: props.row.userDn, wide: true },
                { label: $t('user.610wkzjin800'), value: (props.row.groups || []).map(g => g.groupDn).join('\n'), wide: true },
                { label: $t('user.610wkzjimqk0'), value: props.row.mail },
                { label: $t('user.610wkzjimog0'), value: props.row.mobile },
                { label: $t('user.610wkzjimsw0'), value: props.row.jobNumber },
                { label: $t('user.610wkzjimys0'), value: props.row.position },
                { label: $t('user.610wkzjimwo0'), value: props.row.creator },
                { label: $t('user.610wkzjin0w0'), value: props.row.introduction }
              ]"
            />
          </template>
        </el-table-column>
        <el-table-column sortable prop="username" :label="$t('user.610wkzjikzs0')" width="140">
          <template #default="scope">
            <!-- 悬停显示 User DN -->
            <el-tooltip :content="scope.row.userDn" :disabled="!scope.row.userDn" placement="top">
              <el-tag size="default" class="username-tag">{{ scope.row.username }}</el-tag>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.status')" width="90" align="center">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              size="small"
              :active-value="1"
              :inactive-value="2"
              @change="userStateChanged(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="nickname" :label="$t('user.610wkzjill40')" min-width="130" />
        <el-table-column show-overflow-tooltip prop="givenName" :label="$t('user.610wkzjin4g0')" min-width="120" />
        <el-table-column :label="$t('user.610wkzjin680')" min-width="140">
          <template #default="scope">
            <div class="cell-tags">
              <el-tag v-for="role in scope.row.roles" :key="role.ID" size="small" effect="plain">{{ role.name }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <!-- 部门：单行显示组名标签（悬停看 DN），超过 2 个折叠成 +N；完整 DN 在展开详情里 -->
        <el-table-column :label="$t('user.610wkzjin800')" min-width="220">
          <template #default="scope">
            <div v-if="scope.row.groups && scope.row.groups.length" class="cell-tags cell-tags--nowrap">
              <el-tooltip
                v-for="group in scope.row.groups.slice(0, 2)"
                :key="group.ID"
                :content="group.groupDn"
                placement="top"
              >
                <el-tag size="small" type="info" effect="plain">{{ groupLabel(group) }}</el-tag>
              </el-tooltip>
              <el-popover v-if="scope.row.groups.length > 2" placement="bottom" :width="320" trigger="hover">
                <template #reference>
                  <el-tag size="small" type="info" class="cell-tags__more">+{{ scope.row.groups.length - 2 }}</el-tag>
                </template>
                <ul class="group-popover">
                  <li v-for="group in scope.row.groups" :key="group.ID">
                    <span class="group-popover__name">{{ groupLabel(group) }}</span>
                    <span class="group-popover__dn">{{ group.groupDn }}</span>
                  </li>
                </ul>
              </el-popover>
            </div>
            <span v-else class="cell-empty">—</span>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="CreatedAt" :label="$t('common.createTime')" width="180" />
        <el-table-column show-overflow-tooltip sortable prop="UpdatedAt" :label="$t('common.updateTime')" width="180" />
        <el-table-column fixed="right" :label="$t('common.management')" align="left" width="230">
          <template #default="scope">
            <div class="row-actions">
              <el-button link type="primary" :icon="Edit" @click="update(scope.row)">{{ $t('common.edit') }}</el-button>
              <el-popconfirm :title="$t('user.610wkzjinjg0')" @confirm="singleDelete(scope.row.ID)">
                <template #reference>
                  <el-button link type="danger" :icon="Delete">{{ $t('common.delete') }}</el-button>
                </template>
              </el-popconfirm>
              <el-popconfirm v-if="scope.row.syncState === 2 && scope.row.status === 1" :title="$t('user.610wkzjinn40')" @confirm="singleSync(scope.row.ID)">
                <template #reference>
                  <el-button link type="primary" :icon="Upload">{{ $t('user.610wkzjinlc0') }}</el-button>
                </template>
              </el-popconfirm>
            </div>
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
        style="margin-top: 10px; float: right; margin-bottom: 10px"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />

      <el-dialog v-model="dialogFormVisible" :title="dialogFormTitle" width="720px" @closed="resetForm">
        <el-form
          ref="dialogFormRef"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-position="top"
        >
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjikzs0')" prop="username">
                <el-input
                  v-model.trim="dialogFormData.username"
                  :disabled="disabled"
                  :placeholder="$t('user.610wkzjinp00')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjill40')" prop="nickname">
                <el-input v-model.trim="dialogFormData.nickname" :placeholder="$t('user.610wkzjill40')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjin4g0')" prop="givenName">
                <el-input v-model.trim="dialogFormData.givenName" :placeholder="$t('user.610wkzjin4g0')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjinso0')" prop="avatar">
                <el-input v-model.trim="dialogFormData.avatar" :placeholder="$t('user.imageTips')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjimqk0')" prop="mail">
                <el-input v-model.trim="dialogFormData.mail" :placeholder="$t('user.610wkzjimqk0')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item
                :label="dialogType === 'create' ? $t('user.610wkzjinuk0') : $t('user.610wkzjinyc0')"
                prop="password"
              >
                <el-input
                  v-model.trim="dialogFormData.password"
                  autocomplete="off"
                  :type="passwordType"
                  :placeholder="dialogType === 'create' ? $t('user.610wkzjinwc0') : $t('user.610wkzjio000')"
                >
                  <template #suffix>
                    <span class="show-pwd" @click="showPwd">
                      <component :is="passwordType === 'password' ? IconEyeOff : IconEye" :size="18" stroke="1.75" />
                    </span>
                  </template>
                </el-input>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjin680')" prop="roleIds">
                <el-select
                  v-model="dialogFormData.roleIds"
                  multiple
                  :placeholder="$t('user.610wkzjio200')"
                  style="width: 100%"
                >
                  <el-option v-for="item in roles" :key="item.ID" :label="item.name" :value="item.ID" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('common.status')" prop="status">
                <el-select v-model="dialogFormData.status" :placeholder="$t('user.610wkzjio3s0')" style="width: 100%">
                  <el-option :label="$t('common.enabled')" :value="1" />
                  <el-option :label="$t('common.disabled')" :value="2" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjio6c0')" prop="mobile">
                <el-input v-model.trim="dialogFormData.mobile" :placeholder="$t('user.610wkzjio6c0')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjimsw0')" prop="jobNumber">
                <el-input v-model.trim="dialogFormData.jobNumber" :placeholder="$t('user.610wkzjimsw0')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjimys0')" prop="position">
                <el-input v-model.trim="dialogFormData.position" :placeholder="$t('user.610wkzjio8g0')" />
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item :label="$t('user.610wkzjioa80')" prop="groupIds">
                <TreeSelect
                  v-model="dialogFormData.groupIds"
                  :options="groupsOptions"
                  :placeholder="$t('user.610wkzjioc00')"
                  :normalizer="normalizer"
                  value-consists-of="ALL"
                  :multiple="true"
                  :flat="true"
                  :no-children-text="$t('user.610wkzjiods0')"
                  :no-results-text="$t('user.610wkzjiofk0')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item :label="$t('user.610wkzjiohc0')" prop="postalAddress">
                <el-input
                  v-model.trim="dialogFormData.postalAddress"
                  type="textarea"
                  :placeholder="$t('user.610wkzjiohc0')"
                  :autosize="{ minRows: 3, maxRows: 6 }"
                  show-word-limit
                  maxlength="100"
                />
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item :label="$t('user.610wkzjioj40')" prop="introduction">
                <el-input
                  v-model.trim="dialogFormData.introduction"
                  type="textarea"
                  :placeholder="$t('user.610wkzjiol00')"
                  :autosize="{ minRows: 3, maxRows: 6 }"
                  show-word-limit
                  maxlength="100"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
        <template #footer>
          <div class="dialog-footer">
            <el-switch v-model="notice" :active-text="$t('user.610wkzjiook0')" />
            <el-button @click="cancelForm">{{ $t('common.cancel') }}</el-button>
            <el-button :loading="submitLoading" type="primary" @click="submitForm">{{ $t('common.confirm') }}</el-button>
          </div>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { Search, Plus, Delete, Upload, Download, Edit, IconEye, IconEyeOff } from '@/icons/tabler';
import TreeSelect from '@zanmato/vue3-treeselect';
import RowDetail from '@/components/RowDetail/index.vue';
import '@zanmato/vue3-treeselect/dist/vue3-treeselect.min.css';
import JSEncrypt from 'jsencrypt';
import { getGroupTree } from '@/api/personnel/group';
import {
  batchDeleteUserByIds,
  changeUserStatus,
  createUser,
  getUsers,
  syncDingTalkUsersApi,
  syncFeiShuUsersApi,
  syncOpenLdapUsersApi,
  syncSqlUsers,
  syncWeComUsersApi,
  updateUserById
} from '@/api/personnel/user';
import { getRoles as getRolesApi } from '@/api/system/role';

defineOptions({ name: 'User' });

const { t } = useI18n();

// 部门标签显示组名；没有组名时取 DN 的第一段值（cn=xxx / ou=xxx）
function groupLabel(group: { groupName?: string; groupDn?: string }) {
  if (group.groupName) return group.groupName;
  const first = (group.groupDn || '').split(',')[0] || '';
  return first.includes('=') ? first.split('=')[1] : first;
}

const publicKey = import.meta.env.VITE_PUBLIC_KEY;

// username is not editable when updating (backend ldap-user-name-modify config)
const disabled = ref(false);

const params = reactive({
  username: '',
  nickname: '',
  status: '',
  syncState: '',
  pageNum: 1,
  pageSize: 10
});

const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);
const roles = ref<any[]>([]);
const groupsOptions = ref<any[]>([]);
const passwordType = ref('password');
const notice = ref(true);
const submitLoading = ref(false);
const dialogFormTitle = ref('');
const dialogType = ref('');
const dialogFormVisible = ref(false);
const dialogFormRef = ref<FormInstance>();
const multipleSelection = ref<any[]>([]);

const emptyForm = () => ({
  ID: '' as any,
  mail: '',
  givenName: '',
  username: '',
  password: '',
  nickname: '',
  status: 1 as any,
  mobile: '',
  avatar: '',
  introduction: '',
  roleIds: [] as any[],
  jobNumber: '',
  position: '',
  postalAddress: '',
  groupIds: undefined as any,
  notice: true
});
const dialogFormData = ref(emptyForm());

const checkPhone = (_rule, value, callback) => {
  if (value) {
    const reg = /^(\+|00)??(\d{1,3})??((1|0)\d{8,10})??$/;
    if (reg.test(value)) {
      callback();
    } else {
      callback(new Error(t('user.610wkzjioxc0')));
    }
  } else {
    callback();
  }
};

const checkPasswordComplexity = (value: string) => {
  let matches = 0;
  if (/[a-z]/.test(value)) matches++;
  if (/[A-Z]/.test(value)) matches++;
  if (/\d/.test(value)) matches++;
  if (/[!@#$%^&*,.?_-]/.test(value)) matches++;
  return matches;
};

const validatePasswordCanEmpty = (_rule, value, callback) => {
  if (!value) return callback();
  if (value.length < 8) return callback(new Error(t('loginform.newpassValidErrLenMustThan8')));
  if (checkPasswordComplexity(value) < 3) return callback(new Error(t('loginform.newpassValidErrMustComplex')));
  callback();
};

const dialogFormRules: FormRules = {
  username: [
    { required: true, message: t('user.610wkzjioz80'), trigger: 'blur' },
    { min: 2, max: 20, message: t('valid.length', [2, 20]), trigger: 'blur' }
  ],
  password: [{ required: false, validator: validatePasswordCanEmpty, trigger: 'blur' }],
  mail: [{ required: true, message: t('user.610wkzjip140'), trigger: 'blur' }],
  jobNumber: [
    { required: false, message: t('user.610wkzjip380'), trigger: 'blur' },
    { min: 0, max: 20, message: t('valid.length', [0, 20]), trigger: 'blur' }
  ],
  nickname: [
    { required: true, message: t('user.610wkzjip540'), trigger: 'blur' },
    { min: 2, max: 20, message: t('valid.length', [2, 20]), trigger: 'blur' }
  ],
  mobile: [{ required: false, validator: checkPhone, trigger: 'blur' }],
  status: [{ required: true, message: t('user.610wkzjio3s0'), trigger: 'change' }],
  groupIds: [{ required: false, message: t('user.610wkzjioc00'), trigger: 'blur' }],
  introduction: [
    { required: false, message: t('user.610wkzjioj40'), trigger: 'blur' },
    { min: 0, max: 100, message: t('valid.length', [0, 100]), trigger: 'blur' }
  ]
};

// show success message (backend returns message in data)
const judgeResult = (res) => {
  const message = res && res.data ? res.data : t('user.610wkzjipcg0');
  ElMessage({ showClose: true, message, type: 'success' });
};

const search = () => {
  params.pageNum = 1;
  getTableData();
};

const getTableData = async () => {
  loading.value = true;
  try {
    const { data } = await getUsers(params);
    const users = data.users || [];
    users.forEach((item) => {
      item.groupIds = (item.groups || []).map((g) => +g.ID);
    });
    tableData.value = users;
    total.value = data.total;
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
};

// all groups for the tree select
const getAllGroups = async () => {
  loading.value = true;
  try {
    const { data } = await getGroupTree({ pageNum: 1, pageSize: 1000 });
    groupsOptions.value = [
      { ID: 0, groupName: t('user.610wkzjip6w0'), groupType: 'T', children: data }
    ];
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
};

const getRoles = async () => {
  try {
    const { data } = await getRolesApi(null);
    roles.value = data.roles;
  } catch {
    // error already shown
  }
};

const create = () => {
  dialogFormTitle.value = t('user.610wkzjip8w0');
  dialogType.value = 'create';
  disabled.value = false;
  getAllGroups();
  dialogFormData.value = emptyForm();
  dialogFormData.value.roleIds = [2];
  notice.value = true;
  dialogFormVisible.value = true;
};

const update = (row) => {
  dialogFormTitle.value = t('user.610wkzjipao0');
  dialogType.value = 'update';
  disabled.value = true;
  passwordType.value = 'password';
  getAllGroups();
  dialogFormData.value = {
    ...emptyForm(),
    ID: row.ID,
    mail: row.mail,
    givenName: row.givenName,
    username: row.username,
    password: '',
    nickname: row.nickname,
    status: row.status,
    mobile: row.mobile,
    avatar: row.avatar,
    introduction: row.introduction,
    roleIds: (row.roles || []).map((item) => item.ID),
    jobNumber: row.jobNumber,
    position: row.position,
    postalAddress: row.postalAddress,
    groupIds: row.groupIds
  };
  notice.value = false;
  dialogFormVisible.value = true;
};

const showError = (message: string) => {
  ElMessage({ showClose: true, message, type: 'error' });
};

const submitForm = async () => {
  const form = dialogFormData.value;
  if (form.nickname === '') return showError(t('user.610wkzjipec0'));
  if (form.username === '') return showError(t('user.610wkzjipg00'));
  if (form.mail === '') return showError(t('user.610wkzjipi40'));
  if (form.status === '') return showError(t('user.610wkzjipk40'));
  form.notice = notice.value;
  if (form.roleIds === ('' as any)) return showError(t('user.610wkzjipn40'));

  try {
    await dialogFormRef.value?.validate();
  } catch {
    ElMessage({ showClose: true, message: t('user.610wkzjipp00'), type: 'warning' });
    return;
  }

  submitLoading.value = true;
  const payload: any = { ...form };
  if (form.password !== '') {
    // RSA-encrypt password
    const encryptor = new JSEncrypt();
    encryptor.setPublicKey(publicKey);
    payload.password = encryptor.encrypt(form.password);
  }
  try {
    const res = dialogType.value === 'create'
      ? await createUser(payload)
      : await updateUserById(payload);
    judgeResult(res);
    resetForm();
    getTableData();
  } catch {
    // error already shown
  } finally {
    submitLoading.value = false;
  }
};

const cancelForm = () => {
  resetForm();
};

const resetForm = () => {
  dialogFormVisible.value = false;
  dialogFormRef.value?.resetFields();
  dialogFormData.value = emptyForm();
};

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(t('user.batchDeleteTips'), t('common.prompt'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    });
  } catch {
    ElMessage({ showClose: true, type: 'info', message: t('user.610wkzjipy40') });
    return;
  }
  loading.value = true;
  try {
    const res = await batchDeleteUserByIds({ userIds: multipleSelection.value.map((x) => x.ID) });
    judgeResult(res);
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
  getTableData();
};

const batchSync = async () => {
  try {
    await ElMessageBox.confirm(t('user.betchSyncLdapTips'), t('common.prompt'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    });
  } catch {
    ElMessage({ showClose: true, type: 'info', message: t('user.610wkzjiq080') });
    return;
  }
  loading.value = true;
  try {
    const res = await syncSqlUsers({ userIds: multipleSelection.value.map((x) => x.ID) });
    judgeResult(res);
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
  getTableData();
};

// switch toggled: revert on failure
const userStateChanged = async (row) => {
  try {
    await changeUserStatus({ id: row.ID, status: row.status });
    ElMessage.success(t('user.610wkzjiq3w0'));
  } catch {
    row.status = row.status === 1 ? 2 : 1;
  }
};

const handleSelectionChange = (val) => {
  multipleSelection.value = val;
};

const singleDelete = async (id) => {
  loading.value = true;
  try {
    const res = await batchDeleteUserByIds({ userIds: [id] });
    judgeResult(res);
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
  getTableData();
};

const singleSync = async (id) => {
  loading.value = true;
  try {
    const res = await syncSqlUsers({ userIds: [id] });
    judgeResult(res);
  } catch {
    // error already shown
  } finally {
    loading.value = false;
  }
  getTableData();
};

const showPwd = () => {
  passwordType.value = passwordType.value === 'password' ? '' : 'password';
};

const handleSizeChange = (val) => {
  params.pageSize = val;
  getTableData();
};
const handleCurrentChange = (val) => {
  params.pageNum = val;
  getTableData();
};

// treeselect
const normalizer = (node) => ({
  id: node.ID,
  label: node.groupType + '=' + node.groupName,
  isDisabled: node.groupType === 'ou' || node.groupName === 'root',
  children: node.children
});

// third-party sync (OpenLDAP / DingTalk / FeiShu / WeCom)
const runSync = async (api: (data?: unknown) => Promise<any>) => {
  loading.value = true;
  try {
    const res = await api(undefined);
    judgeResult(res);
  } catch {
    loading.value = false;
    return; // error already shown
  }
  loading.value = false;
  getTableData();
};

onMounted(() => {
  getTableData();
  getRoles();
});
</script>

<style lang="scss" scoped>
.username-tag {
  max-width: 100%;
  cursor: default;
}

.group-popover {
  margin: 0;
  padding: 0;
  list-style: none;

  li {
    display: flex;
    flex-direction: column;
    padding: 6px 0;
    border-bottom: 1px solid var(--el-border-color-lighter);

    &:last-child {
      border-bottom: 0;
    }
  }
}

.group-popover__name {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.group-popover__dn {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
}

.container-card {
  margin: 10px;
  margin-bottom: 100px;
}

.delete-popover {
  margin-left: 10px;
}

.show-pwd {
  display: inline-flex;
  align-items: center;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  user-select: none;
}

</style>
