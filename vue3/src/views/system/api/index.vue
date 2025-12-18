<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="small" :inline="true" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('api.611p62ddlio0')">
          <el-input v-model.trim="params.path" clearable :placeholder="$t('api.611p62ddlio0')" @keyup.enter="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('api.611p62ddlz40')">
          <el-input v-model.trim="params.category" clearable :placeholder="$t('api.611p62ddlz40')" @keyup.enter="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('api.611p62ddm380')">
          <el-select v-model="params.method" clearable :placeholder="$t('api.611p62ddm6c0')" @change="search" @clear="search">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="PATCH" value="PATCH" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('api.611p62ddm980')">
          <el-input v-model.trim="params.creator" clearable :placeholder="$t('api.611p62ddm980')" @keyup.enter="search" @clear="search" />
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" type="primary" @click="search">
            <el-icon><Search /></el-icon>
            {{ $t('api.611p62ddmdc0') }}
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" type="warning" @click="create">
            <el-icon><Plus /></el-icon>
            {{ $t('api.611p62ddmg40') }}
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button :disabled="multipleSelection.length === 0" :loading="loading" type="danger" @click="batchDelete">
            <el-icon><Delete /></el-icon>
            {{ $t('api.611p62ddmis0') }}
          </el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column show-overflow-tooltip sortable prop="path" :label="$t('api.611p62ddlio0')" />
        <el-table-column show-overflow-tooltip sortable prop="category" :label="$t('api.611p62ddlz40')" />
        <el-table-column show-overflow-tooltip sortable prop="method" :label="$t('api.611p62ddm6c0')" align="center">
          <template #default="scope">
            <el-tag size="small" :type="methodTagFilter(scope.row.method)" disable-transitions>{{ scope.row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="creator" :label="$t('api.611p62ddm980')" />
        <el-table-column show-overflow-tooltip sortable prop="remark" :label="$t('api.611p62ddmlo0')" />
        <el-table-column fixed="right" :label="$t('api.611p62ddmoc0')" align="center" width="120">
          <template #default="scope">
            <el-tooltip :content="$t('api.611p62ddmr40')" effect="dark" placement="top">
              <template #default>
                <el-button size="small" circle type="primary" @click="update(scope.row)">
                  <el-icon><Edit /></el-icon>
                </el-button>
              </template>
            </el-tooltip>
            <el-tooltip class="delete-popover" :content="$t('api.611p62ddmu00')" effect="dark" placement="top">
              <template #default>
                <el-popconfirm :title="$t('api.611p62ddmwo0')" @confirm="singleDelete(scope.row.ID)">
                  <template #reference>
                    <el-button size="small" circle type="danger">
                      <el-icon><Delete /></el-icon>
                </el-button>
                  </template>
                </el-popconfirm>
              </template>
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

      <el-dialog :title="dialogFormTitle" v-model="dialogFormVisible">
        <el-form ref="dialogForm" size="small" :model="dialogFormData" :rules="dialogFormRules" label-width="120px">
          <el-form-item :label="$t('api.611p62ddlio0')" prop="path">
            <el-input v-model.trim="dialogFormData.path" :placeholder="$t('api.611p62ddlio0')" />
          </el-form-item>
          <el-form-item :label="$t('api.611p62ddlz40')" prop="category">
            <el-input v-model.trim="dialogFormData.category" :placeholder="$t('api.611p62ddlz40')" />
          </el-form-item>
          <el-form-item :label="$t('api.611p62ddm6c0')" prop="method">
            <el-select v-model.trim="dialogFormData.method" :placeholder="$t('api.611p62ddmzc0')">
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
              <el-option label="PUT" value="PUT" />
              <el-option label="PATCH" value="PATCH" />
              <el-option label="DELETE" value="DELETE" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('api.611p62ddmlo0')" prop="remark">
            <el-input v-model.trim="dialogFormData.remark" type="textarea" :placeholder="$t('api.611p62ddmlo0')" show-word-limit maxlength="100" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button size="small" @click="cancelForm()">{{ $t('api.611p62ddn240') }}</el-button>
          <el-button size="small" :loading="submitLoading" type="primary" @click="submitForm()">{{ $t('api.611p62ddn4s0') }}</el-button>
        </template>
      </el-dialog>

    </el-card>
  </div>
</template>

<script>
import { defineComponent } from 'vue';
import { getApis, createApi, updateApiById, batchDeleteApiByIds } from '@/api/system/api';
import { Search, Plus, Delete, Edit } from '@element-plus/icons-vue';

export default defineComponent({
  name: 'Api',
  components: {
    Search,
    Plus,
    Delete,
    Edit
  },
  data() {
    return {
      params: {
        path: '',
        method: '',
        category: '',
        creator: '',
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
        path: '',
        category: '',
        method: '',
        remark: ''
      },
      dialogFormRules: {
        path: [
          { required: true, message: this.$t('api.611p62ddn7s0'), trigger: 'blur' },
          { min: 1, max: 100, message: this.$t('valid.length', [1, 100]), trigger: 'blur' }
        ],
        category: [
          { required: true, message: this.$t('api.611p62ddnao0'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('valid.length', [1, 50]), trigger: 'blur' }
        ],
        method: [
          { required: true, message: this.$t('api.611p62ddmzc0'), trigger: 'change' }
        ],
        remark: [
          { required: false, message: this.$t('api.611p62ddmlo0'), trigger: 'blur' },
          { min: 0, max: 100, message: this.$t('valid.length', [0, 100]), trigger: 'blur' }
        ]
      },
      popoverVisible: false,
      multipleSelection: []
    };
  },
  created() {
    this.getTableData();
  },
  methods: {
    search() {
      this.params.pageNum = 1;
      this.getTableData();
    },
    methodTagFilter(val) {
      switch (val) {
        case 'GET':
          return 'info';
        case 'POST':
          return 'success';
        case 'PUT':
          return 'info';
        case 'PATCH':
          return 'warning';
        case 'DELETE':
          return 'danger';
        default:
          return 'info';
      }
    },
    async getTableData() {
      this.loading = true;
      try {
        const { data } = await getApis(this.params);
        this.tableData = data.apis;
        this.total = data.total;
      } finally {
        this.loading = false;
      }
    },
    create() {
      this.dialogFormTitle = this.$t('api.611p62ddnd40');
      this.dialogType = 'create';
      this.dialogFormVisible = true;
    },
    update(row) {
      this.dialogFormData.ID = row.ID;
      this.dialogFormData.path = row.path;
      this.dialogFormData.category = row.category;
      this.dialogFormData.method = row.method;
      this.dialogFormData.remark = row.remark;

      this.dialogFormTitle = this.$t('api.611p62ddng00');
      this.dialogType = 'update';
      this.dialogFormVisible = true;
    },
    judgeResult(res) {
      if (res.code === 200 || res.code === 0) {
        ElMessage({
          showClose: true,
          message: this.$t('api.611p62ddnjk0'),
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
              await createApi(this.dialogFormData).then(res => {
                this.judgeResult(res);
              });
            } else {
              await updateApiById(this.dialogFormData).then(res => {
                this.judgeResult(res);
              });
            }
          } finally {
            this.submitLoading = false;
          }
          this.resetForm();
          this.getTableData();
        } else {
          ElMessage({
            showClose: true,
            message: this.$t('api.611p62ddnm40'),
            type: 'warn'
          });
          return false;
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
        ID: '',
        path: '',
        category: '',
        method: '',
        remark: ''
      };
    },
    batchDelete() {
      ElMessageBox.confirm(this.$t('tips.deleteWarning'), this.$t('api.611p62ddnos0'), {
        confirmButtonText: this.$t('api.611p62ddn4s0'),
        cancelButtonText: this.$t('api.611p62ddn240'),
        type: 'warning'
      }).then(async () => {
        this.loading = true;
        const apiIds = [];
        this.multipleSelection.forEach(x => {
          apiIds.push(x.ID);
        });
        try {
          await batchDeleteApiByIds({ apiIds: apiIds }).then(res => {
            this.judgeResult(res);
          });
        } finally {
          this.loading = false;
        }
        this.getTableData();
      }).catch(() => {
        ElMessage({
          showClose: true,
          type: 'info',
          message: this.$t('api.611p62ddnr00')
        });
      });
    },
    handleSelectionChange(val) {
      this.multipleSelection = val;
    },
    async singleDelete(Id) {
      this.loading = true;
      try {
        await batchDeleteApiByIds({ apiIds: [Id] }).then(res => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getTableData();
    },
    handleSizeChange(val) {
      this.params.pageSize = val;
      this.getTableData();
    },
    handleCurrentChange(val) {
      this.params.pageNum = val;
      this.getTableData();
    }
  }
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
