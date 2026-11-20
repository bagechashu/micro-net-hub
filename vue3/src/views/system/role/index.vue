<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="small" :inline="true" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('role.611p12potc80')">
          <el-input v-model.trim="params.name" clearable :placeholder="$t('role.611p12potc80')" @keyup.enter="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('role.611p12pov3g0')">
          <el-input v-model.trim="params.keyword" clearable :placeholder="$t('role.611p12pov3g0')" @keyup.enter="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('role.611p12pov9c0')">
          <el-select v-model="params.status" clearable :placeholder="$t('role.611p12pov9c0')" @change="search" @clear="search">
            <el-option :label="$t('role.611p12povd80')" :value="1" />
            <el-option :label="$t('role.611p12povgg0')" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" type="primary" @click="search"><el-icon><SearchIcon /></el-icon>{{ $t('role.611p12povjw0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" type="warning" @click="create"><el-icon><Plus /></el-icon>{{ $t('role.611p12povn40') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :disabled="multipleSelection.length === 0" :loading="loading" type="danger" @click="batchDelete"><el-icon><Delete /></el-icon>{{ $t('role.611p12povqk0') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column show-overflow-tooltip sortable prop="name" :label="$t('role.611p12potc80')" />
        <el-table-column show-overflow-tooltip sortable prop="keyword" :label="$t('role.611p12pov3g0')" />
        <el-table-column show-overflow-tooltip sortable prop="sort" :label="$t('role.611p12povu40')" />
        <el-table-column show-overflow-tooltip sortable prop="status" :label="$t('role.611p12pov9c0')" align="center">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.status === 1 ? 'success':'danger'" disable-transitions>{{ scope.row.status === 1 ? $t('role.611p12povd80'):$t('role.611p12povgg0') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="creator" :label="$t('role.611p12pow400')" />
        <el-table-column show-overflow-tooltip sortable prop="remark" :label="$t('role.611p12powdg0')" />
        <el-table-column fixed="right" :label="$t('role.611p12powhw0')" align="center" width="140">
          <template #default="scope">
            <el-tooltip :content="$t('role.611p12powkk0')" effect="dark" placement="top">
              <el-button size="small" circle type="primary" @click="update(scope.row)" ><el-icon><Edit /></el-icon></el-button>
            </el-tooltip>
            <el-tooltip :content="$t('role.611p12pown80')" effect="dark" placement="top">
              <el-button size="small" circle type="warning" @click="updatePermission(scope.row.ID)" ><el-icon><Lock /></el-icon></el-button>
            </el-tooltip>
            <el-tooltip :content="$t('role.611p12powpk0')" effect="dark" placement="top">
              <span>
                <el-popconfirm style="margin-left:10px" :title="$t('role.611p12powrw0')" @confirm="singleDelete(scope.row.ID)">
                  <template #reference>
                    <el-button size="small" circle type="danger" ><el-icon><Delete /></el-icon></el-button>
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

      <el-dialog :title="dialogFormTitle" v-model="dialogFormVisible" width="580px">
        <el-form ref="dialogForm" :inline="true" size="small" :model="dialogFormData" :rules="dialogFormRules" label-width="100px">
          <el-form-item :label="$t('role.611p12potc80')" prop="name">
            <el-input v-model.trim="dialogFormData.name" :placeholder="$t('role.611p12potc80')" style="width: 420px" />
          </el-form-item>
          <el-form-item :label="$t('role.611p12pov3g0')" prop="keyword">
            <el-input v-model.trim="dialogFormData.keyword" :placeholder="$t('role.611p12pov3g0')" style="width: 420px" />
          </el-form-item>
          <el-form-item :label="$t('role.611p12pov9c0')" prop="status">
            <el-select v-model.trim="dialogFormData.status" :placeholder="$t('role.611p12powu00')" style="width: 180px">
              <el-option :label="$t('role.611p12povd80')" :value="1" />
              <el-option :label="$t('role.611p12povgg0')" :value="2" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('role.611p12powwk0')" prop="sort">
            <el-input-number v-model.number="dialogFormData.sort" controls-position="right" :min="1" :max="999" />
          </el-form-item>
          <el-form-item :label="$t('role.611p12powdg0')" prop="remark">
            <el-input v-model.trim="dialogFormData.remark" style="width: 420px" type="textarea" :placeholder="$t('role.611p12powdg0')" show-word-limit maxlength="100" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button size="small" @click="cancelForm()">{{ $t('role.611p12powys0') }}</el-button>
          <el-button size="small" :loading="submitLoading" type="primary" @click="submitForm()">{{ $t('role.611p12pox180') }}</el-button>
        </template>
      </el-dialog>

      <el-dialog :title="$t('role.611p12pox540')" v-model="permsDialogVisible" width="580px" custom-class="perms-dialog">
        <el-tabs>
          <el-tab-pane>
            <template #label><svg-icon icon-class="tree-table" class-name="role-menu" /> {{ $t('role.611p12pox740') }}</template>
            <el-tree
              ref="roleMenuTree"
              v-loading="menuTreeLoading"
              :props="{children: 'children',label: 'name'}"
              :data="menuTree"
              show-checkbox
              node-key="ID"
              check-strictly
              :default-checked-keys="defaultCheckedRoleMenu"
            />
          </el-tab-pane>

          <el-tab-pane>
            <template #label><svg-icon icon-class="tree" class-name="role-menu" /> {{ $t('role.611p12pox9c0') }}</template>
            <el-tree
              ref="roleApiTree"
              v-loading="apiTreeLoading"
              :props="{children: 'children',label: 'remark'}"
              :data="apiTree"
              show-checkbox
              node-key="ID"
              :default-checked-keys="defaultCheckedRoleApi"
            />
          </el-tab-pane>
        </el-tabs>
        <template #footer>
          <el-button size="small" :loading="permissionLoading" @click="cancelPermissionForm()">{{ $t('role.611p12powys0') }}</el-button>
          <el-button size="small" type="primary" @click="submitPermissionForm()">{{ $t('role.611p12pox180') }}</el-button>
        </template>
      </el-dialog>

    </el-card>
  </div>
</template>

<script>
import { defineComponent } from 'vue';
import { getRoles, createRole, updateRoleById, batchDeleteRoleByIds, getRoleMenusById, getRoleApisById, updateRoleMenusById, updateRoleApisById } from '@/api/system/role';
import { getMenuTree } from '@/api/system/menu';
import { getApiTree } from '@/api/system/api';
import { Search, Plus, Delete, Edit, Lock } from '@element-plus/icons-vue';

export default defineComponent({
  name: 'Role',
  components: {
    SearchIcon: Search,
    Plus,
    Delete,
    Edit,
    Lock
  },
  data() {
    return {
      params: {
        name: '',
        keyword: '',
        status: '',
        pageNum: 1,
        pageSize: 10
      },
      tableData: [],
      total: 0,
      loading: false,
      submitLoading: false,
      dialogFormTitle: '',
      dialogType: '',
      dialogFormVisible: false,
      dialogFormData: {
        ID: '',
        name: '',
        keyword: '',
        status: 1,
        sort: 999,
        remark: ''
      },
      dialogFormRules: {
        name: [
          { required: true, message: this.$t('role.611p12poxb40'), trigger: 'blur' },
          { min: 1, max: 20, message: this.$t('valid.length', [1, 20]), trigger: 'blur' }
        ],
        keyword: [
          { required: true, message: this.$t('role.611p12poxcw0'), trigger: 'blur' },
          { min: 1, max: 20, message: this.$t('valid.length', [1, 20]), trigger: 'blur' }
        ],
        status: [
          { required: true, message: this.$t('role.611p12powu00'), trigger: 'change' }
        ],
        remark: [
          { required: false, message: this.$t('role.611p12powdg0'), trigger: 'blur' },
          { min: 0, max: 100, message: this.$t('valid.length', [0, 100]), trigger: 'blur' }
        ]
      },
      popoverVisible: false,
      multipleSelection: [],
      permsDialogVisible: false,
      permissionLoading: false,
      menuTree: [],
      defaultCheckedRoleMenu: [],
      apiTree: [],
      defaultCheckedRoleApi: [],
      roleId: 0,
      menuTreeLoading: false,
      apiTreeLoading: false
    };
  },
  created() {
    this.getTableData();
    this.getMenuTree();
    this.getApiTree();
  },
  methods: {
    search() {
      this.params.pageNum = 1;
      this.getTableData();
    },
    async getTableData() {
      this.loading = true;
      try {
        const { data } = await getRoles(this.params);
        this.tableData = data.roles;
        this.total = data.total;
      } finally {
        this.loading = false;
      }
    },
    create() {
      this.dialogFormTitle = this.$t('role.611p12poxf40');
      this.dialogType = 'create';
      this.dialogFormVisible = true;
    },
    update(row) {
      this.dialogFormData.ID = row.ID;
      this.dialogFormData.name = row.name;
      this.dialogFormData.keyword = row.keyword;
      this.dialogFormData.sort = row.sort;
      this.dialogFormData.status = row.status;
      this.dialogFormData.remark = row.remark;

      this.dialogFormTitle = this.$t('role.611p12poxgw0');
      this.dialogType = 'update';
      this.dialogFormVisible = true;
    },
    judgeResult(res) {
      if (res.code === 200 || res.code === 0) {
        ElMessage({
          showClose: true,
          message: this.$t('role.611p12poxik0'),
          type: 'success'
        });
      }
    },
    submitForm() {
      // @ts-expect-error refs
      this.$refs['dialogForm'].validate(async (valid) => {
        if (valid) {
          this.submitLoading = true;
          try {
            if (this.dialogType === 'create') {
              await createRole(this.dialogFormData).then(res => {
                this.judgeResult(res);
              });
            } else {
              await updateRoleById(this.dialogFormData).then(res => {
                this.judgeResult(res);
              });
            }
          } finally {
            this.submitLoading = false;
          }
          this.resetForm();
          this.getTableData();
        }
      });
    },
    cancelForm() {
      this.resetForm();
    },
    resetForm() {
      this.dialogFormVisible = false;
      // @ts-expect-error refs
      this.$refs['dialogForm'].resetFields();
      this.dialogFormData = {
        name: '',
        keyword: '',
        status: 1,
        sort: 999,
        remark: ''
      };
    },
    batchDelete() {
      ElMessageBox.confirm(this.$t('tips.deleteWarning'), this.$t('role.611p12poxkg0'), {
        confirmButtonText: this.$t('role.611p12pox180'),
        cancelButtonText: this.$t('role.611p12powys0'),
        type: 'warning'
      }).then(async () => {
        this.loading = true;
        const roleIds = [];
        this.multipleSelection.forEach(x => {
          roleIds.push(x.ID);
        });
        try {
          await batchDeleteRoleByIds({ roleIds: roleIds }).then(res => {
            this.judgeResult(res);
          });
        } finally {
          this.loading = false;
        }
        this.getTableData();
      }).catch(() => {
        ElMessage({
          type: 'info',
          message: this.$t('role.611p12poxm80')
        });
      });
    },
    handleSelectionChange(val) {
      this.multipleSelection = val;
    },
    async singleDelete(id) {
      this.loading = true;
      try {
        await batchDeleteRoleByIds({ roleIds: [id] }).then(res => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
    },
    async updatePermission(roleId) {
      this.roleId = roleId;
      this.permsDialogVisible = true;
      this.getMenuTree();
      this.getApiTree();
      this.getRoleMenusById(roleId);
      this.getRoleApisById(roleId);
    },
    async getMenuTree() {
      this.menuTreeLoading = true;
      try {
        const { data } = await getMenuTree();
        this.menuTree = data;
      } finally {
        this.menuTreeLoading = false;
      }
    },
    async getApiTree() {
      this.apiTreeLoading = true;
      try {
        const { data } = await getApiTree();
        this.apiTree = data;
      } finally {
        this.apiTreeLoading = false;
      }
    },
    async getRoleMenusById(roleId) {
      this.permissionLoading = true;
      let rseData;
      const params = {};
      params.roleId = roleId;
      try {
        const { data } = await getRoleMenusById(params);
        rseData = data;
      } finally {
        this.permissionLoading = false;
      }

      const menus = rseData;
      const ids = [];
      menus.forEach(x => { ids.push(x.ID); });
      this.defaultCheckedRoleMenu = ids;
      // @ts-expect-error refs
      this.$refs.roleMenuTree.setCheckedKeys(this.defaultCheckedRoleMenu);
    },
    async getRoleApisById(roleId) {
      this.permissionLoading = true;
      let rseData;
      const params = {};
      params.roleId = roleId;
      try {
        const { data } = await getRoleApisById(params);
        rseData = data;
      } finally {
        this.permissionLoading = false;
      }
      const apis = rseData;
      const ids = [];
      apis.forEach(x => { ids.push(x.ID); });
      this.defaultCheckedRoleApi = ids;
      // @ts-expect-error refs
      this.$refs.roleApiTree.setCheckedKeys(this.defaultCheckedRoleApi);
    },
    cancelPermissionForm() {
      this.permsDialogVisible = false;
      this.defaultCheckedRoleApi = [];
      this.defaultCheckedRoleMenu = [];
    },
    submitPermissionForm() {
      // @ts-expect-error refs
      const menuCheckedKeys = this.$refs.roleMenuTree.getCheckedKeys(true);
      // @ts-expect-error refs
      const apiCheckedKeys = this.$refs.roleApiTree.getCheckedKeys(true);
      const menus = menuCheckedKeys.checked;
      const apis = apiCheckedKeys.checked;
      this.permissionLoading = true;
      try {
        updateRoleMenusById({ roleId: this.roleId, menuIds: menus });
        updateRoleApisById({ roleId: this.roleId, apiIds: apis });
        ElMessage({ message: this.$t('role.611p12poxik0'), type: 'success' });
      } finally {
        this.permissionLoading = false;
        this.permsDialogVisible = false;
        this.defaultCheckedRoleApi = [];
        this.defaultCheckedRoleMenu = [];
      }
    },
    handleSizeChange(size) {
      this.params.pageSize = size;
      this.getTableData();
    },
    handleCurrentChange(page) {
      this.params.pageNum = page;
      this.getTableData();
    }
  }
});
</script>

<style scoped >
  .container-card{
    margin: 10px;
    margin-bottom: 100px;
  }

  .role-menu{
    font-size: 1rem;
    margin-right: 5px;
  }
</style>

<style lang="scss">
  .perms-dialog > .el-dialog__body{
    padding-top: 0;
    padding-bottom: 15px;
  }
</style>
