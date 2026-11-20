<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form :inline="true" class="demo-form-inline">
        <el-form-item>
          <el-button :loading="loading" type="primary" @click="create" :icon="Plus">{{ $t('menu.61131vuglrw0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :disabled="multipleSelection.length === 0" :loading="loading" type="danger" @click="batchDelete" :icon="Delete">{{ $t('menu.61131vugmb40') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :tree-props="{children: 'children', hasChildren: 'hasChildren'}" row-key="ID" :data="tableData" stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="48" align="center" />
        <el-table-column show-overflow-tooltip prop="name" :label="$t('menu.61131vugmf40')" width="200" />
        <el-table-column show-overflow-tooltip prop="icon" :label="$t('menu.61131vugmhg0')" width="120">
          <template v-slot="scope">
            <span>{{ scope.row.icon }}</span>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip prop="path" :label="$t('menu.61131vugmjo0')" min-width="160" />
        <el-table-column show-overflow-tooltip prop="component" :label="$t('menu.61131vugmm40')" min-width="160" />
        <el-table-column show-overflow-tooltip prop="redirect" :label="$t('menu.61131vugmoo0')" min-width="140" />
        <el-table-column show-overflow-tooltip prop="sort" :label="$t('menu.61131vugmqw0')" width="80" align="center" />
        <el-table-column show-overflow-tooltip prop="status" :label="$t('menu.61131vugmts0')" width="90" align="center">
          <template v-slot="scope">
            <el-tag size="small" :type="scope.row.status === 1 ? 'success':'danger'">{{ scope.row.status === 1 ? '否':'是' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip prop="hidden" :label="$t('menu.61131vugmw40')" width="90" align="center">
          <template v-slot="scope">
            <el-tag size="small" :type="scope.row.hidden === 1 ? 'danger':'success'">{{ scope.row.hidden === 1 ? '是':'否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip prop="noCache" :label="$t('menu.61131vugmy80')" width="90" align="center">
          <template v-slot="scope">
            <el-tag size="small" :type="scope.row.noCache === 1 ? 'danger':'success'">{{ scope.row.noCache === 1 ? '否':'是' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column fixed="right" :label="$t('menu.61131vugn0s0')" align="left" width="160">
          <template v-slot="scope">
            <div class="row-actions">
              <el-button link type="primary" :icon="Edit" @click="update(scope.row)">{{ $t('common.edit') }}</el-button>
              <el-popconfirm :title="$t('menu.61131vugn7s0')" @confirm="singleDelete(scope.row.ID)">
                <template #reference>
                  <el-button link type="danger" :icon="Delete">{{ $t('common.delete') }}</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="dialogFormTitle" v-model="dialogFormVisible" width="720px">
        <el-form ref="dialogForm" class="form-grid" :model="dialogFormData" :rules="dialogFormRules" label-position="top">
          <el-form-item :label="$t('menu.61131vugmf40')" prop="name">
            <el-input v-model.trim="dialogFormData.name" :placeholder="$t('menu.61131vugna00')" />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugnt40')" prop="parentId">
            <TreeSelect
              v-model="dialogFormData.parentId"
              :options="treeselectData"
              :normalizer="normalizer"
              @input="treeselectInput"
            />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmjo0')" prop="path">
            <el-input v-model.trim="dialogFormData.path" :placeholder="$t('menu.61131vugneo0')" />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmm40')" prop="component">
            <el-input v-model.trim="dialogFormData.component" :placeholder="$t('menu.61131vugngs0')" />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmoo0')" prop="redirect">
            <el-input v-model.trim="dialogFormData.redirect" :placeholder="$t('menu.61131vugnj40')" />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmhg0')" prop="icon">
            <el-popover
              placement="bottom-start"
              width="450"
              trigger="click"
              @show="$refs['iconSelect'].reset()"
            >
              <IconSelect ref="iconSelect" @selected="selected" />
              <template #reference>
                <el-input v-model="dialogFormData.icon" :placeholder="$t('menu.61131vugncg0')" readonly>
                  <template v-if="dialogFormData.icon" #prefix>
                    <svg-icon :icon-class="dialogFormData.icon" class="el-input__icon" style="height: 32px;width: 16px;" />
                  </template>
                </el-input>
              </template>
            </el-popover>
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmqw0')" prop="sort">
            <el-input-number v-model.number="dialogFormData.sort" controls-position="right" :min="1" :max="999" />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmts0')" prop="status">
            <el-radio-group v-model="dialogFormData.status">
              <el-radio-button :label="$t('menu.61131vugnmc0')" />
              <el-radio-button :label="$t('menu.61131vugnow0')" />
            </el-radio-group>
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmw40')" prop="hidden">
            <el-radio-group v-model="dialogFormData.hidden">
              <el-radio-button :label="$t('menu.61131vugnmc0')" />
              <el-radio-button :label="$t('menu.61131vugnow0')" />
            </el-radio-group>
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmy80')" prop="noCache">
            <el-radio-group v-model="dialogFormData.noCache">
              <el-radio-button :label="$t('menu.61131vugnmc0')" />
              <el-radio-button :label="$t('menu.61131vugnow0')" />
            </el-radio-group>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="cancelForm()">{{ $t('menu.61131vugnuw0') }}</el-button>
          <el-button :loading="submitLoading" type="primary" @click="submitForm()">{{ $t('menu.61131vugnwo0') }}</el-button>
        </template>
      </el-dialog>

    </el-card>
  </div>
</template>

<script>
import { defineComponent } from 'vue';
import IconSelect from '@/components/IconSelect';
import TreeSelect from '@zanmato/vue3-treeselect';
import '@zanmato/vue3-treeselect/dist/vue3-treeselect.min.css';
import { getMenuTree, createMenu, updateMenuById, batchDeleteMenuByIds } from '@/api/system/menu';
import { Edit, Delete, Plus } from '@/icons/tabler';

export default defineComponent({
  name: 'SystemMenu',
  components: {
    IconSelect,
    TreeSelect
  },
  setup() {
    // 供模板里 el-button 的 :icon 使用
    return { Edit, Delete, Plus };
  },
  data() {
    return {
      tableData: [],
      loading: false,
      treeselectData: [],
      treeselectValue: undefined,
      submitLoading: false,
      dialogFormTitle: '',
      dialogType: '',
      dialogFormVisible: false,
      dialogFormData: {
        ID: '',
        name: '',
        icon: '',
        path: '',
        component: 'Layout',
        redirect: '',
        sort: 999,
        status: this.$t('menu.61131vugnow0'),
        hidden: this.$t('menu.61131vugnow0'),
        noCache: this.$t('menu.61131vugnmc0'),
        alwaysShow: 2,
        breadcrumb: 1,
        parentId: 0
      },
      dialogFormRules: {
        name: [
          { required: true, message: this.$t('menu.61131vugo7s0'), trigger: 'blur' },
          { min: 1, max: 20, message: this.$t('valid.length', [1, 20]), trigger: 'blur' },
          { validator: (rule, value, callback) => {
            if (!value || !/\s/.test(value)) {
              callback();
            } else {
              callback(new Error('因为i18n引用了该字段, 不能有空格'));
            }
          }, trigger: 'blur' }
        ],
        path: [
          { required: true, message: this.$t('menu.61131vugoa40'), trigger: 'blur' },
          { min: 1, max: 100, message: this.$t('valid.length', [1, 100]), trigger: 'blur' }
        ],
        component: [
          { required: false, message: this.$t('menu.61131vugoc00'), trigger: 'blur' },
          { min: 0, max: 100, message: this.$t('valid.length', [0, 100]), trigger: 'blur' }
        ],
        redirect: [
          { required: false, message: this.$t('menu.61131vugoe40'), trigger: 'blur' },
          { min: 0, max: 100, message: this.$t('valid.length', [0, 100]), trigger: 'blur' }
        ],
        parentId: [
          { required: true, message: this.$t('menu.61131vugofw0'), trigger: 'change' }
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
    async getTableData() {
      this.loading = true;
      try {
        const { data } = await getMenuTree();
        this.tableData = data;
        this.treeselectData = [{ ID: 0, name: this.$t('menu.61131vugohw0'), children: data }];
      } finally {
        this.loading = false;
      }
    },
    create() {
      this.dialogFormTitle = this.$t('menu.61131vugojk0');
      this.dialogType = 'create';
      this.dialogFormVisible = true;
    },
    update(row) {
      this.dialogFormData.ID = row.ID;
      this.dialogFormData.name = row.name;
      this.dialogFormData.icon = row.icon;
      this.dialogFormData.path = row.path;
      this.dialogFormData.component = row.component;
      this.dialogFormData.redirect = row.redirect;
      this.dialogFormData.sort = row.sort;
      this.dialogFormData.status = row.status === 1 ? this.$t('menu.61131vugnow0') : this.$t('menu.61131vugnmc0');
      this.dialogFormData.hidden = row.hidden === 1 ? this.$t('menu.61131vugnmc0') : this.$t('menu.61131vugnow0');
      this.dialogFormData.noCache = row.noCache === 1 ? this.$t('menu.61131vugnow0') : this.$t('menu.61131vugnmc0');
      this.dialogFormData.parentId = row.parentId;

      this.dialogFormTitle = this.$t('menu.61131vugon80');
      this.dialogType = 'update';
      this.dialogFormVisible = true;
    },
    judgeResult(res) {
      if (res.code === 0 || res.code === 200) {
        ElMessage({
          showClose: true,
          message: this.$t('menu.61131vugop80'),
          type: 'success'
        });
      }
    },
    submitForm() {
      // @ts-expect-error refs
      this.$refs['dialogForm'].validate(async (valid) => {
        if (valid) {
          this.submitLoading = true;
          if (this.dialogFormData.ID === this.dialogFormData.parentId) {
            ElMessage({
              showClose: true,
              message: this.$t('menu.61131vugoqw0'),
              type: 'error'
            });
            this.submitLoading = false;
            return;
          }

          this.dialogFormData.component = this.dialogFormData.component || 'Layout';
          this.dialogFormData.status = this.dialogFormData.status === this.$t('menu.61131vugnmc0') ? 2 : 1;
          this.dialogFormData.hidden = this.dialogFormData.hidden === this.$t('menu.61131vugnmc0') ? 1 : 2;
          this.dialogFormData.noCache = this.dialogFormData.noCache === this.$t('menu.61131vugnmc0') ? 2 : 1;

          const dialogFormDataCopy = typeof this.treeselectValue !== 'undefined'
            ? { ...this.dialogFormData, parentId: this.treeselectValue }
            : this.dialogFormData;

          try {
            const res = this.dialogType === 'create'
              ? await createMenu(dialogFormDataCopy)
              : await updateMenuById(dialogFormDataCopy);

            this.judgeResult(res);
          } finally {
            this.submitLoading = false;
          }
          this.resetForm();
          this.getTableData();
        } else {
          ElMessage({
            showClose: true,
            message: this.$t('menu.61131vugosw0'),
            type: 'error'
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
        name: '',
        icon: '',
        path: '',
        component: 'Layout',
        redirect: '',
        sort: 999,
        status: this.$t('menu.61131vugnow0'),
        hidden: this.$t('menu.61131vugnow0'),
        noCache: this.$t('menu.61131vugnmc0'),
        alwaysShow: 2,
        breadcrumb: 1,
        parentId: 0
      };
    },
    batchDelete() {
      ElMessageBox.confirm(this.$t('tips.deleteWarning'), this.$t('menu.61131vugov00'), {
        confirmButtonText: this.$t('menu.61131vugnwo0'),
        cancelButtonText: this.$t('menu.61131vugnuw0'),
        type: 'warning'
      }).then(async () => {
        this.loading = true;
        const menuIds = [];
        this.multipleSelection.forEach(x => {
          menuIds.push(x.ID);
        });
        try {
          await batchDeleteMenuByIds({ menuIds: menuIds }).then(res => {
            this.judgeResult(res);
          });
        } finally {
          this.loading = false;
        }
        this.getTableData();
      }).catch(() => {
        ElMessage({
          type: 'info',
          message: this.$t('menu.61131vugox00')
        });
      });
    },
    handleSelectionChange(val) {
      this.multipleSelection = val;
    },
    async singleDelete(Id) {
      this.loading = true;
      try {
        await batchDeleteMenuByIds({ menuIds: [Id] }).then(res => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getTableData();
    },
    selected(name) {
      this.dialogFormData.icon = name;
    },
    normalizer(node) {
      return {
        id: node.ID,
        label: node.name,
        children: node.children
      };
    },
    treeselectInput(value) {
      this.treeselectValue = value;
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
