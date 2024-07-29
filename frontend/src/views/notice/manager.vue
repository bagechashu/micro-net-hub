<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="mini" :inline="true">
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-plus"
            type="warning"
            @click="addNotice"
          >新增</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            icon="el-icon-delete"
            type="danger"
            @click="batchDeleteNotices"
          >批量删除</el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-if="noticeData.length > 0"
        v-loading="loading"
        :data="noticeData"
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
          prop="level"
          label="级别"
          width="80"
          align="center"
        >
          <template #default="scope">
            <span :style="{ color: getLevelColor(scope.row.level) , fontWeight: 'bold' }">
              {{ getLevelText(scope.row.level) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="content"
          label="内容"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="creator"
          label="创建人"
          width="100"
        />
        <el-table-column fixed="right" label="操作" align="center" width="120">
          <template slot-scope="scope">
            <el-tooltip content="编辑" effect="dark" placement="top">
              <el-button
                size="mini"
                icon="el-icon-edit"
                circle
                type="primary"
                @click="updateNotice(scope.row)"
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
                @confirm="deleteNotice(scope.row.ID)"
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
      <el-empty v-else :description="$t('common.nodata')" />

      <el-dialog :title="noticeFormTitle" :visible.sync="noticeFormVisible">
        <el-form
          ref="noticeForm"
          size="small"
          :model="noticeForm"
          :rules="noticeFormRules"
          label-width="auto"
        >
          <el-form-item label="级别" prop="level">
            <el-select v-model="noticeForm.level" placeholder="请选择">
              <el-option
                v-for="item in levelOptions"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="内容" prop="content">
            <el-input
              v-model.trim="noticeForm.content"
              type="textarea"
              placeholder="内容"
              show-word-limit
              maxlength="100"
            />
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="noticeFormCancel()">取 消</el-button>
          <el-button
            size="mini"
            :loading="loading"
            type="primary"
            @click="noticeFormSubmit()"
          >确 定</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import {
  getNotice,
  createNotice,
  updateNotice,
  batchDeleteNoticeByIds
} from "@/api/notice/notice";
import { Message } from "element-ui";

export default {
  name: "NoticeManager",
  data() {
    return {
      loading: false,

      noticeData: [], // 导航页数据

      noticeFormTitle: "",
      noticeFormType: "",
      noticeFormVisible: false,
      noticeForm: {
        level: 1,
        content: ""
      },
      noticeFormRules: {
        level: [{ required: true, message: "必须选择级别", trigger: "blur" }],
        content: [
          { required: true, message: "输入内容", trigger: "blur" },
          {
            min: 10,
            max: 100,
            message: "长度在 10 到 50 个字符",
            trigger: "blur"
          }
        ]
      },
      levelOptions: [
        { name: "一般", id: 1 },
        { name: "普通", id: 2 },
        { name: "重要", id: 3 },
        { name: "紧急", id: 4 }
      ],
      // 表格多选
      multipleSelection: []
    };
  },
  created() {
    this.getData();
  },
  methods: {
    // 获取表格数据
    async getData() {
      this.loading = true;
      try {
        const { data } = await getNotice();
        this.noticeData = data;
      } finally {
        this.loading = false;
      }
    },
    getLevelText(level) {
      const option = this.levelOptions.find(option => option.id === level);
      return option ? option.name : "未知";
    },
    getLevelColor(level) {
      switch (level) {
        case 1:
          return "#999999"; // 一般
        case 2:
          return "#66bb6a"; // 普通
        case 3:
          return "#ffa726"; // 重要
        case 4:
          return "#e53935"; // 紧急
        default:
          return "#999999"; // 默认黑色
      }
    },
    // 新增
    addNotice() {
      this.noticeFormTitle = "新增站点";
      this.noticeFormType = "add";
      this.noticeFormVisible = true;
    },

    // 修改
    updateNotice(row) {
      this.noticeForm.id = row.ID;
      this.noticeForm.level = row.level;
      this.noticeForm.content = row.content;

      this.noticeFormTitle = "更新站点";
      this.noticeFormType = "update";
      this.noticeFormVisible = true;
    },

    // 单个删除
    async deleteNotice(id) {
      this.loading = true;
      try {
        await batchDeleteNoticeByIds({ ids: [id] }).then((res) => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getData();
    },
    // 批量删除
    batchDeleteNotices() {
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
            await batchDeleteNoticeByIds({ ids: ids }).then((res) => {
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
    noticeFormSubmit() {
      this.$refs["noticeForm"].validate(async(valid) => {
        if (valid) {
          this.loading = true;
          try {
            if (this.noticeFormType === "add") {
              await createNotice(this.noticeForm).then((res) => {
                this.judgeResult(res);
              });
            } else if (this.noticeFormType === "update") {
              await updateNotice(this.noticeForm).then((res) => {
                this.judgeResult(res);
              });
            }
          } finally {
            this.loading = false;
          }
          this.noticeFormReset();
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
    noticeFormCancel() {
      this.noticeFormReset();
    },

    noticeFormReset() {
      this.noticeFormVisible = false;
      this.$refs["noticeForm"].resetFields();
      this.noticeForm = {
        level: 1,
        content: ""
      };
    },

    // 判断结果
    judgeResult(res) {
      if (res.code === 200) {
        const message = res.msg ? res.msg : "操作成功";
        Message({
          showClose: true,
          message: message,
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
