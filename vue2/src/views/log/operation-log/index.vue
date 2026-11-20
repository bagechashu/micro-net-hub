<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="mini" :inline="true" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('operation-log.6111tvyrbi80')">
          <el-input v-model.trim="params.username" clearable :placeholder="$t('operation-log.6111tvyrbi80')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('operation-log.ipaddr')">
          <el-input v-model.trim="params.ip" clearable :placeholder="$t('operation-log.ipaddr')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('operation-log.6111tvyrbt00')">
          <el-input v-model.trim="params.path" clearable :placeholder="$t('operation-log.6111tvyrbt00')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('operation-log.6111tvyrbw80')">
          <el-input v-model.trim="params.status" clearable :placeholder="$t('operation-log.6111tvyrbw80')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" icon="el-icon-search" type="primary" @click="search">{{ $t('operation-log.6111tvyrbyo0') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :disabled="multipleSelection.length === 0" :loading="loading" icon="el-icon-delete" type="danger" @click="batchDelete">{{ $t('operation-log.6111tvyrc0o0') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column show-overflow-tooltip sortable prop="username" :label="$t('operation-log.6111tvyrbi80')" width="120" />
        <el-table-column show-overflow-tooltip sortable prop="ip" :label="$t('operation-log.ipaddr')" width="120" />
        <el-table-column show-overflow-tooltip sortable prop="method" :label="$t('operation-log.6111tvyrc2s0')" width="100" />
        <el-table-column show-overflow-tooltip sortable prop="path" :label="$t('operation-log.6111tvyrbt00')" />
        <el-table-column show-overflow-tooltip sortable prop="status" :label="$t('operation-log.6111tvyrbw80')" width="100" align="center">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.status | statusTagFilter" disable-transitions>{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="startTime" :label="$t('operation-log.6111tvyrc4w0')" width="300">
          <!-- <template slot-scope="scope">
            {{ parseGoTime(scope.row.startTime) }}
          </template> -->
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="timeCost" :label="$t('operation-log.6111tvyrc8s0')" width="130" align="center">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.timeCost | timeCostTagFilter" disable-transitions>{{ scope.row.timeCost }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="desc" :label="$t('operation-log.6111tvyrcaw0')" />
        <el-table-column fixed="right" :label="$t('operation-log.6111tvyrccw0')" align="center" width="80">
          <template slot-scope="scope">
            <el-tooltip :content="$t('operation-log.6111tvyrcf40')" effect="dark" placement="top">
              <el-popconfirm :title="$t('operation-log.6111tvyrchk0')" @confirm="singleDelete(scope.row.ID)">
                <el-button slot="reference" size="mini" icon="el-icon-delete" circle type="danger" />
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
        style="margin-top: 10px;float:right;margin-bottom: 10px;"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>
  </div>
</template>

<script>
import { getOperationLogs, batchDeleteOperationLogByIds } from "@/api/log/operationLog";
import { parseGoTime } from "@/utils/index";
import { Message } from "element-ui";

export default {
  name: "OperationLog",
  filters: {
    statusTagFilter(val) {
      if (val === 200) {
        return "success";
      } else if (val === 400) {
        return "warning";
      } else if (val === 401) {
        return "danger";
      } else if (val === 403) {
        return "danger";
      } else if (val === 500) {
        return "danger";
      } else {
        return "info";
      }
    },
    timeCostTagFilter(val) {
      if (val <= 200) {
        return "success";
      } else if (val > 200 && val <= 1000) {
        return "";
      } else if (val > 1000 && val <= 2000) {
        return "warning";
      } else {
        return "danger";
      }
    }
  },
  data() {
    return {
      // 查询参数
      params: {
        username: "",
        ip: "",
        path: "",
        status: "",
        pageNum: 1,
        pageSize: 10
      },
      // 表格数据
      tableData: [],
      total: 0,
      loading: false,

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
    parseGoTime,
    // 查询
    search() {
      this.params.pageNum = 1;
      this.getTableData();
    },

    // 获取表格数据
    async getTableData() {
      this.loading = true;
      try {
        const { data } = await getOperationLogs(this.params);
        this.tableData = data.logs;
        this.total = data.total;
      } finally {
        this.loading = false;
      }
    },

    // 判断结果
    judgeResult(res) {
      if (res.code === 200 || res.code === 0) {
        Message({
          showClose: true,
          message: this.$t("operation-log.6111tvyrcjk0"),
          type: "success"
        });
      }
    },

    // 批量删除
    batchDelete() {
      this.$confirm(this.$t("tips.deleteWarning"), this.$t("operation-log.6111tvyrcls0"), {
        confirmButtonText: this.$t("operation-log.6111tvyrcnw0"),
        cancelButtonText: this.$t("operation-log.6111tvyrcq40"),
        type: "warning"
      }).then(async res => {
        this.loading = true;
        const operationLogIds = [];
        this.multipleSelection.forEach(x => {
          operationLogIds.push(x.ID);
        });
        try {
          await batchDeleteOperationLogByIds({ operationLogIds: operationLogIds }).then(res => {
            this.judgeResult(res);
          });
        } finally {
          this.loading = false;
        }
        this.getTableData();
      }).catch(() => {
        Message({
          showClose: true,
          type: "info",
          message: this.$t("operation-log.6111tvyrcs40")
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
        await batchDeleteOperationLogByIds({ operationLogIds: [Id] }).then(res => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getTableData();
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
  .container-card{
    margin: 10px;
    margin-bottom: 100px;
  }

  .delete-popover{
    margin-left: 10px;
  }
</style>
