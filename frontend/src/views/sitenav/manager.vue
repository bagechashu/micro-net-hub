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
        <el-form-item :label="$t('sitenav.6112v35e5140')" prop="name">
          <el-input
            v-model.trim="navGroupForm.name"
            :placeholder="$t('sitenav.6112v35e5ik0')"
          />
        </el-form-item>
        <el-form-item :label="$t('sitenav.6112v35e5o40')" prop="title">
          <el-input
            v-model.trim="navGroupForm.title"
            :placeholder="$t('sitenav.6112v35e5r40')"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            size="small"
            :loading="loading"
            type="primary"
            @click="addGroup()"
          >{{ $t('sitenav.6112v35e5v40') }}</el-button>
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
                >{{ $t('sitenav.6112v35e5xw0') }}</el-button>
              </el-form-item>
              <el-form-item>
                <el-button
                  :disabled="multipleSelection.length === 0"
                  :loading="loading"
                  icon="el-icon-delete"
                  type="danger"
                  @click="batchDeleteSites"
                >{{ $t('sitenav.6112v35e60s0') }}</el-button>
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
                :label="$t('sitenav.6112v35e63c0')"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="icon"
                :label="$t('sitenav.6112v35e66c0')"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="desc"
                :label="$t('sitenav.6112v35e68w0')"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="link"
                :label="$t('sitenav.6112v35e6bw0')"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="doc"
                :label="$t('sitenav.6112v35e6eo0')"
              />
              <el-table-column
                fixed="right"
                :label="$t('sitenav.6112v35e6ho0')"
                align="center"
                width="120"
              >
                <template slot-scope="scope">
                  <el-tooltip :content="$t('sitenav.6112v35e6k80')" effect="dark" placement="top">
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
                    :content="$t('sitenav.6112v35e6mw0')"
                    effect="dark"
                    placement="top"
                  >
                    <el-popconfirm
                      :title="$t('sitenav.6112v35e6pk0')"
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
      <el-empty v-else :description="$t('common.nodata')" />

      <el-dialog :title="navSiteFormTitle" :visible.sync="navSiteFormVisible">
        <el-form
          ref="navSiteForm"
          size="small"
          :model="navSiteForm"
          :rules="navSiteFormRules"
          label-width="auto"
        >
          <el-form-item :label="$t('sitenav.6112v35e6s80')" prop="groupid">
            <el-select v-model="navSiteForm.groupid" :placeholder="$t('sitenav.6112v35e6v40')">
              <el-option
                v-for="item in groupOptions"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e63c0')" prop="name">
            <el-input v-model.trim="navSiteForm.name" :placeholder="$t('sitenav.6112v35e63c0')" />
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e66c0')" prop="icon">
            <el-select
              v-model.trim="navSiteForm.icon"
              filterable
              allow-create
              default-first-option
              :placeholder="$t('sitenav.6112v35e6y00')"
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
              :placeholder="$t('sitenav.6112v35e66c0')"
            /> -->
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e6bw0')" prop="link">
            <el-input v-model.trim="navSiteForm.link" :placeholder="$t('sitenav.6112v35e6bw0')" />
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e6eo0')" prop="doc">
            <el-input
              v-model.trim="navSiteForm.doc"
              :placeholder="$t('sitenav.6112v35e70o0')"
            />
          </el-form-item>
          <el-form-item :label="$t('sitenav.6112v35e68w0')" prop="desc">
            <el-input
              v-model.trim="navSiteForm.desc"
              type="textarea"
              :placeholder="$t('sitenav.6112v35e68w0')"
              show-word-limit
              maxlength="100"
            />
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="navSiteFormCancel()">{{ $t('sitenav.6112v35e74g0') }}</el-button>
          <el-button
            size="mini"
            :loading="loading"
            type="primary"
            @click="navSiteFormSubmit()"
          >{{ $t('sitenav.6112v35e76o0') }}</el-button>
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
            message: this.$t("sitenav.6112v35e7980"),
            trigger: "blur"
          },
          {
            min: 4,
            max: 20,
            message: this.$t("valid.length", [4, 20]),
            trigger: "blur"
          }
        ],
        title: [
          {
            required: true,
            message: this.$t("sitenav.6112v35e7bc0"),
            trigger: "blur"
          },
          {
            min: 4,
            max: 20,
            message: this.$t("valid.length", [4, 20]),
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
          { required: true, message: this.$t("sitenav.6112v35e7dc0"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        icon: [
          { required: true, message: this.$t("sitenav.6112v35e7fw0"), trigger: "blur" },
          {
            min: 1,
            max: 100,
            message: this.$t("valid.length", [1, 100]),
            trigger: "blur"
          }
        ],
        link: [
          { required: true, message: this.$t("sitenav.6112v35e7hw0"), trigger: "change" },
          {
            min: 0,
            max: 100,
            message: this.$t("valid.length", [1, 100]),
            trigger: "blur"
          }
        ],
        doc: [
          { required: false, message: this.$t("sitenav.6112v35e7k00"), trigger: "blur" },
          {
            min: 0,
            max: 100,
            message: this.$t("valid.length", [1, 100]),
            trigger: "blur"
          }
        ],
        desc: [
          { required: true, message: this.$t("sitenav.6112v35e7m40"), trigger: "blur" },
          {
            min: 0,
            max: 200,
            message: this.$t("valid.length", [1, 200]),
            trigger: "blur"
          }
        ],
        groupid: [{ required: true, message: this.$t("sitenav.6112v35e7o40"), trigger: "blur" }]
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
            message: this.$t("sitenav.6112v35e7qg0"),
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
        this.$t("sitenav.deleteSiteTips"),
        this.$t("sitenav.6112v35e7sg0"),
        {
          confirmButtonText: this.$t("sitenav.6112v35e76o0"),
          cancelButtonText: this.$t("sitenav.6112v35e74g0"),
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
            message: this.$t("sitenav.6112v35e7uw0")
          });
        });
    },
    // 新增
    addSite() {
      this.getGroupOptions();
      const navGroupId = this.getGroupIDFromTabname(this.navGroupActiveTab);
      this.navSiteForm.groupid = navGroupId;

      this.navSiteFormTitle = this.$t("sitenav.6112v35e7ww0");
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

      this.navSiteFormTitle = this.$t("sitenav.6112v35e7z00");
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
      this.$confirm(this.$t("tips.deleteWarning"), this.$t("sitenav.6112v35e7sg0"), {
        confirmButtonText: this.$t("sitenav.6112v35e76o0"),
        cancelButtonText: this.$t("sitenav.6112v35e74g0"),
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
            message: this.$t("sitenav.6112v35e7uw0")
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
            message: this.$t("sitenav.6112v35e7qg0"),
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
          message: this.$t("sitenav.6112v35e8140"),
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
