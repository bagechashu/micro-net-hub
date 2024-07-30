<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        size="mini"
        :inline="true"
        :model="params"
        class="demo-form-inline"
      >
        <el-form-item :label="$t('group.610yx3auj8w0')">
          <el-input
            v-model.trim="params.groupName"
            style="width: 100px"
            clearable
            :placeholder="$t('group.610yx3auj8w0')"
            @keyup.enter.native="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item :label="$t('group.610yx3auk3k0')">
          <el-input
            v-model.trim="params.remark"
            style="width: 100px"
            clearable
            :placeholder="$t('group.610yx3auk3k0')"
            @keyup.enter.native="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item :label="$t('group.610yx3auk780')">
          <el-select
            v-model.trim="params.syncState"
            style="width: 110px"
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
            icon="el-icon-search"
            type="primary"
            @click="search"
          >{{ $t('group.610yx3auke40') }}</el-button>
        </el-form-item>
        <!-- <el-form-item>
          <el-button :loading="loading" icon="el-icon-plus" type="warning" @click="resetData">{{$t('group.610yx3aukgg0')}}</el-button>
        </el-form-item> -->
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-plus"
            type="warning"
            @click="create"
          >{{ $t('group.610yx3aukig0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            icon="el-icon-delete"
            type="danger"
            @click="batchDelete"
          >{{ $t('group.610yx3aukks0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            icon="el-icon-upload2"
            type="success"
            @click="batchSync"
          >{{ $t('group.610yx3aukmw0') }}</el-button>
        </el-form-item>
        <br>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-download"
            type="warning"
            @click="syncOpenLdapDepts"
          >{{ $t('group.610yx3aukos0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-download"
            type="warning"
            @click="syncDingTalkDepts"
          >{{ $t('group.610yx3aukr40') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-download"
            type="warning"
            @click="syncFeiShuDepts"
          >{{ $t('group.610yx3aukvw0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-download"
            type="warning"
            @click="syncWeComDepts"
          >{{ $t('group.610yx3aukyc0') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="loading"
        :default-expand-all="true"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        row-key="ID"
        :data="infoTableData"
        border
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="groupName"
          :label="$t('group.610yx3auj8w0')"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="groupType"
          :label="$t('group.610yx3aul140')"
          width="180"
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
          width="500"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="remark"
          :label="$t('group.610yx3auk3k0')"
          width="320"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="CreatedAt"
          :label="$t('group.610yx3aul380')"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="UpdatedAt"
          :label="$t('group.610yx3aul5c0')"
        />
        <el-table-column fixed="right" :label="$t('group.610yx3aul7g0')" align="center" width="220">
          <template #default="scope">
            <el-tooltip
              v-if="
                scope.row.groupType != 'ou' && scope.row.groupName != 'root'
              "
              :content="$t('group.610yx3aul9g0')"
              effect="dark"
              placement="top"
            >
              <el-button
                size="mini"
                icon="el-icon-setting"
                circle
                type="info"
                @click="handleGetUserOfGroup(scope.row)"
              />
            </el-tooltip>
            <el-tooltip :content="$t('group.610yx3aulbs0')" effect="dark" placement="top">
              <el-button
                size="mini"
                icon="el-icon-edit"
                circle
                type="primary"
                @click="update(scope.row)"
              />
            </el-tooltip>
            <el-tooltip
              class="delete-popover"
              :content="$t('group.610yx3auleo0')"
              effect="dark"
              placement="top"
            >
              <el-popconfirm
                :title="$t('group.610yx3aulgo0')"
                @confirm="singleDelete(scope.row.ID)"
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
            <el-tooltip
              v-if="scope.row.syncState === 2"
              class="delete-popover"
              :content="$t('group.610yx3aulis0')"
              effect="dark"
              placement="top"
            >
              <el-popconfirm
                :title="$t('group.610yx3aulks0')"
                @confirm="singleSync(scope.row.ID)"
              >
                <el-button
                  slot="reference"
                  size="mini"
                  icon="el-icon-upload2"
                  circle
                  type="success"
                />
              </el-popconfirm>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
      <!-- 新增 -->
      <el-dialog :title="dialogFormTitle" :visible.sync="dialogGroupAddFormVisible">
        <el-form
          ref="dialogForm"
          size="small"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-width="120px"
        >
          <el-form-item :label="$t('group.610yx3auj8w0')" prop="groupName">
            <el-input
              v-model.trim="dialogFormData.groupName"
              :placeholder="$t('group.610yx3auj8w0')"
            />
          </el-form-item>
          <el-form-item :label="$t('group.610yx3aulns0')" prop="groupType">
            <el-select
              v-model.trim="dialogFormData.groupType"
              :placeholder="$t('group.610yx3aulpo0')"
              style="width: 100%"
            >
              <el-option label="ou [organizationalUnit]" value="ou" />
              <el-option label="cn [groupOfUniqueNames]" value="cn" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('group.610yx3aulrk0')" prop="parentId">
            <treeselect
              v-model="dialogFormData.parentId"
              :options="treeselectData"
              :normalizer="normalizer"
              :placeholder="$t('group.610yx3aultc0')"
              @input="treeselectInput"
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
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('group.610yx3aulv00') }}</el-button>
          <el-button
            size="mini"
            :loading="submitLoading"
            type="primary"
            @click="submitForm()"
          >{{ $t('group.610yx3aulws0') }}</el-button>
        </div>
      </el-dialog>
      <!-- 编辑 -->
      <el-dialog :title="dialogFormTitle" :visible.sync="dialogGroupEditFormVisible">
        <el-form
          ref="dialogForm"
          size="small"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-width="120px"
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
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('group.610yx3aulv00') }}</el-button>
          <el-button
            size="mini"
            :loading="submitLoading"
            type="primary"
            @click="submitForm()"
          >{{ $t('group.610yx3aulws0') }}</el-button>
        </div>
      </el-dialog>
      <!-- 组用户管理 -->
      <el-dialog :title="dialogUsersMgrTitle" :visible.sync="dialogUsersMgrVisible">
        <el-row>
          <el-col :span="9">
            <el-form size="mini" :inline="true" :model="userOfGroupParams">
              <el-form-item>
                <el-input
                  v-model.trim="userOfGroupParams.nickname"
                  clearable
                  :placeholder="$t('group.610yx3aulyk0')"
                  @keyup.enter.native="getUsersOfGroupData()"
                  @clear="getUsersOfGroupData()"
                />
              </el-form-item>
              <el-button
                size="mini"
                :loading="submitLoading"
                icon="el-icon-search"
                type="primary"
                @click="getUsersOfGroupData()"
              >{{ $t('group.610yx3aum0c0') }}</el-button>
            </el-form></el-col>
          <el-col :span="9">
            <el-form size="mini" :inline="true" :model="groupAddUserParams">
              <el-form-item>
                <el-select
                  v-model="groupAddUserList"
                  multiple
                  filterable
                  remote
                  reserve-keyword
                  :placeholder="$t('group.610yx3aulyk0')"
                  :remote-method="getUsersNoInGroupData"
                  :loading="submitLoading"
                >
                  <el-option
                    v-for="item in groupAddUserListOption"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
              <el-button-group>
                <el-button
                  size="mini"
                  :disabled="groupAddUserList.length === 0"
                  :loading="submitLoading"
                  icon="el-icon-delete"
                  type="success"
                  @click="handleGroupAddUser()"
                >{{ $t('group.610yx3aum1w0') }}</el-button>
              </el-button-group>
            </el-form></el-col>
          <el-col :span="4">
            <el-form size="mini" class="flex-inline-form" :inline="true">
              <el-button
                size="mini"
                :disabled="userOfGroupListMultipleSelection.length === 0"
                :loading="submitLoading"
                icon="el-icon-delete"
                type="danger"
                @click="handleGroupDelUser()"
              >{{ $t('group.610yx3aum3o0') }}</el-button>
            </el-form></el-col>
        </el-row>

        <el-table
          :data="userOfGroupList"
          stripe
          height="500"
          style="width: 100%"
          max-height="500"
          :default-sort="{ prop: 'userName', order: 'ascending' }"
          @selection-change="handleUserOfGroupListSelectionChange"
        >
          <el-table-column type="selection" />
          <el-table-column type="expand">
            <template slot-scope="props">
              <el-form label-position="left" class="table-expand">
                <el-form-item :label="$t('group.610yx3aum5k0')">
                  <span>{{ props.row.mobile }}</span>
                </el-form-item>
                <el-form-item :label="$t('group.610yx3aum7g0')">
                  <span>{{ props.row.mail }}</span>
                </el-form-item>
              </el-form>
            </template>
          </el-table-column>
          <el-table-column prop="userName" :label="$t('group.610yx3aum940')" sortable />
          <el-table-column prop="nickName" :label="$t('group.610yx3aumb40')" />
          <el-table-column prop="introduction" :label="$t('group.610yx3aumcw0')" />
        </el-table>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import Treeselect from "@riophae/vue-treeselect";
import "@riophae/vue-treeselect/dist/vue-treeselect.css";
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
} from "@/api/personnel/group";

import { validateName } from "@/utils/validate";
import { Message } from "element-ui";

export default {
  name: "Group",
  components: {
    Treeselect
  },
  filters: {
    methodTagFilter(val) {
      if (val === "GET") {
        return "";
      } else if (val === "POST") {
        return "success";
      } else {
        return "info";
      }
    }
  },
  data() {
    return {
      // 查询参数
      params: {
        groupName: undefined,
        remark: undefined,
        syncState: undefined,
        pageNum: 1,
        pageSize: 1000 // 平常百姓人家应该不会有这么多数据吧,后台限制最大单次获取1000条
      },
      // 表格数据
      groupTree: [],
      infoTableData: [],
      total: 0,
      loading: false,
      // 上级目录数据
      treeselectData: [],
      treeselectValue: 0,
      dialogGroupAddFormVisible: false, // 新增
      // dialog对话框
      submitLoading: false,
      dialogFormTitle: "",
      dialogType: "",
      dialogGroupEditFormVisible: false,
      dialogFormData: {
        ID: "",
        groupName: "",
        parentId: 0,
        syncState: 1,
        groupType: "",
        remark: ""
      },
      dialogFormRules: {
        groupName: [
          { required: true, message: this.$t("group.610yx3aumeo0"), trigger: "blur" },
          {
            min: 2,
            max: 30,
            message: this.$t("valid.length", [2, 30]),
            trigger: "blur"
          },
          { required: true, validator: validateName, trigger: "blur" }
        ],
        groupType: [
          { required: true, message: this.$t("group.610yx3aumgg0"), trigger: "blur" },
          { min: 1, max: 50, message: this.$t("group.selectGroupTypeTips"), trigger: "blur" }
        ],
        parentId: [
          { required: true, message: this.$t("group.610yx3aumi40"), trigger: "blur" },
          {
            validator: (rule, value, callBack) => {
              if (value >= 0) {
                callBack();
              } else {
                callBack(this.$t("group.610yx3aumjw0"));
              }
            }
          }
        ],
        remark: [
          { required: false, message: this.$t("group.610yx3aumlk0"), trigger: "blur" },
          {
            min: 0,
            max: 100,
            message: this.$t("valid.length", [0, 100]),
            trigger: "blur"
          }
        ]
      },

      // 删除按钮弹出框
      popoverVisible: false,
      // 表格多选
      multipleSelection: [],

      dialogUsersMgrVisible: false,
      dialogUsersMgrTitle: "",
      userOfGroupParams: {
        groupId: "",
        nickname: ""
      },
      userOfGroupList: [],
      userOfGroupListMultipleSelection: [],

      groupAddUserParams: {
        groupId: "",
        nickname: ""
      },
      userNoInGroupList: [],
      groupAddUserListOption: [],
      groupAddUserList: [],

      renderFunc(h, option) {
        return (
          <span>
            {option.key} - {option.label}
          </span>
        );
      },
      userArrInfo: [], // 初始人员列表数据
      data: [], // 转化后人员列表数据
      value3: [], // 右侧默认人员列表数据
      userId: [], // 送到后台 -> 勾选的数据code数组
      ui: {
        submitLoading: false
      },
      statusTrans: ""
    };
  },
  created() {
    this.getGroupTreeData();
  },
  methods: {
    // // 查询
    search() {
      // 初始化表格数据
      this.infoTableData = JSON.parse(JSON.stringify(this.groupTree));
      this.infoTableData = this.deal(
        this.infoTableData,
        (node) =>
          node.groupName.includes(this.params.groupName) ||
          node.remark.includes(this.params.remark) ||
          node.syncState.toString().includes(this.params.syncState)
      );
    },
    resetData() {
      this.infoTableData = JSON.parse(JSON.stringify(this.groupTree));
    },
    // 页面数据过滤
    deal(nodes, predicate) {
      // 如果已经没有节点了，结束递归
      if (!(nodes && nodes.length)) {
        return [];
      }
      const newChildren = [];
      for (const node of nodes) {
        if (predicate(node)) {
          // 如果节点符合条件，直接加入新的节点集
          newChildren.push(node);
          node.children = this.deal(node.children, predicate);
        } else {
          // 如果当前节点不符合条件，递归过滤子节点，
          // 把符合条件的子节点提升上来，并入新节点集
          newChildren.push(...this.deal(node.children, predicate));
        }
      }
      return newChildren;
    },
    // 获取表格数据
    async getGroupTreeData() {
      this.loading = true;
      try {
        const { data } = await getGroupTree(this.params);
        this.groupTree = data;
        this.infoTableData = JSON.parse(JSON.stringify(data));
        this.treeselectData = [
          { ID: 0, groupName: this.$t("group.610yx3aumo00"), children: data }
        ];
      } finally {
        this.loading = false;
      }
    },
    // 用户管理
    handleGetUserOfGroup(row) {
      this.dialogUsersMgrVisible = true;
      this.dialogUsersMgrTitle = this.$t("group.group", [row.groupName]);
      this.userOfGroupParams.groupId = row.ID;
      this.userOfGroupParams.nickname = "";
      this.groupAddUserParams.groupId = row.ID;
      this.groupAddUserParams.nickname = "";
      this.getUsersOfGroupData();
    },
    // 获取组内用户数据
    async getUsersOfGroupData() {
      this.submitLoading = true;
      try {
        const { data } = await userInGroup(this.userOfGroupParams);
        this.userOfGroupList = data.userList;
        // console.log(this.userOfGroupList);
      } finally {
        this.submitLoading = false;
      }
    },
    handleUserOfGroupListSelectionChange(val) {
      this.userOfGroupListMultipleSelection = val;
    },
    // 删除组用户
    async handleGroupDelUser() {
      this.$confirm(this.$t("group.610yx3aumps0"), this.$t("group.610yx3aumrs0"), {
        confirmButtonText: this.$t("group.610yx3aulws0"),
        cancelButtonText: this.$t("group.610yx3aulv00"),
        type: "warning"
      })
        .then(async(res) => {
          this.loading = true;
          const user = [];
          this.userOfGroupListMultipleSelection.forEach((x) => {
            user.push(x.userId);
          });
          try {
            await groupDelUser({
              groupId: Number(this.userOfGroupParams.groupId),
              userIds: user
            }).then((res) => {
              if (res.code === 0) {
                Message({
                  showClose: true,
                  message: this.$t("group.610yx3aumtg0"),
                  type: "success"
                });
              }
            });
          } finally {
            this.loading = false;
          }
          this.getUsersOfGroupData();
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: this.$t("group.610yx3aumvc0")
          });
        });
    },
    // 获取非组内用户数据
    async getUsersNoInGroupData(query) {
      this.submitLoading = true;
      this.groupAddUserParams.nickname = query;
      try {
        const { data } = await userNoInGroup(this.groupAddUserParams);
        this.userNoInGroupList = data.userList;
        // console.log(this.userNoInGroupList);
        this.groupAddUserListOption = this.userNoInGroupList.map((item) => {
          return {
            label: item.userName,
            value: item.userId
          };
        });
        // console.log(this.groupAddUserListOption);
      } finally {
        this.submitLoading = false;
      }
    },
    handleGroupAddUser(query) {
      this.$confirm(this.$t("group.610yx3aumx00"), this.$t("group.610yx3aumrs0"), {
        confirmButtonText: this.$t("group.610yx3aulws0"),
        cancelButtonText: this.$t("group.610yx3aulv00"),
        type: "warning"
      })
        .then(async(res) => {
          // console.log(this.groupAddUserList)
          this.loading = true;
          try {
            await groupAddUser({
              groupId: Number(this.groupAddUserParams.groupId),
              userIds: this.groupAddUserList
            }).then((res) => {
              if (res.code === 0) {
                Message({
                  showClose: true,
                  message: this.$t("group.610yx3aumtg0"),
                  type: "success"
                });
              }
            });
          } finally {
            this.loading = false;
          }
          this.userOfGroupParams.nickname = "";
          this.groupAddUserList = [];
          this.groupAddUserParams.nickname = "";
          this.getUsersOfGroupData();
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: this.$t("group.610yx3aumvc0")
          });
        });
    },
    // 新增
    create() {
      this.dialogFormTitle = this.$t("group.610yx3aumyw0");
      this.dialogType = "create";
      this.dialogGroupAddFormVisible = true; // 新增的展示
    },
    // 修改
    update(row) {
      this.dialogFormData.ID = row.ID;
      this.dialogFormData.groupName = row.groupName;
      this.dialogFormData.remark = row.remark;
      this.dialogFormTitle = this.$t("group.610yx3aun0k0");
      this.dialogType = "update";
      this.dialogGroupEditFormVisible = true;
    },

    // 判断结果
    judgeResult(res) {
      if (res.code === 200) {
        Message({
          showClose: true,
          message: this.$t("group.610yx3aumtg0"),
          type: "success"
        });
      }
    },

    // 提交表单
    submitForm() {
      this.$refs["dialogForm"].validate(async(valid) => {
        if (valid) {
          this.submitLoading = true;
          try {
            if (this.dialogType === "create") {
              await groupAdd(this.dialogFormData).then((res) => {
                this.judgeResult(res);
              });
            } else {
              await groupUpdate(this.dialogFormData).then((res) => {
                this.judgeResult(res);
              });
            }
          } finally {
            this.submitLoading = false;
          }
          this.resetForm();
          this.getGroupTreeData();
        } else {
          Message({
            showClose: true,
            message: this.$t("group.610yx3aun2c0"),
            type: "warn"
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
      this.dialogGroupEditFormVisible = false;
      this.dialogGroupAddFormVisible = false;
      this.$refs["dialogForm"].resetFields();
      this.dialogFormData = {
        ID: "",
        groupName: "",
        parentId: 0,
        syncState: 1,
        groupType: "",
        remark: ""
      };
    },

    // 批量删除
    batchDelete() {
      this.$confirm(this.$t("group.batchDeleteTips"), this.$t("group.610yx3aumrs0"), {
        confirmButtonText: this.$t("group.610yx3aulws0"),
        cancelButtonText: this.$t("group.610yx3aulv00"),
        type: "warning"
      })
        .then(async(res) => {
          this.loading = true;
          const groupIds = [];
          this.multipleSelection.forEach((x) => {
            groupIds.push(x.ID);
          });
          try {
            await groupDel({ groupIds: groupIds }).then((res) => {
              this.judgeResult(res);
            });
          } finally {
            this.loading = false;
          }
          this.getGroupTreeData();
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: this.$t("group.610yx3aumvc0")
          });
        });
    },
    // 批量同步
    batchSync() {
      this.$confirm(this.$t("group.betchSyncLdapTips"), this.$t("group.610yx3aumrs0"), {
        confirmButtonText: this.$t("group.610yx3aulws0"),
        cancelButtonText: this.$t("group.610yx3aulv00"),
        type: "warning"
      })
        .then(async(res) => {
          this.loading = true;
          const groupIds = [];
          this.multipleSelection.forEach((x) => {
            groupIds.push(x.ID);
          });
          try {
            await syncSqlGroups({ groupIds: groupIds }).then((res) => {
              this.judgeResult(res);
            });
          } finally {
            this.loading = false;
          }
          this.getGroupTreeData();
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: this.$t("group.610yx3aun440")
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
        await groupDel({ groupIds: [Id] }).then((res) => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getGroupTreeData();
    },
    // 单个同步
    async singleSync(Id) {
      this.loading = true;
      try {
        await syncSqlGroups({ groupIds: [Id] }).then((res) => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getGroupTreeData();
    },

    // 分页
    handleSizeChange(val) {
      this.params.pageSize = val;
      this.getGroupTreeData();
    },
    handleCurrentChange(val) {
      this.params.pageNum = val;
      this.getGroupTreeData();
    },
    // treeselect
    normalizer(node) {
      return {
        id: node.ID,
        label: node.groupName,
        children: node.children
      };
    },
    treeselectInput(value) {
      this.treeselectValue = value;
    },
    syncDingTalkDepts() {
      this.loading = true;
      syncDingTalkDeptsApi().then((res) => {
        this.judgeResult(res);
        this.loading = false;
        this.getGroupTreeData();
      }).finally(() => {
        this.loading = false;
      });
    },
    syncWeComDepts() {
      this.loading = true;
      syncWeComDeptsApi().then((res) => {
        this.judgeResult(res);
        this.loading = false;
        this.getGroupTreeData();
      }).finally(() => {
        this.loading = false;
      });
    },
    syncFeiShuDepts() {
      this.loading = true;
      syncFeiShuDeptsApi().then((res) => {
        this.judgeResult(res);
        this.loading = false;
        this.getGroupTreeData();
      }).finally(() => {
        this.loading = false;
      });
    },
    syncOpenLdapDepts() {
      this.loading = true;
      syncOpenLdapDeptsApi().then((res) => {
        this.judgeResult(res);
        this.loading = false;
        this.getGroupTreeData();
      }).finally(() => {
        this.loading = false;
      });
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
.transfer-footer {
  margin-left: 20px;
  padding: 6px 5px;
}

.table-expand label {
  display: inline-block;
  width: 90px;
  color: #99a9bf;
}

.table-expand .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
  width: 50%;
}
</style>
