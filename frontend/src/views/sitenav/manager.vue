<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        ref="navGroupForm"
        :inline="true"
        size="small"
        :model="navGroupForm"
        :rules="navGroupFormRules"
      >
        <el-form-item label="标识名:" prop="name">
          <el-input
            v-model.trim="navGroupForm.name"
            placeholder="导航组的唯一标识"
          />
        </el-form-item>
        <el-form-item label="展示名:" prop="title">
          <el-input
            v-model.trim="navGroupForm.title"
            placeholder="导航组的实际展示名"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            size="small"
            :loading="loading"
            type="primary"
            @click="addGroup()"
          >添加导航组</el-button>
        </el-form-item>
      </el-form>

      <el-tabs
        v-if="navData.length > 0"
        v-model="navGroupActiveTab"
        type="border-card"
        closable
        @tab-remove="deleteGroup"
      >
        <el-tab-pane
          v-for="item in navData"
          :key="item.name"
          :label="item.title"
          :name="item.name"
        >
          <template>
            <el-form size="mini" :inline="true" class="demo-form-inline">
              <el-form-item>
                <el-button
                  :loading="loading"
                  icon="el-icon-plus"
                  type="warning"
                  @click="addSite"
                >新增</el-button>
              </el-form-item>
              <el-form-item>
                <el-button
                  :disabled="multipleSelection.length === 0"
                  :loading="loading"
                  icon="el-icon-delete"
                  type="danger"
                  @click="batchDeleteSites"
                >批量删除</el-button>
              </el-form-item>
            </el-form>

            <el-table
              v-loading="loading"
              :data="item.sites"
              border
              stripe
              size="mini"
              style="width: 100%"
              @selection-change="handleSelectionChange"
            >
              <el-table-column type="selection" width="55" align="center" />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="name"
                label="站名"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="icon"
                label="图标"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="desc"
                label="描述"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="link"
                label="链接"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="doc"
                label="文档"
              />
              <el-table-column
                fixed="right"
                label="操作"
                align="center"
                width="120"
              >
                <template slot-scope="scope">
                  <el-tooltip content="编辑" effect="dark" placement="top">
                    <el-button
                      size="mini"
                      icon="el-icon-edit"
                      circle
                      type="primary"
                      @click="updateSite(scope.row)"
                    />
                  </el-tooltip>
                  <el-tooltip
                    class="delete-popover"
                    content="删除"
                    effect="dark"
                    placement="top"
                  >
                    <el-popconfirm
                      title="确定删除吗？"
                      @confirm="deleteSite(scope.row.ID)"
                    >
                      <el-button
                        slot="reference"
                        size="mini"
                        icon="el-icon-delete"
                        circle
                        type="danger"
                      />
                    </el-popconfirm>
                  </el-tooltip>
                </template>
              </el-table-column>
            </el-table>
          </template>
        </el-tab-pane>
      </el-tabs>
      <el-empty v-else description="暂无数据" />

      <el-dialog :title="navSiteFormTitle" :visible.sync="navSiteFormVisible">
        <el-form
          ref="navSiteForm"
          size="small"
          :model="navSiteForm"
          :rules="navSiteFormRules"
          label-width="auto"
        >
          <el-form-item label="组ID" prop="groupid">
            <el-select v-model="navSiteForm.groupid" placeholder="请选择">
              <el-option
                v-for="item in groupOptions"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="站名" prop="name">
            <el-input v-model.trim="navSiteForm.name" placeholder="站名" />
          </el-form-item>
          <el-form-item label="图标" prop="icon">
            <el-select
              v-model.trim="navSiteForm.icon"
              filterable
              allow-create
              default-first-option
              placeholder="可以输入图片URL"
            >
              <el-option
                v-for="item in iconOptions"
                :key="item.url"
                :label="item.url"
                :value="item.url"
              />
            </el-select>
            <!-- <el-input
              v-model.trim="navSiteForm.icon"
              placeholder="图标"
            /> -->
          </el-form-item>
          <el-form-item label="链接" prop="link">
            <el-input v-model.trim="navSiteForm.link" placeholder="链接" />
          </el-form-item>
          <el-form-item label="文档" prop="doc">
            <el-input
              v-model.trim="navSiteForm.doc"
              placeholder="文档, 不填不展示"
            />
          </el-form-item>
          <el-form-item label="描述" prop="desc">
            <el-input
              v-model.trim="navSiteForm.desc"
              type="textarea"
              placeholder="描述"
              show-word-limit
              maxlength="100"
            />
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="navSiteFormCancel()">取 消</el-button>
          <el-button
            size="mini"
            :loading="loading"
            type="primary"
            @click="navSiteFormSubmit()"
          >确 定</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import {
  getNav,
  getIcons,
  createNavGroup,
  batchDeleteNavGroupByIds,
  createNavSite,
  updateNavSite,
  batchDeleteNavSiteByIds
} from "@/api/sitenav/sitenav";
import { Message } from "element-ui";

