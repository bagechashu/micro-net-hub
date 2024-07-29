<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        size="mini"
        :inline="true"
        :model="params"
        class="demo-form-inline"
      >
        <el-form-item :label="$t('fieldRelation.610zu0w6ymo0')">
          <el-input
            v-model.trim="params.remark"
            clearable
            :placeholder="$t('fieldRelation.610zu0w6z440')"
            @keyup.enter.native="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-search"
            type="primary"
            @click="search"
          >{{ $t('fieldRelation.610zu0w6z7w0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-plus"
            type="warning"
            @click="create"
          >{{ $t('fieldRelation.610zu0w6zag0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            icon="el-icon-delete"
            type="danger"
            @click="batchDelete"
          >{{ $t('fieldRelation.610zu0w6zdw0') }}</el-button>
        </el-form-item>
        <br>
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
          prop="ID"
          :label="$t('fieldRelation.610zu0w6zgo0')"
          width="80"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="CreatedAt"
          :label="$t('fieldRelation.610zu0w6ziw0')"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="Flag"
          :label="$t('fieldRelation.610zu0w6ymo0')"
        />
        <el-table-column show-overflow-tooltip sortable :label="$t('fieldRelation.610zu0w6zm80')">
          <template slot-scope="props">
            <el-form>
              <el-form-item>
                <span>{{ props.row.Attributes }}</span>
              </el-form-item>
            </el-form>
          </template>
        </el-table-column>
        <el-table-column fixed="right" :label="$t('fieldRelation.610zu0w6zos0')" align="center" width="150">
          <template #default="scope">
            <el-tooltip :content="$t('fieldRelation.610zu0w6zr00')" effect="dark" placement="top">
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
              :content="$t('fieldRelation.610zu0w6zu00')"
              effect="dark"
              placement="top"
            >
              <el-popconfirm
                :title="$t('fieldRelation.610zu0w6zwk0')"
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
          </template>
        </el-table-column>
      </el-table>

      <!-- 新增 -->
      <el-dialog :title="dialogFormTitle" :visible.sync="updateLoading">
        <div class="components-container">
          <aside>
            {{ $t('fieldRelation.610zu0w6zyo0') }}
            <a
              href="http://ldapdoc.eryajf.net/pages/84953d/"
              target="_blank"
            >{{ $t('fieldRelation.610zu0w70140') }}</a>
          </aside>
        </div>
        <el-form
          ref="dialogForm"
          size="small"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-width="180px"
        >
          <el-form-item :label="$t('fieldRelation.610zu0w703g0')">
            <el-checkbox-group v-model="fieldRelationChecked">
              <el-checkbox-button
                v-for="type in fieldRelationTypes"
                :key="type"
                :label="type"
                @change="fieldRelationCheck(type)"
              >
                {{ type }}
              </el-checkbox-button>
            </el-checkbox-group>
          </el-form-item>

          <template v-if="fieldRelationChecked.length === 1 && fieldRelationChecked[0] === 'userFieldRelation'">
            <el-form-item :label="$t('fieldRelation.610zu0w708g0')">
              <el-select
                v-model="userVal"
                :placeholder="$t('fieldRelation.610zu0w70ao0')"
                @change="changeUser(userVal)"
              >
                <el-option
                  v-for="item in userOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70cs0')" prop="username">
              <el-input
                v-model.trim="dialogFormData.username"
                :placeholder="$t('fieldRelation.610zu0w70cs0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70fc0')" prop="nickname">
              <el-input
                v-model.trim="dialogFormData.nickname"
                :placeholder="$t('fieldRelation.610zu0w70fc0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70ik0')" prop="givenName">
              <el-input
                v-model.trim="dialogFormData.givenName"
                :placeholder="$t('fieldRelation.610zu0w70ik0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70kk0')" prop="mail">
              <el-input v-model.trim="dialogFormData.mail" :placeholder="$t('fieldRelation.610zu0w70kk0')" />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70nc0')" prop="jobNumber">
              <el-input
                v-model.trim="dialogFormData.jobNumber"
                :placeholder="$t('fieldRelation.610zu0w70nc0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70p80')" prop="mobile">
              <el-input
                v-model.trim="dialogFormData.mobile"
                :placeholder="$t('fieldRelation.610zu0w70p80')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70qw0')" prop="avatar">
              <el-input
                v-model.trim="dialogFormData.avatar"
                :placeholder="$t('fieldRelation.610zu0w70qw0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70sw0')" prop="postalAddress">
              <el-input
                v-model.trim="dialogFormData.postalAddress"
                :placeholder="$t('fieldRelation.610zu0w70sw0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70uo0')" prop="position">
              <el-input
                v-model.trim="dialogFormData.position"
                :placeholder="$t('fieldRelation.610zu0w70uo0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70ws0')" prop="sourceUserId">
              <el-input
                v-model.trim="dialogFormData.sourceUserId"
                :placeholder="$t('fieldRelation.610zu0w70ws0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70ys0')" prop="sourceUnionId">
              <el-input
                v-model.trim="dialogFormData.sourceUnionId"
                :placeholder="$t('fieldRelation.610zu0w70ys0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w710s0')" prop="introduction">
              <el-input
                v-model.trim="dialogFormData.introduction"
                :placeholder="$t('fieldRelation.610zu0w710s0')"
              />
            </el-form-item>
            <!-- <el-form-item label="说明" prop="introduction">
              <el-input
                v-model.trim="dialogFormData.introduction"
                type="textarea"
                :placeholder="$t('fieldRelation.610zu0w710s0')"
                :autosize="{ minRows: 3, maxRows: 6 }"
                show-word-limit
                maxlength="100"
              />
            </el-form-item> -->
          </template>
          <template v-else-if="fieldRelationChecked.length === 1 && fieldRelationChecked[0] === 'groupFieldRelation'">
            <el-form-item :label="$t('fieldRelation.610zu0w708g0')">
              <el-select
                v-model="groupVal"
                :placeholder="$t('fieldRelation.610zu0w70ao0')"
                @change="changeGroup(groupVal)"
              >
                <el-option
                  v-for="item in options"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w714g0')" prop="groupName">
              <el-input
                v-model.trim="dialogFormData.groupName"
                :placeholder="$t('fieldRelation.610zu0w714g0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w716c0')" prop="sourceDeptParentId">
              <el-input
                v-model.trim="dialogFormData.sourceDeptParentId"
                :placeholder="$t('fieldRelation.610zu0w716c0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w71840')" prop="sourceDeptId">
              <el-input
                v-model.trim="dialogFormData.sourceDeptId"
                :placeholder="$t('fieldRelation.610zu0w71840')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w71as0')" prop="remark">
              <el-input
                v-model.trim="dialogFormData.remark"
                :placeholder="$t('fieldRelation.610zu0w71as0')"
              />
            </el-form-item>
          </template>
          <template v-else>
            <el-form-item><b>↑ {{ $t('fieldRelation.610zu0w71ck0') }}</b></el-form-item>
          </template>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('fieldRelation.610zu0w71ec0') }}</el-button>
          <el-button
            size="mini"
            :loading="submitLoading"
            type="primary"
            @click="submitForm('A')"
          >{{ $t('fieldRelation.610zu0w71gc0') }}</el-button>
        </div>
      </el-dialog>

      <!-- 编辑 -->
      <el-dialog :title="dialogFormTitle" :visible.sync="dialogFormVisible">
        <div class="components-container">
          <aside>
            {{ $t('fieldRelation.610zu0w6zyo0') }}
            <a
              href="http://ldapdoc.eryajf.net/pages/84953d/"
              target="_blank"
            >{{ $t('fieldRelation.610zu0w70140') }}</a>
          </aside>
        </div>
        <el-form
          ref="dialogForm"
          size="small"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-width="180px"
        >
          <template v-if="fieldRelationChecked.length === 1 && fieldRelationChecked[0] === 'userFieldRelation'">
            <el-form-item :label="$t('fieldRelation.610zu0w703g0')">
              <el-button type="primary">{{ $t('fieldRelation.610zu0w70600') }}</el-button>
            </el-form-item>

            <el-form-item :label="$t('fieldRelation.610zu0w708g0')">
              <el-select
                v-model="userVal"
                :placeholder="$t('fieldRelation.610zu0w70ao0')"
                @change="changeUser(userVal)"
              >
                <el-option
                  v-for="item in userOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70cs0')" prop="username">
              <el-input
                v-model.trim="dialogFormData.username"
                :placeholder="$t('fieldRelation.610zu0w70cs0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70fc0')" prop="nickname">
              <el-input
                v-model.trim="dialogFormData.nickname"
                :placeholder="$t('fieldRelation.610zu0w70fc0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70ik0')" prop="givenName">
              <el-input
                v-model.trim="dialogFormData.givenName"
                :placeholder="$t('fieldRelation.610zu0w70ik0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70kk0')" prop="mail">
              <el-input v-model.trim="dialogFormData.mail" :placeholder="$t('fieldRelation.610zu0w70kk0')" />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70nc0')" prop="jobNumber">
              <el-input
                v-model.trim="dialogFormData.jobNumber"
                :placeholder="$t('fieldRelation.610zu0w70nc0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70p80')" prop="mobile">
              <el-input
                v-model.trim="dialogFormData.mobile"
                :placeholder="$t('fieldRelation.610zu0w70p80')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70qw0')" prop="avatar">
              <el-input
                v-model.trim="dialogFormData.avatar"
                :placeholder="$t('fieldRelation.610zu0w70qw0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70sw0')" prop="postalAddress">
              <el-input
                v-model.trim="dialogFormData.postalAddress"
                :placeholder="$t('fieldRelation.610zu0w70sw0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70uo0')" prop="position">
              <el-input
                v-model.trim="dialogFormData.position"
                :placeholder="$t('fieldRelation.610zu0w70uo0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70ws0')" prop="sourceUserId">
              <el-input
                v-model.trim="dialogFormData.sourceUserId"
                :placeholder="$t('fieldRelation.610zu0w70ws0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w70ys0')" prop="sourceUnionId">
              <el-input
                v-model.trim="dialogFormData.sourceUnionId"
                :placeholder="$t('fieldRelation.610zu0w70ys0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w710s0')" prop="introduction">
              <el-input
                v-model.trim="dialogFormData.introduction"
                :placeholder="$t('fieldRelation.610zu0w710s0')"
              />
            </el-form-item>
          </template>
          <template v-else>
            <el-form-item :label="$t('fieldRelation.610zu0w703g0')">
              <el-button type="primary">{{ $t('fieldRelation.610zu0w712o0') }}</el-button>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w708g0')">
              <el-select
                v-model="groupVal"
                :placeholder="$t('fieldRelation.610zu0w70ao0')"
                @change="changeGroup(groupVal)"
              >
                <el-option
                  v-for="item in options"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w714g0')" prop="groupName">
              <el-input
                v-model.trim="dialogFormData.groupName"
                :placeholder="$t('fieldRelation.610zu0w714g0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w716c0')" prop="sourceDeptParentId">
              <el-input
                v-model.trim="dialogFormData.sourceDeptParentId"
                :placeholder="$t('fieldRelation.610zu0w716c0')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w71840')" prop="sourceDeptId">
              <el-input
                v-model.trim="dialogFormData.sourceDeptId"
                :placeholder="$t('fieldRelation.610zu0w71840')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.610zu0w71as0')" prop="remark">
              <el-input
                v-model.trim="dialogFormData.remark"
                :placeholder="$t('fieldRelation.610zu0w71as0')"
              />
            </el-form-item>
          </template>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('fieldRelation.610zu0w71ec0') }}</el-button>
          <el-button
            size="mini"
            :loading="submitLoading"
            type="primary"
            @click="submitForm('B')"
          >{{ $t('fieldRelation.610zu0w71gc0') }}</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
// import Treeselect from '@riophae/vue-treeselect'
// import '@riophae/vue-treeselect/dist/vue-treeselect.css'
import {
  relationList,
  relationAdd,
  relationUp,
  relationDel
} from "@/api/personnel/fieldRelation";
import { Message } from "element-ui";

export default {
  name: "FieldRelation",
  components: {
    // Treeselect
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
    // var checkPhone = (rule, value, callback) => {
    //   if (!value) {
    //     return callback(new Error('手机号不能为空'))
    //   } else {
    //     const reg = /^(\+|00)??(\d{1,3})??((1|0)\d{8,10})??$/
    //     if (reg.test(value)) {
    //       callback()
    //     } else {
    //       return callback(new Error('请输入正确的手机号'))
    //     }
    //   }
    // }
    return {
      options: [
        { label: this.$t("fieldRelation.610zu0w71i00"), value: "feishu_group" },
        { label: this.$t("fieldRelation.610zu0w71jo0"), value: "dingtalk_group" },
        { label: this.$t("fieldRelation.610zu0w71lg0"), value: "wecom_group" }
      ],
      userOptions: [
        { label: this.$t("fieldRelation.610zu0w71i00"), value: "feishu_user" },
        { label: this.$t("fieldRelation.610zu0w71jo0"), value: "dingtalk_user" },
        { label: this.$t("fieldRelation.610zu0w71lg0"), value: "wecom_user" }
      ],
      userVal: "",
      groupVal: "",
      updateId: "",
      fieldRelationChecked: ["userFieldRelation"], // 新增数据默认选中
      fieldRelationTypes: ["userFieldRelation", "groupFieldRelation"], // 新增默认选中
      // 查询参数
      params: {
        flag: "",
        pageNum: 1,
        pageSize: 1000 // 平常百姓人家应该不会有这么多数据吧,后台限制最大单次获取1000条
      },
      // 表格数据
      tableData: [],
      infoTableData: [],
      total: 0,
      loading: false,
      // 上级目录数据
      // treeselectData: [],
      // treeselectValue: 0,
      updateLoading: false, // 新增
      // dialog对话框
      submitLoading: false,
      dialogFormTitle: "",
      dialogType: "",
      dialogFormVisible: false,
      dialogFormData: {
        username: "", // 用户名(通常为用户名拼音) name_pinyin
        nickname: "", // 中文名字 name
        givenName: "", // 花名 name
        mail: "", // 邮箱 email
        jobNumber: "", // 工号 job_number
        mobile: "", // 手机号 mobile
        avatar: "", // 头像 avatar
        postalAddress: "", // 地址 work_place
        position: "", // 职位 title
        introduction: "", // 说明 remark
        sourceUserId: "", // 源用户ID  userid
        sourceUnionId: "", // 源用户唯一ID   unionid
        groupName: "", // 分组名称（通常为分组名的拼音）
        remark: "", // 分组描述
        sourceDeptId: "", // 部门ID
        sourceDeptParentId: "" // 父部门ID
      },
      //   dialogFromGroup: {

      //   },
      dialogFormRules: {
        sourceDeptParentId: [
          { required: true, message: this.$t("fieldRelation.610zu0w71n40"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        sourceDeptId: [
          { required: true, message: this.$t("fieldRelation.610zu0w71ow0"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        username: [
          { required: true, message: this.$t("fieldRelation.610zu0w71r40"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        givenName: [
          { required: true, message: this.$t("fieldRelation.610zu0w71to0"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        avatar: [
          { required: true, message: this.$t("fieldRelation.610zu0w71to0"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        postalAddress: [
          { required: true, message: this.$t("fieldRelation.610zu0w71to0"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        position: [
          { required: true, message: this.$t("fieldRelation.610zu0w71to0"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        sourceUserId: [
          { required: true, message: this.$t("fieldRelation.610zu0w71to0"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: ["blur", "change"]
          }
        ],
        sourceUnionId: [
          { required: true, message: this.$t("fieldRelation.610zu0w71to0"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: ["blur", "change"]
          }
        ],
        groupName: [
          { required: true, message: this.$t("fieldRelation.610zu0w71w00"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        remark: [
          { required: true, message: this.$t("fieldRelation.610zu0w71xs0"), trigger: "blur" },
          {
            min: 1,
            max: 50,
            message: this.$t("valid.length", [1, 50]),
            trigger: "blur"
          }
        ],
        // mail: [
        //   { required: true, message: '请输入邮箱', trigger: 'blur' },
        //   { type: 'email', message: '请输入正确的邮箱地址', trigger: ['blur', 'change'] }
        // ],
        mail: [
          { required: true, message: this.$t("fieldRelation.610zu0w71zg0"), trigger: "blur" },
          { min: 1, max: 50, message: this.$t("valid.length", [1, 50]), trigger: "blur" }
        ],
        jobNumber: [
          { required: false, message: this.$t("fieldRelation.610zu0w722w0"), trigger: "blur" },
          {
            min: 0,
            max: 20,
            message: this.$t("valid.length", [2, 20]),
            trigger: "blur"
          }
        ],
        nickname: [
          { required: true, message: this.$t("fieldRelation.610zu0w724o0"), trigger: "blur" },
          {
            min: 2,
            max: 20,
            message: this.$t("valid.length", [2, 20]),
            trigger: "blur"
          }
        ],
        mobile: [{ required: false, message: this.$t("fieldRelation.610zu0w726g0"), trigger: "blur" }],
        introduction: [
          { required: false, message: this.$t("fieldRelation.610zu0w710s0"), trigger: "blur" },
          {
            min: 0,
            max: 100,
            message: this.$t("valid.length", [0, 100]),
            trigger: "blur"
          }
        ]
      },
      // 表格多选
      multipleSelection: []
      // typeFlag:
    };
  },
  created() {
    this.getTableData();
  },
  methods: {
    fieldRelationCheck(type) {
      this.fieldRelationChecked = this.fieldRelationChecked.includes(type) ? [type] : [];
      // this.value = this.type;
    },
    changeUser(e) {
      this.userVal = e;
    },
    changeGroup(e) {
      this.groupVal = e;
    },
    // 查询
    search() {
      // 初始化表格数据
      this.infoTableData = JSON.parse(JSON.stringify(this.tableData));
      this.infoTableData = this.deal(this.infoTableData, (node) =>
        node.Flag.includes(this.params.flag)
      );
    },
    resetData() {
      this.infoTableData = JSON.parse(JSON.stringify(this.tableData));
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
    async getTableData() {
      this.loading = true;
      try {
        const { data } = await relationList(this.params);
        this.tableData = data;

        this.infoTableData = JSON.parse(JSON.stringify(data));
      } finally {
        this.loading = false;
      }
    },

    // 新增
    create() {
      this.fieldRelationChecked = ["userFieldRelation"];
      this.userVal = "";
      this.groupVal = "";
      this.dialogFormData = {};
      this.dialogFromGroup = {};
      this.dialogFormTitle = this.$t("fieldRelation.610zu0w6zag0");
      this.updateLoading = true; // 新增的展示
      this.dialogType = "create";
    },
    // 修改
    update(row) {
      const typeDialog = row.Flag.split("_")[1];

      const {
        avatar,
        givenName,
        introduction,
        jobNumber,
        mail,
        mobile,
        nickname,
        position,
        postalAddress,
        sourceUnionId,
        sourceUserId,
        username,
        groupName,
        remark,
        sourceDeptId,
        sourceDeptParentId
      } = row.Attributes;

      if (typeDialog === "user") {
        this.updateId = row.ID;
        this.fieldRelationChecked = ["userFieldRelation"];

        this.userVal = row.Flag;
        this.dialogFormData.username = username; // 用户名(通常为用户名拼音) name_pinyin
        this.dialogFormData.nickname = nickname; // 中文名字 name
        this.dialogFormData.givenName = givenName; // 花名 name
        this.dialogFormData.mail = mail; // 邮箱 email
        this.dialogFormData.jobNumber = jobNumber; // 工号 job_number
        this.dialogFormData.mobile = mobile; // 手机号 mobile
        this.dialogFormData.avatar = avatar; // 头像 avatar
        this.dialogFormData.postalAddress = postalAddress; // 地址 work_place
        this.dialogFormData.position = position; // 职位 title
        this.dialogFormData.introduction = introduction; // 说明 remark
        this.dialogFormData.sourceUserId = sourceUserId; // 源用户ID  userid
        this.dialogFormData.sourceUnionId = sourceUnionId; // 源用户唯一ID   unionid
      } else {
        this.updateId = row.ID;
        this.fieldRelationChecked = ["groupFieldRelation"];
        this.groupVal = row.Flag;
        this.dialogFormData.groupName = groupName; // 分组名称（通常为分组名的拼音）
        this.dialogFormData.remark = remark; // 分组描述
        this.dialogFormData.sourceDeptId = sourceDeptId; // 部门ID
        this.dialogFormData.sourceDeptParentId = sourceDeptParentId; // 父部门ID
      }

      this.dialogFormTitle = this.$t("fieldRelation.610zu0w72880");
      this.dialogType = "update";
      this.dialogFormVisible = true;
    },

    // 提交表单
    submitForm(e) {
      let flag, attributes;
      if (this.fieldRelationChecked[0] === "userFieldRelation") {
        if (this.userVal === "") {
          Message({
            message: this.$t("fieldRelation.610zu0w72a40"),
            type: "warning"
          });
          return false;
        }
        flag = this.userVal;
        attributes = this.dialogFormData;
      } else {
        if (this.groupVal === "") {
          Message({
            message: this.$t("fieldRelation.610zu0w72a40"),
            type: "warning"
          });
          return false;
        }
        flag = this.groupVal;
        attributes = this.dialogFormData;
      }
      this.$refs["dialogForm"].validate(async(valid) => {
        if (valid) {
          this.submitLoading = true;
          try {
            if (this.dialogType === "create") {
              await relationAdd({
                flag: flag,
                attributes: attributes
              });
            } else {
              await relationUp({
                id: this.updateId,
                flag: flag,
                attributes: attributes
              });
            }
          } finally {
            this.submitLoading = false;
          }
          this.resetForm();
          this.getTableData();
          Message({
            showClose: true,
            message: this.$t("fieldRelation.610zu0w72bs0"),
            type: "success"
          });
        } else {
          Message({
            showClose: true,
            message: this.$t("fieldRelation.610zu0w72dg0"),
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
      this.updateLoading = false;
      this.$refs["dialogForm"].resetFields();
      this.dialogFormData = {
        groupName: "",
        remark: ""
      };
    },

    // 批量删除
    batchDelete() {
      this.$confirm(this.$t("fieldRelation.batchDeleteTips"), this.$t("fieldRelation.610zu0w72fk0"), {
        confirmButtonText: this.$t("fieldRelation.610zu0w71gc0"),
        cancelButtonText: this.$t("fieldRelation.610zu0w71ec0"),
        type: "warning"
      })
        .then(async(res) => {
          this.loading = true;
          const groupIds = [];
          this.multipleSelection.forEach((x) => {
            groupIds.push(x.ID);
          });
          try {
            await relationDel({ fieldRelationIds: groupIds });
          } finally {
            this.loading = false;
          }
          this.getTableData();
          Message({
            showClose: true,
            message: this.$t("fieldRelation.610zu0w72hg0"),
            type: "success"
          });
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: this.$t("fieldRelation.610zu0w72jk0")
          });
        });
    },
    // 单个删除
    async singleDelete(Id) {
      this.loading = true;
      try {
        await relationDel({ fieldRelationIds: [Id] });
      } finally {
        this.loading = false;
      }
      this.getTableData();
    },

    // 表格多选
    handleSelectionChange(val) {
      this.multipleSelection = val;
    },

    // 分页
    handleSizeChange(val) {
      this.params.pageSize = val;
      this.getTableData();
    },
    handleCurrentChange(val) {
      this.params.pageNum = val;
      this.getTableData();
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
.demo-table-expand {
  font-size: 0;
}
.demo-table-expand label {
  width: 90px;
  color: #99a9bf;
  text-align: left !important;
}
.demo-table-expand .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
  width: 50%;
}
.link-title {
  margin-left: 30px;
  margin-bottom: 10px;
}

/* .el-form-item /deep/ label{
    label{

    }

} */
</style>
