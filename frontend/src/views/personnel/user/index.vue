<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        size="mini"
        :inline="true"
        :model="params"
        class="demo-form-inline"
      >
        <el-form-item :label="$t('user.610wkzjikzs0')">
          <el-input
            v-model.trim="params.username"
            style="width: 100px"
            clearable
            :placeholder="$t('user.610wkzjikzs0')"
            @keyup.enter.native="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item :label="$t('user.610wkzjill40')">
          <el-input
            v-model.trim="params.nickname"
            style="width: 100px"
            clearable
            :placeholder="$t('user.610wkzjill40')"
            @keyup.enter.native="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item :label="$t('common.status')">
          <el-select
            v-model.trim="params.status"
            style="width: 100px"
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
            v-model.trim="params.syncState"
            style="width: 100px"
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
          <el-button
            :loading="loading"
            icon="el-icon-search"
            type="primary"
            @click="search"
          >{{ $t('common.query') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-plus"
            type="warning"
            @click="create"
          >{{ $t('common.add') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            icon="el-icon-delete"
            type="danger"
            @click="batchDelete"
          >{{ $t('user.610wkzjim8k0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            icon="el-icon-upload2"
            type="success"
            @click="batchSync"
          >{{ $t('user.610wkzjimb40') }}</el-button>
        </el-form-item>
        <br>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-download"
            type="warning"
            @click="syncOpenLdapUsers"
          >{{ $t('user.610wkzjimd80') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-download"
            type="warning"
            @click="syncDingTalkUsers"
          >{{ $t('user.610wkzjimfk0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-download"
            type="warning"
            @click="syncFeiShuUsers"
          >{{ $t('user.610wkzjimho0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-download"
            type="warning"
            @click="syncWeComUsers"
          >{{ $t('user.610wkzjimk00') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column :label="$t('user.610wkzjimm80')" width="65" type="expand">
          <template slot-scope="props">
            <el-form label-position="left" class="table-expand">
              <el-form-item label="userDN">
                <span>{{ props.row.userDn }}</span>
              </el-form-item>
              <el-form-item :label="$t('user.610wkzjimog0')">
                <span>{{ props.row.mobile }}</span>
              </el-form-item>
              <el-form-item :label="$t('user.610wkzjimqk0')">
                <span>{{ props.row.mail }}</span>
              </el-form-item>
              <el-form-item :label="$t('user.610wkzjimsw0')">
                <span>{{ props.row.jobNumber }}</span>
              </el-form-item>
              <el-form-item :label="$t('user.610wkzjimwo0')">
                <span>{{ props.row.creator }}</span>
              </el-form-item>
              <el-form-item :label="$t('user.610wkzjimys0')">
                <span>{{ props.row.position }}</span>
              </el-form-item>
              <el-form-item :label="$t('user.610wkzjin0w0')">
                <span>{{ props.row.introduction }}</span>
              </el-form-item>
            </el-form>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable :label="$t('user.610wkzjikzs0')" width="120">
          <template slot-scope="scope">
            <div slot="reference" class="name-wrapper">
              <el-tag size="medium">{{ scope.row.username }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.status')" width="80" align="center">
          <template slot-scope="scope">
            <el-switch
              v-model="scope.row.status"
              size="small"
              :active-value="1"
              :inactive-value="2"
              @change="userStateChanged(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="nickname"
          :label="$t('user.610wkzjill40')"
          width="120"
        />
        <el-table-column show-overflow-tooltip prop="givenName" :label="$t('user.610wkzjin4g0')" width="120" />
        <el-table-column
          show-overflow-tooltip
          sortable
          :label="$t('user.610wkzjin680')"
          width="150"
        >
          <template slot-scope="scope">
            <span v-for="(role, index) in scope.row.roles" :key="role.ID">
              <el-tag size="mini" effect="plain">{{ role.name }}</el-tag>
              <span v-if="index !== scope.row.roles.length - 1">&nbsp;|&nbsp;</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip :label="$t('user.610wkzjin800')" width="420">
          <template slot-scope="scope">
            <div v-if="scope.row.groups && scope.row.groups.length">
              <span v-for="(group, index) in scope.row.groups" :key="group.ID">
                {{ group.groupDn }}
                <span v-if="index !== scope.row.groups.length - 1"><br> </span>
              </span>
            </div>
            <div v-else>
              ---
            </div>
          </template>
        </el-table-column>
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="CreatedAt"
          :label="$t('common.createTime')"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="UpdatedAt"
          :label="$t('common.updateTime')"
        />
        <el-table-column fixed="right" :label="$t('common.management')" align="center" width="150">
          <template slot-scope="scope">
            <el-tooltip :content="$t('common.edit')" effect="dark" placement="top">
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
              :content="$t('common.delete')"
              effect="dark"
              placement="top"
            >
              <el-popconfirm
                :title="$t('user.610wkzjinjg0')"
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
              v-if="scope.row.syncState === 2 && scope.row.status === 1"
              class="delete-popover"
              :content="$t('user.610wkzjinlc0')"
              effect="dark"
              placement="top"
            >
              <el-popconfirm
                :title="$t('user.610wkzjinn40')"
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

      <el-dialog
        :title="dialogFormTitle"
        :visible.sync="dialogFormVisible"
        width="50%"
      >
        <el-form
          ref="dialogForm"
          size="small"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-width="150px"
        >
          <el-row>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjikzs0')" prop="username">
                <el-input
                  ref="password"
                  v-model.trim="dialogFormData.username"
                  :disabled="disabled"
                  :placeholder="$t('user.610wkzjinp00')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjill40')" prop="nickname">
                <el-input
                  v-model.trim="dialogFormData.nickname"
                  :placeholder="$t('user.610wkzjill40')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjin4g0')" prop="givenName">
                <el-input
                  v-model.trim="dialogFormData.givenName"
                  :placeholder="$t('user.610wkzjin4g0')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjinso0')" prop="avatar">
                <el-input
                  v-model.trim="dialogFormData.avatar"
                  :placeholder="$t('user.imageTips')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjimqk0')" prop="mail">
                <el-input
                  v-model.trim="dialogFormData.mail"
                  :placeholder="$t('user.610wkzjimqk0')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item
                v-if="dialogType === 'create'"
                :label="$t('user.610wkzjinuk0')"
                prop="password"
              >
                <el-input
                  v-model.trim="dialogFormData.password"
                  autocomplete="off"
                  :type="passwordType"
                  :placeholder="$t('user.610wkzjinwc0')"
                />
                <span class="show-pwd" @click="showPwd">
                  <svg-icon
                    :icon-class="
                      passwordType === 'password' ? 'eye' : 'eye-open'
                    "
                  />
                </span>
              </el-form-item>
              <el-form-item v-else :label="$t('user.610wkzjinyc0')" prop="password">
                <el-input
                  v-model.trim="dialogFormData.password"
                  autocomplete="off"
                  :type="passwordType"
                  :placeholder="$t('user.610wkzjio000')"
                />
                <span class="show-pwd" @click="showPwd">
                  <svg-icon
                    :icon-class="
                      passwordType === 'password' ? 'eye' : 'eye-open'
                    "
                  />
                </span>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjin680')" prop="roleIds">
                <el-select
                  v-model.trim="dialogFormData.roleIds"
                  multiple
                  :placeholder="$t('user.610wkzjio200')"
                  style="width: 100%"
                >
                  <el-option
                    v-for="item in roles"
                    :key="item.ID"
                    :label="item.name"
                    :value="item.ID"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('common.status')" prop="status">
                <el-select
                  v-model.trim="dialogFormData.status"
                  :placeholder="$t('user.610wkzjio3s0')"
                  style="width: 100%"
                >
                  <el-option :label="$t('common.enabled')" :value="1" />
                  <el-option :label="$t('common.disabled')" :value="2" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjio6c0')" prop="mobile">
                <el-input
                  v-model.trim="dialogFormData.mobile"
                  :placeholder="$t('user.610wkzjio6c0')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjimsw0')" prop="jobNumber">
                <el-input
                  v-model.trim="dialogFormData.jobNumber"
                  :placeholder="$t('user.610wkzjimsw0')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.610wkzjimys0')" prop="position">
                <el-input
                  v-model.trim="dialogFormData.position"
                  :placeholder="$t('user.610wkzjio8g0')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item :label="$t('user.610wkzjioa80')" prop="groupIds">
                <treeselect
                  v-model="dialogFormData.groupIds"
                  :options="groupsOptions"
                  :placeholder="$t('user.610wkzjioc00')"
                  :normalizer="normalizer"
                  value-consists-of="ALL"
                  :multiple="true"
                  :flat="true"
                  :no-children-text="$t('user.610wkzjiods0')"
                  :no-results-text="$t('user.610wkzjiofk0')"
                  @input="treeselectInput"
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
        <div slot="footer" class="dialog-footer">
          <el-switch
            v-model="notice"
            :active-text="$t('user.610wkzjiook0')"
            style="margin-right: 10px"
          />
          <el-button size="mini" @click="cancelForm()">{{ $t('common.cancel') }}</el-button>
          <el-button
            size="mini"
            :loading="submitLoading"
            type="primary"
            @click="submitForm()"
          >{{ $t('common.confirm') }}</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import { getGroupTree } from "@/api/personnel/group";
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
} from "@/api/personnel/user";
import { getRoles } from "@/api/system/role";
import { validatePasswordCanEnpty } from "@/utils/validate";
import Treeselect from "@riophae/vue-treeselect";
import "@riophae/vue-treeselect/dist/vue-treeselect.css";
import { Message } from "element-ui";
import JSEncrypt from "jsencrypt";

export default {
  name: "User",
  components: {
    Treeselect
  },
  data() {
    var checkPhone = (rule, value, callback) => {
      if (value) {
        const reg = /^(\+|00)??(\d{1,3})??((1|0)\d{8,10})??$/;
        if (reg.test(value)) {
          callback();
        } else {
          return callback(new Error(this.$t("user.610wkzjioxc0")));
        }
      } else {
        return callback();
        // return callback(new Error('手机号不能为空'))
      }
    };
    return {
      disabled: {
        // username 默认不可编辑，若需要至为可编辑，请（在新增和编辑处）去掉这个值的控制，且配合后端的ldap-user-name-modify配置使用
        type: Boolean,
        default: false
      },
      // 查询参数
      params: {
        username: "",
        nickname: "",
        status: "",
        syncState: "",
        pageNum: 1,
        pageSize: 10
      },
      // 表格数据
      tableData: [],
      total: 0,
      loading: false,
      isUpdate: false,
      // 部门信息数据
      treeselectValue: 0,
      // 角色
      roles: [],
      // 部门信息
      groupsOptions: [],

      passwordType: "password",

      publicKey: process.env.VUE_APP_PUBLIC_KEY,

      notice: true,
      // dialog对话框
      submitLoading: false,
      dialogFormTitle: "",
      dialogType: "",
      dialogFormVisible: false,
      dialogFormData: {
        ID: "",
        mail: "",
        givenName: "",
        username: "",
        password: "",
        nickname: "",
        status: 1,
        mobile: "",
        avatar: "",
        introduction: "",
        roleIds: [],
        jobNumber: "",
        position: "",
        postalAddress: "",
        groupIds: undefined,
        notice: true
      },

      dialogFormRules: {
        username: [
          {
            required: true,
            message: this.$t("user.610wkzjioz80"),
            trigger: "blur"
          },
          {
            min: 2,
            max: 20,
            message: this.$t("valid.length", [2, 20]),
            trigger: "blur"
          }
        ],
        password: [
          { required: false, validator: validatePasswordCanEnpty, trigger: "blur" }
        ],
        mail: [
          {
            required: true,
            message: this.$t("user.610wkzjip140"),
            trigger: "blur"
          }
        ],
        jobNumber: [
          {
            required: false,
            message: this.$t("user.610wkzjip380"),
            trigger: "blur"
          },
          {
            min: 0,
            max: 20,
            message: this.$t("valid.length", [0, 20]),
            trigger: "blur"
          }
        ],
        nickname: [
          {
            required: true,
            message: this.$t("user.610wkzjip540"),
            trigger: "blur"
          },
          {
            min: 2,
            max: 20,
            message: this.$t("valid.length", [2, 20]),
            trigger: "blur"
          }
        ],
        mobile: [
          {
            required: false,
            validator: checkPhone,
            trigger: "blur"
          }
        ],
        status: [{ required: true, message: this.$t("user.610wkzjio3s0"), trigger: "change" }],
        groupIds: [
          { required: false, message: this.$t("user.610wkzjioc00"), trigger: "blur" }
          // {
          //   validator: (rule, value, callBack) => {
          //     if (value < 1) {
          //       callBack("请选择有效的部门");
          //     } else {
          //       callBack();
          //     }
          //   }
          // }
        ],
        introduction: [
          { required: false, message: this.$t("user.610wkzjioj40"), trigger: "blur" },
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
      changeUserStatusFormData: {
        id: "",
        status: ""
      }
    };
  },
  created() {
    this.getTableData();
    this.getRoles();
  },
  methods: {
    // 查询
    search() {
      this.params.pageNum = 1;
      this.getTableData();
    },

    // 获取表格数据
    async getTableData() {
      this.loading = true;
      try {
        const { data } = await getUsers(this.params);
        data.users.forEach((item) => {
          const dataIntArr = [];
          item.groups.forEach((g) => {
            dataIntArr.push(+g.ID);
          });
          item.groupIds = dataIntArr;
        });
        this.tableData = data.users;
        this.total = data.total;
      } finally {
        this.loading = false;
      }
    },
    // 获取所有的分组信息，用于弹框选取上级分组
    async getAllGroups() {
      this.loading = true;
      try {
        const checkParams = {
          pageNum: 1,
          pageSize: 1000 // 平常百姓人家应该不会有这么多数据吧
        };
        const { data } = await getGroupTree(checkParams);
        this.groupsOptions = [
          {
            ID: 0,
            groupName: this.$t("user.610wkzjip6w0"),
            groupType: "T",
            children: data
          }
        ];
      } finally {
        this.loading = false;
      }
    },
    // 获取角色数据
    async getRoles() {
      const res = await getRoles(null);

      this.roles = res.data.roles;
    },

    // 新增
    create() {
      this.dialogFormTitle = this.$t("user.610wkzjip8w0");
      this.dialogType = "create";
      this.disabled = false;
      this.getAllGroups();
      this.dialogFormVisible = true;

      this.dialogFormData.roleIds = [2];
      this.notice = true;
    },

    // 修改
    update(row) {
      this.dialogFormTitle = this.$t("user.610wkzjipao0");
      this.dialogType = "update";
      this.disabled = true;
      this.passwordType = "password";
      this.dialogFormVisible = true;

      this.getAllGroups();
      this.dialogFormData.ID = row.ID;
      this.dialogFormData.mail = row.mail;
      this.dialogFormData.givenName = row.givenName;
      this.dialogFormData.username = row.username;
      this.dialogFormData.password = "";
      this.dialogFormData.nickname = row.nickname;
      this.dialogFormData.status = row.status;
      this.dialogFormData.mobile = row.mobile;
      this.dialogFormData.avatar = row.avatar;
      this.dialogFormData.introduction = row.introduction;
      // 遍历角色数组，获取角色ID
      this.dialogFormData.roleIds = row.roles.map((item) => item.ID);

      this.dialogFormData.jobNumber = row.jobNumber;
      this.dialogFormData.position = row.position;
      this.dialogFormData.postalAddress = row.postalAddress;
      this.dialogFormData.groupIds = row.groupIds;
      this.notice = false;
    },

    // 判断结果
    judgeResult(res) {
      if (res.code === 200) {
        const message = res.data ? res.data : this.$t("user.610wkzjipcg0");
        Message({
          showClose: true,
          message: message,
          type: "success"
        });
      }
    },

    // 提交表单
    submitForm() {
      if (this.dialogFormData.nickname === "") {
        Message({
          showClose: true,
          message: this.$t("user.610wkzjipec0"),
          type: "error"
        });
        return false;
      }
      if (this.dialogFormData.username === "") {
        Message({
          showClose: true,
          message: this.$t("user.610wkzjipg00"),
          type: "error"
        });
        return false;
      }
      if (this.dialogFormData.mail === "") {
        Message({
          showClose: true,
          message: this.$t("user.610wkzjipi40"),
          type: "error"
        });
        return false;
      }
      // if (this.dialogFormData.jobNumber === '') {
      //   Message({
      //     showClose: true,
      //     message: '请填写工号',
      //     type: 'error'
      //   })
      //   return false
      // }
      // if (this.dialogFormData.mobile === '') {
      //   Message({
      //     showClose: true,
      //     message: '请填写手机号',
      //     type: 'error'
      //   })
      //   return false
      // }
      if (this.dialogFormData.status === "") {
        Message({
          showClose: true,
          message: this.$t("user.610wkzjipk40"),
          type: "error"
        });
        return false;
      }
      this.dialogFormData.notice = this.notice;
      if (this.dialogFormData.roleIds === "") {
        Message({
          showClose: true,
          message: this.$t("user.610wkzjipn40"),
          type: "error"
        });
        return false;
      }
      this.$refs["dialogForm"].validate(async(valid) => {
        if (valid) {
          this.submitLoading = true;
          this.dialogFormDataCopy = { ...this.dialogFormData };
          if (this.dialogFormData.password !== "") {
            // 密码RSA加密处理
            const encryptor = new JSEncrypt();
            // 设置公钥
            encryptor.setPublicKey(this.publicKey);
            // 加密密码
            const encPassword = encryptor.encrypt(this.dialogFormData.password);
            this.dialogFormDataCopy.password = encPassword;
          }
          try {
            if (this.dialogType === "create") {
              await createUser(this.dialogFormDataCopy).then((res) => {
                this.judgeResult(res);
              });
            } else {
              await updateUserById(this.dialogFormDataCopy).then((res) => {
                this.judgeResult(res);
              });
            }
          } finally {
            this.submitLoading = false;
          }
          this.resetForm();
          this.getTableData();
        } else {
          Message({
            showClose: true,
            message: this.$t("user.610wkzjipp00"),
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
      this.dialogFormVisible = false;
      this.$refs["dialogForm"].resetFields();
      this.dialogFormData = {
        ID: "",
        mail: "",
        givenName: "",
        username: "",
        password: "",
        nickname: "",
        status: 1,
        mobile: "",
        avatar: "",
        introduction: "",
        roleIds: [],
        jobNumber: "",
        postalAddress: "",
        position: "",
        groupIds: undefined,
        notice: true
      };
    },

    // 批量删除
    batchDelete() {
      this.$confirm(this.$t("user.batchDeleteTips"), this.$t("common.prompt"), {
        confirmButtonText: this.$t("common.confirm"),
        cancelButtonText: this.$t("common.cancel"),
        type: "warning"
      })
        .then(async(res) => {
          this.loading = true;
          const userIds = [];
          this.multipleSelection.forEach((x) => {
            userIds.push(x.ID);
          });
          try {
            await batchDeleteUserByIds({ userIds: userIds }).then((res) => {
              this.judgeResult(res);
            });
          } finally {
            this.loading = false;
          }
          this.getTableData();
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: this.$t("user.610wkzjipy40")
          });
        });
    },
    // 批量同步
    batchSync() {
      this.$confirm(this.$t("user.betchSyncLdapTips"), this.$t("common.prompt"), {
        confirmButtonText: this.$t("common.confirm"),
        cancelButtonText: this.$t("common.cancel"),
        type: "warning"
      })
        .then(async(res) => {
          this.loading = true;
          const userIds = [];
          this.multipleSelection.forEach((x) => {
            userIds.push(x.ID);
          });
          try {
            await syncSqlUsers({ userIds: userIds }).then((res) => {
              this.judgeResult(res);
            });
          } finally {
            this.loading = false;
          }
          this.getTableData();
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: this.$t("user.610wkzjiq080")
          });
        });
    },

    // 监听 switch 开关 状态改变
    async userStateChanged(userInfo) {
      this.changeUserStatusFormData.id = userInfo.ID;
      this.changeUserStatusFormData.status = userInfo.status;
      const { code } = await changeUserStatus(this.changeUserStatusFormData);
      if (code !== 200) {
        // error  Possible race condition: `userInfo.status` might be reassigned based on an outdated value of `userInfo.status`  require-atomic-updates
        // userInfo.status = !userInfo.status;

        // Create a new userInfo object with the updated status to ensure atomic update
        const updatedUserInfo = { ...userInfo, status: !userInfo.status };
        // Update userInfo after the status change
        userInfo = updatedUserInfo;
        return Message.error(this.$t("user.610wkzjiq240"));
      }
      return Message.success(this.$t("user.610wkzjiq3w0"));
    },

    // 表格多选
    handleSelectionChange(val) {
      this.multipleSelection = val;
    },

    // 单个删除
    async singleDelete(Id) {
      this.loading = true;
      try {
        await batchDeleteUserByIds({ userIds: [Id] }).then((res) => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getTableData();
    },
    // 单个同步
    async singleSync(Id) {
      this.loading = true;
      try {
        await syncSqlUsers({ userIds: [Id] }).then((res) => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getTableData();
    },

    showPwd() {
      if (this.passwordType === "password") {
        this.passwordType = "";
      } else {
        this.passwordType = "password";
      }
    },

    // 分页
    handleSizeChange(val) {
      this.params.pageSize = val;
      this.getTableData();
    },
    handleCurrentChange(val) {
      this.params.pageNum = val;
      this.getTableData();
    },
    // treeselect
    normalizer(node) {
      return {
        id: node.ID,
        label: node.groupType + "=" + node.groupName,
        isDisabled: node.groupType === "ou" || node.groupName === "root",
        children: node.children
      };
    },
    treeselectInput(value) {
      this.treeselectValue = value;
    },
    syncDingTalkUsers() {
      this.loading = true;
      syncDingTalkUsersApi().then((res) => {
        this.judgeResult(res);
        this.loading = false;
        this.getTableData();
      }).finally(() => {
        this.loading = false;
      });
    },
    syncWeComUsers() {
      this.loading = true;
      syncWeComUsersApi().then((res) => {
        this.judgeResult(res);
        this.loading = false;
        this.getTableData();
      }).finally(() => {
        this.loading = false;
      });
    },
    syncFeiShuUsers() {
      this.loading = true;
      syncFeiShuUsersApi().then((res) => {
        this.judgeResult(res);
        this.loading = false;
        this.getTableData();
      }).finally(() => {
        this.loading = false;
      });
    },
    syncOpenLdapUsers() {
      this.loading = true;
      syncOpenLdapUsersApi().then((res) => {
        this.judgeResult(res);
        this.loading = false;
        this.getTableData();
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

.show-pwd {
  position: absolute;
  right: 10px;
  top: 3px;
  font-size: 16px;
  color: #889aa4;
  cursor: pointer;
  user-select: none;
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