export default {
  name: "SiteManager",
  data() {
    return {
      loading: false,

      navData: [], // 导航页数据
      navGroupForm: {
        name: "",
        title: ""
      },
      navGroupFormRules: {
        name: [
          {
            required: true,
            message: "请输入导航组的唯一标识",
            trigger: "blur"
          },
          {
            min: 4,
            max: 20,
            message: "长度在 4 到 20 个字符",
            trigger: "blur"
          }
        ],
        title: [
          {
            required: true,
            message: "请输入导航组的实际展示名",
            trigger: "blur"
          },
          {
            min: 4,
            max: 20,
            message: "长度在 4 到 20 个字符",
            trigger: "blur"
          }
        ]
      },
      navGroupActiveTab: "",

      navSiteFormTitle: "",
      navSiteFormType: "",
      navSiteFormVisible: false,
      navSiteForm: {
        name: "",
        icon: "",
        link: "",
        doc: "",
        desc: "",
        groupid: 1
      },
      navSiteFormRules: {
        name: [
          { required: true, message: "请输入站名", trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: "长度在 1 到 50 个字符",
            trigger: "blur"
          }
        ],
        icon: [
          { required: true, message: "请输入图标链接", trigger: "blur" },
          {
            min: 1,
            max: 100,
            message: "长度在 1 到 100 个字符",
            trigger: "blur"
          }
        ],
        link: [
          { required: true, message: "请输入链接地址", trigger: "change" },
          {
            min: 0,
            max: 100,
            message: "长度在 0 到 100 个字符",
            trigger: "blur"
          }
        ],
        doc: [
          { required: false, message: "请输入文档地址", trigger: "blur" },
          {
            min: 0,
            max: 100,
            message: "长度在 0 到 100 个字符",
            trigger: "blur"
          }
        ],
        desc: [
          { required: true, message: "输入描述", trigger: "blur" },
          {
            min: 0,
            max: 200,
            message: "长度在 0 到 200 个字符",
            trigger: "blur"
          }
        ],
        groupid: [{ required: true, message: "必须选择分组", trigger: "blur" }]
      },
      iconOptions: [],
      groupOptions: [],
      // 表格多选
      multipleSelection: []
    };
  },
  created() {
    this.getData();
    this.getIconOptions();
  },
  methods: {
    // 获取表格数据
    async getData() {
      this.loading = true;
      try {
        const { data } = await getNav();
        this.navData = data;

        // console.log(`navGroupActiveTab type: ${typeof(this.navGroupActiveTab)}, value: ${this.navGroupActiveTab}`);
        // default navGroupActiveTab type: string, value: 0
        if ((this.navGroupActiveTab === "0" || this.navGroupActiveTab === "") && this.navData.length > 0) {
          this.navGroupActiveTab = this.navData[0].name;
        }
      } finally {
        this.loading = false;
      }
    },
    getIconOptions() {
      getIcons().then((res) => {
        this.iconOptions = res.logos;
      });
    },
    getGroupOptions() {
      this.groupOptions = this.navData.map((item) => {
        return {
          id: item.ID,
          name: item.title
        };
      });
    },
    addGroup() {
      this.$refs["navGroupForm"].validate(async(valid) => {
        if (valid) {
          this.loading = true;
          try {
            await createNavGroup(this.navGroupForm).then((res) => {
              this.judgeResult(res);
            });
          } finally {
            this.loading = false;
          }
          this.getData();
        } else {
          Message({
            showClose: true,
            message: "表单校验失败",
            type: "warn"
          });
          return false;
        }
      });
    },
    // 根据 tabname 查找 navgroup的 ID
    getGroupIDFromTabname(tabname) {
      for (let i = 0; i < this.navData.length; i++) {
        if (this.navData[i].name === tabname) {
          // console.log(this.navData[i].ID);
          return this.navData[i].ID;
        }
      }
    },
    async deleteGroup(tabname) {
      const navGroupId = this.getGroupIDFromTabname(tabname);
      this.$confirm(
        "此操作将永久删除该导航组及其包含的记录, 是否继续?",
        "提示",
        {
          confirmButtonText: "确定",
          cancelButtonText: "取消",
          type: "warning"
        }
      )
        .then(async() => {
          try {
            await batchDeleteNavGroupByIds({
              ids: [navGroupId]
            }).then((res) => {
              this.judgeResult(res);
            });
            this.navGroupActiveTab = "0";
            this.getData();
          } finally {
            this.loading = false;
          }
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: "已取消删除"
          });
        });
    },
    // 新增
    addSite() {
      this.getGroupOptions();
      const navGroupId = this.getGroupIDFromTabname(this.navGroupActiveTab);
      this.navSiteForm.groupid = navGroupId;

      this.navSiteFormTitle = "新增站点";
      this.navSiteFormType = "add";
      this.navSiteFormVisible = true;
    },

    // 修改
    updateSite(row) {
      this.getGroupOptions();
      this.navSiteForm.name = row.name;
      this.navSiteForm.icon = row.icon;
      this.navSiteForm.desc = row.desc;
      this.navSiteForm.link = row.link;
      this.navSiteForm.doc = row.doc;
      this.navSiteForm.groupid = row.groupid;

      this.navSiteFormTitle = "更新站点";
      this.navSiteFormType = "update";
      this.navSiteFormVisible = true;
    },

    // 单个删除
    async deleteSite(id) {
      this.loading = true;
      try {
        await batchDeleteNavSiteByIds({ ids: [id] }).then((res) => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getData();
    },
    // 批量删除
    batchDeleteSites() {
      this.$confirm("此操作将永久删除, 是否继续?", "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning"
      })
        .then(async() => {
          this.loading = true;
          const ids = [];
          this.multipleSelection.forEach((x) => {
            ids.push(x.ID);
          });
          try {
            await batchDeleteNavSiteByIds({ ids: ids }).then((res) => {
              this.judgeResult(res);
            });
          } finally {
            this.loading = false;
          }
          this.getData();
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: "已取消删除"
          });
        });
    },
    // 提交表单
    // https://stackoverflow.com/questions/73772552/typeerror-this-refs-is-not-a-function
    navSiteFormSubmit() {
      this.$refs["navSiteForm"].validate(async(valid) => {
        if (valid) {
          this.loading = true;
          try {
            if (this.navSiteFormType === "add") {
              await createNavSite(this.navSiteForm).then((res) => {
                this.judgeResult(res);
              });
            } else if (this.navSiteFormType === "update") {
              await updateNavSite(this.navSiteForm).then((res) => {
                this.judgeResult(res);
              });
            }
          } finally {
            this.loading = false;
          }
          this.navSiteFormReset();
          this.getData();
        } else {
          Message({
            showClose: true,
            message: "表单校验失败",
            type: "warn"
          });
          return false;
        }
      });
    },

    // 提交表单
    navSiteFormCancel() {
      this.navSiteFormReset();
    },

    navSiteFormReset() {
      this.navSiteFormVisible = false;
      this.$refs["navSiteForm"].resetFields();
      this.navSiteForm = {
        name: "",
        icon: "",
        desc: "",
        link: "",
        doc: ""
      };
    },

    // 判断结果
    judgeResult(res) {
      if (res.code === 200) {
        Message({
          showClose: true,
          message: "操作成功",
          type: "success"
        });
      }
    },
    // 表格多选
    handleSelectionChange(val) {
      this.multipleSelection = val;
    }
  }
};
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
