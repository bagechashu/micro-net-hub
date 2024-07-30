<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="mini" :inline="true" class="demo-form-inline">
        <el-form-item>
          <el-button :loading="loading" icon="el-icon-plus" type="warning" @click="create">{{ $t('menu.61131vuglrw0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :disabled="multipleSelection.length === 0" :loading="loading" icon="el-icon-delete" type="danger" @click="batchDelete">{{ $t('menu.61131vugmb40') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :tree-props="{children: 'children', hasChildren: 'hasChildren'}" row-key="ID" :data="tableData" border stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column show-overflow-tooltip prop="name" :label="$t('menu.61131vugmf40')" width="200" />
        <el-table-column show-overflow-tooltip prop="icon" :label="$t('menu.61131vugmhg0')" />
        <el-table-column show-overflow-tooltip prop="path" :label="$t('menu.61131vugmjo0')" />
        <el-table-column show-overflow-tooltip prop="component" :label="$t('menu.61131vugmm40')" />
        <el-table-column show-overflow-tooltip prop="redirect" :label="$t('menu.61131vugmoo0')" />
        <el-table-column show-overflow-tooltip prop="sort" :label="$t('menu.61131vugmqw0')" align="center" width="80" />
        <el-table-column show-overflow-tooltip prop="status" :label="$t('menu.61131vugmts0')" align="center" width="80">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.status === 1 ? 'success':'danger'">{{ scope.row.status === 1 ? '否':'是' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip prop="hidden" :label="$t('menu.61131vugmw40')" align="center" width="80">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.hidden === 1 ? 'danger':'success'">{{ scope.row.hidden === 1 ? '是':'否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip prop="noCache" :label="$t('menu.61131vugmy80')" align="center" width="80">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.noCache === 1 ? 'danger':'success'">{{ scope.row.noCache === 1 ? '否':'是' }}</el-tag>
          </template>
        </el-table-column>
        <!-- <el-table-column show-overflow-tooltip prop="activeMenu" label="高亮菜单" /> -->
        <el-table-column fixed="right" :label="$t('menu.61131vugn0s0')" align="center" width="120">
          <template slot-scope="scope">
            <el-tooltip fixed :content="$t('menu.61131vugn2w0')" effect="dark" placement="top">
              <el-button size="mini" icon="el-icon-edit" circle type="primary" @click="update(scope.row)" />
            </el-tooltip>
            <el-tooltip class="delete-popover" fixed :content="$t('menu.61131vugn5g0')" effect="dark" placement="top">
              <el-popconfirm :title="$t('menu.61131vugn7s0')" @confirm="singleDelete(scope.row.ID)">
                <el-button slot="reference" size="mini" icon="el-icon-delete" circle type="danger" />
              </el-popconfirm>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="dialogFormTitle" :visible.sync="dialogFormVisible" width="800px">
        <el-form ref="dialogForm" :inline="true" size="small" :model="dialogFormData" :rules="dialogFormRules" label-width="140px">
          <el-form-item :label="$t('menu.61131vugmf40')" prop="name">
            <el-input v-model.trim="dialogFormData.name" :placeholder="$t('menu.61131vugna00')" style="width: 600px" />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmqw0')" prop="sort" style="width: 600px">
            <el-input-number v-model.number="dialogFormData.sort" controls-position="right" :min="1" :max="999" />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmhg0')" prop="icon">
            <el-popover
              placement="bottom-start"
              width="450"
              trigger="click"
              @show="$refs['iconSelect'].reset()"
            >
              <IconSelect ref="iconSelect" @selected="selected" />
              <el-input slot="reference" v-model="dialogFormData.icon" style="width: 600px;" :placeholder="$t('menu.61131vugncg0')" readonly>
                <svg-icon v-if="dialogFormData.icon" slot="prefix" :icon-class="dialogFormData.icon" class="el-input__icon" style="height: 32px;width: 16px;" />
                <i v-else slot="prefix" class="el-icon-search el-input__icon" />
              </el-input>
            </el-popover>
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmjo0')" prop="path">
            <el-input v-model.trim="dialogFormData.path" :placeholder="$t('menu.61131vugneo0')" style="width: 600px" />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmm40')" prop="component">
            <el-input v-model.trim="dialogFormData.component" :placeholder="$t('menu.61131vugngs0')" style="width: 600px" />
          </el-form-item>
          <el-form-item :label="$t('menu.61131vugmoo0')" prop="redirect">
            <el-input v-model.trim="dialogFormData.redirect" :placeholder="$t('menu.61131vugnj40')" style="width: 600px" />
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
          <!-- <el-form-item label="高亮菜单" prop="activeMenu">
            <el-input v-model.trim="dialogFormData.activeMenu" :placeholder="$t('menu.61131vugnr80')" style="width: 440px" />
          </el-form-item> -->
          <el-form-item :label="$t('menu.61131vugnt40')" prop="parentId">
            <!-- <el-cascader
              v-model="dialogFormData.parentId"
              :show-all-levels="false"
              :options="treeselectData"
              :props="{ checkStrictly: true, label:'title', value:'ID', emitPath:false}"
              clearable
              filterable
            /> -->
            <treeselect
              v-model="dialogFormData.parentId"
              :options="treeselectData"
              :normalizer="normalizer"
              style="width:600px"
              @input="treeselectInput"
            />
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('menu.61131vugnuw0') }}</el-button>
          <el-button size="mini" :loading="submitLoading" type="primary" @click="submitForm()">{{ $t('menu.61131vugnwo0') }}</el-button>
        </div>
      </el-dialog>

    </el-card>
  </div>
</template>

<script>
import IconSelect from "@/components/IconSelect";
import Treeselect from "@riophae/vue-treeselect";
import "@riophae/vue-treeselect/dist/vue-treeselect.css";
import { getMenuTree, createMenu, updateMenuById, batchDeleteMenuByIds } from "@/api/system/menu";
import { Message } from "element-ui";

export default {
  name: "Menu",
  components: {
    IconSelect,
    Treeselect
  },
  data() {
    return {
      // 表格数据
      tableData: [],
      loading: false,

      // 上级目录数据
      treeselectData: [],
      treeselectValue: undefined,

      // dialog对话框
      submitLoading: false,
      dialogFormTitle: "",
      dialogType: "",
      dialogFormVisible: false,
      dialogFormData: {
        ID: "",
        name: "",
        icon: "",
        path: "",
        component: "Layout",
        redirect: "",
        sort: 999,
        status: this.$t("menu.61131vugnow0"),
        hidden: this.$t("menu.61131vugnow0"),
        noCache: this.$t("menu.61131vugnmc0"),
        alwaysShow: 2,
        breadcrumb: 1,
        // activeMenu: '',
        parentId: 0
      },
      dialogFormRules: {
        name: [
          { required: true, message: this.$t("menu.61131vugo7s0"), trigger: "blur" },
          { min: 1, max: 20, message: this.$t("valid.length", [1, 20]), trigger: "blur" },
          { validator: (rule, value, callback) => {
            if (!value || !/\s/.test(value)) {
              callback();
            } else {
              callback(new Error("因为i18n引用了该字段, 不能有空格"));
            }
          }, trigger: "blur" }
        ],
        path: [
          { required: true, message: this.$t("menu.61131vugoa40"), trigger: "blur" },
          { min: 1, max: 100, message: this.$t("valid.length", [1, 100]), trigger: "blur" }
        ],
        component: [
          { required: false, message: this.$t("menu.61131vugoc00"), trigger: "blur" },
          { min: 0, max: 100, message: this.$t("valid.length", [0, 100]), trigger: "blur" }
        ],
        redirect: [
          { required: false, message: this.$t("menu.61131vugoe40"), trigger: "blur" },
          { min: 0, max: 100, message: this.$t("valid.length", [0, 100]), trigger: "blur" }
        ],
        // activeMenu: [
        //   { required: false, message: '请输入高亮菜单', trigger: 'blur' },
        //   { min: 0, max: 100, message: '长度在 0 到 100 个字符', trigger: 'blur' }
        // ],
        parentId: [
          { required: true, message: this.$t("menu.61131vugofw0"), trigger: "change" }
        ]

      },

      // 删除按钮弹出框
      popoverVisible: false,
      // 表格多选
      multipleSelection: []
    };
  },
  created() {
    this.getTableData();
  },
  methods: {
    // 获取表格数据
    async getTableData() {
      this.loading = true;
      try {
        const { data } = await getMenuTree();

        this.tableData = data;
        this.treeselectData = [{ ID: 0, name: this.$t("menu.61131vugohw0"), children: data }];
      } finally {
        this.loading = false;
      }
    },

    // 新增
    create() {
      this.dialogFormTitle = this.$t("menu.61131vugojk0");
      this.dialogType = "create";
      this.dialogFormVisible = true;
    },

    // 修改
    update(row) {
      this.dialogFormData.ID = row.ID;
      this.dialogFormData.name = row.name;
      this.dialogFormData.icon = row.icon;
      this.dialogFormData.path = row.path;
      this.dialogFormData.component = row.component;
      this.dialogFormData.redirect = row.redirect;
      this.dialogFormData.sort = row.sort;
      this.dialogFormData.status = row.status === 1 ? this.$t("menu.61131vugnow0") : this.$t("menu.61131vugnmc0");
      this.dialogFormData.hidden = row.hidden === 1 ? this.$t("menu.61131vugnmc0") : this.$t("menu.61131vugnow0");
      this.dialogFormData.noCache = row.noCache === 1 ? this.$t("menu.61131vugnow0") : this.$t("menu.61131vugnmc0");
      // this.dialogFormData.activeMenu = row.activeMenu
      this.dialogFormData.parentId = row.parentId;

      this.dialogFormTitle = this.$t("menu.61131vugon80");
      this.dialogType = "update";
      this.dialogFormVisible = true;
    },

    // 判断结果
    judgeResult(res) {
      if (res.code === 0 || res.code === 200) {
        Message({
          showClose: true,
          message: this.$t("menu.61131vugop80"),
          type: "success"
        });
      }
    },

    // 提交表单
    submitForm() {
      this.$refs["dialogForm"].validate(async valid => {
        if (valid) {
          this.submitLoading = true;
          if (this.dialogFormData.ID === this.dialogFormData.parentId) {
            return Message({
              showClose: true,
              message: this.$t("menu.61131vugoqw0"),
              type: "error"
            });
          }

          // 处理表单项
          this.dialogFormData.component = this.dialogFormData.component || "Layout";
          this.dialogFormData.status = this.dialogFormData.status === this.$t("menu.61131vugnmc0") ? 2 : 1;
          this.dialogFormData.hidden = this.dialogFormData.hidden === this.$t("menu.61131vugnmc0") ? 1 : 2;
          this.dialogFormData.noCache = this.dialogFormData.noCache === this.$t("menu.61131vugnmc0") ? 2 : 1;

          // 创建副本逻辑
          const dialogFormDataCopy = typeof this.treeselectValue !== "undefined"
            ? { ...this.dialogFormData, parentId: this.treeselectValue }
            : this.dialogFormData;

          try {
            const res = this.dialogType === "create"
              ? await createMenu(dialogFormDataCopy)
              : await updateMenuById(dialogFormDataCopy);

            this.judgeResult(res);
          } finally {
            this.submitLoading = false;
          }
          this.resetForm();
          this.getTableData();
        } else {
          Message({
            showClose: true,
            message: this.$t("menu.61131vugosw0"),
            type: "error"
          });
          return false;
        }
      });
    },

    // 提交表单
    cancelForm() {
      this.resetForm();
    },

    resetForm() {
      this.dialogFormVisible = false;
      this.$refs["dialogForm"].resetFields();
      this.dialogFormData = {
        name: "",
        icon: "",
        path: "",
        component: "Layout",
        redirect: "",
        sort: 999,
        status: this.$t("menu.61131vugnow0"),
        hidden: this.$t("menu.61131vugnow0"),
        noCache: this.$t("menu.61131vugnmc0"),
        alwaysShow: 2,
        breadcrumb: 1,
        // activeMenu: '',
        parentId: 0
      };
    },

    // 批量删除
    batchDelete() {
      this.$confirm(this.$t("tips.deleteWarning"), this.$t("menu.61131vugov00"), {
        confirmButtonText: this.$t("menu.61131vugnwo0"),
        cancelButtonText: this.$t("menu.61131vugnuw0"),
        type: "warning"
      }).then(async res => {
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
        Message({
          type: "info",
          message: this.$t("menu.61131vugox00")
        });
      });
    },

    // 表格多选
    handleSelectionChange(val) {
      this.multipleSelection = val;
    },

    // 单个删除
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

    // 选中图标
    selected(name) {
      this.dialogFormData.icon = name;
    },

    // treeselect
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
};
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
