<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        ref="dnsZoneForm"
        :inline="true"
        size="small"
        :model="dnsZoneForm"
        :rules="dnsZoneFormRules"
      >
        <el-form-item label="Zone:" prop="name">
          <el-input
            v-model.trim="dnsZoneForm.name"
            placeholder="eg: example.com"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            size="small"
            :loading="loading"
            type="primary"
            @click="addZone()"
          >{{ $t('dns.6112hr284k00') }} Zone</el-button>
        </el-form-item>
      </el-form>
      <el-tabs
        v-if="dnsData.length > 0"
        v-model="dnsZoneActiveTab"
        type="border-card"
        closable
        @tab-remove="deleteZone"
      >
        <el-tab-pane
          v-for="item in dnsData"
          :key="item.ID"
          :label="item.name"
          :name="item.name"
        >
          <template>
            <el-form size="mini" :inline="true" class="demo-form-inline">
              <el-form-item>
                <el-button
                  :loading="loading"
                  icon="el-icon-plus"
                  type="warning"
                  @click="addRecord"
                >{{ $t('dns.6112hr2851s0') }}</el-button>
              </el-form-item>
              <el-form-item>
                <el-button
                  :disabled="multipleSelection.length === 0"
                  :loading="loading"
                  icon="el-icon-delete"
                  type="danger"
                  @click="batchDeleteRecords"
                >{{ $t('dns.6112hr2856c0') }}</el-button>
              </el-form-item>
            </el-form>

            <el-table
              v-loading="loading"
              :data="item.records"
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
                prop="type"
                label="Type"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="host"
                label="Host"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="value"
                label="Value"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="ttl"
                label="TTL"
              />
              <el-table-column
                show-overflow-tooltip
                sortable
                prop="creator"
                :label="$t('dns.6112hr285980')"
              />
              <el-table-column
                fixed="right"
                :label="$t('dns.6112hr285c40')"
                align="center"
                width="120"
              >
                <template slot-scope="scope">
                  <el-tooltip :content="$t('dns.6112hr285ew0')" effect="dark" placement="top">
                    <el-button
                      size="mini"
                      icon="el-icon-edit"
                      circle
                      type="primary"
                      @click="updateRecord(scope.row)"
                    />
                  </el-tooltip>
                  <el-tooltip
                    class="delete-popover"
                    :content="$t('dns.6112hr285hs0')"
                    effect="dark"
                    placement="top"
                  >
                    <el-popconfirm
                      :title="$t('dns.6112hr285kk0')"
                      @confirm="deleteRecord(scope.row.ID)"
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
      <el-dialog
        :title="dnsRecordFormTitle"
        :visible.sync="dnsRecordFormVisible"
      >
        <el-form
          ref="dnsRecordForm"
          size="small"
          :model="dnsRecordForm"
          :rules="dnsRecordFormRules"
          label-width="auto"
        >
          <el-form-item label="Zone" prop="zone_id">
            <el-select v-model="dnsRecordForm.zone_id" placeholder="choose one">
              <el-option
                v-for="item in zoneOptions"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="Type" prop="type">
            <el-select
              v-model.trim="dnsRecordForm.type"
              filterable
              default-first-option
              placeholder="choose one"
            >
              <el-option
                v-for="item in typeOptions"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="Host" prop="host">
            <el-input v-model.trim="dnsRecordForm.host" placeholder="eg: www/@" />
          </el-form-item>
          <el-form-item label="value" prop="value">
            <el-input v-model.trim="dnsRecordForm.value" placeholder="value" />
          </el-form-item>
          <el-form-item label="TTL" prop="ttl">
            <el-input
              v-model.number="dnsRecordForm.ttl"
              :placeholder="$t('dns.6112hr285n40')"
            />
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button
            size="mini"
            @click="dnsRecordFormCancel()"
          >{{ $t('dns.6112hr285q00') }}</el-button>
          <el-button
            size="mini"
            :loading="loading"
            type="primary"
            @click="dnsRecordFormSubmit()"
          >{{ $t('dns.6112hr285so0') }}</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import {
  getDnsRecords,
  createDnsZone,
  batchDeleteDnsZoneByIds,
  createDnsRecord,
  updateDnsRecord,
  batchDeleteDnsRecordByIds
} from "@/api/dns/dns";
import { validateSecondLevelDomain } from "@/utils/validate";
import { Message } from "element-ui";

export default {
  name: "DnsManager",
  data() {
    return {
      loading: false,

      dnsData: [], // Dns 数据
      dnsZoneForm: {
        name: ""
      },
      dnsZoneFormRules: {
        name: [
          {
            required: true,
            message: this.$t("dns.inputZoneNameTips"),
            trigger: "blur"
          },
          {
            min: 4,
            max: 20,
            message: this.$t("valid.length", [4, 20]),
            trigger: "blur"
          },
          { required: true, validator: validateSecondLevelDomain, trigger: "blur" }
        ]
      },
      dnsZoneActiveTab: "",

      dnsRecordFormTitle: "",
      dnsRecordFormType: "",
      dnsRecordFormVisible: false,
      dnsRecordForm: {
        id: "",
        zone_id: "",
        type: "",
        host: "",
        value: "",
        ttl: ""
      },
      dnsRecordFormRules: {
        zone_id: [{ required: true, message: this.$t("dns.6112hr285v80"), trigger: "blur" }],
        type: [{ required: true, message: this.$t("dns.6112hr285xs0"), trigger: "blur" }],
        host: [
          { required: true, message: this.$t("dns.inputHostNameTips"), trigger: "blur" },
          {
            min: 1,
            max: 30,
            message: this.$t("valid.length", [1, 30]),
            trigger: "blur"
          }
        ],
        value: [
          { required: true, message: this.$t("dns.6112hr2860o0"), trigger: "change" },
          {
            min: 1,
            max: 100,
            message: this.$t("valid.length", [1, 100]),
            trigger: "blur"
          }
        ],
        ttl: [
          { required: true, message: this.$t("dns.inputTtlTips"), trigger: "blur" },
          { type: "number", message: "must be number", trigger: "blur" }
        ]
      },
      typeOptions: ["A", "CNAME", "TXT"],
      zoneOptions: [],
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
        const { data } = await getDnsRecords();
        this.dnsData = data;

        // console.log(`dnsZoneActiveTab type: ${typeof(this.dnsZoneActiveTab)}, value: ${this.dnsZoneActiveTab}`);
        // default dnsZoneActiveTab type: string, value: 0
        if ((this.dnsZoneActiveTab === "0" || this.dnsZoneActiveTab === "") && this.dnsData.length > 0) {
          this.dnsZoneActiveTab = this.dnsData[0].name;
        }
      } finally {
        this.loading = false;
      }
    },
    getZoneOptions() {
      this.zoneOptions = this.dnsData.map((item) => {
        return {
          id: item.ID,
          name: item.name
        };
      });
    },
    addZone() {
      this.$refs["dnsZoneForm"].validate(async(valid) => {
        if (valid) {
          this.loading = true;
          try {
            await createDnsZone(this.dnsZoneForm).then((res) => {
              this.judgeResult(res);
            });
          } finally {
            this.loading = false;
          }
          this.getData();
        } else {
          Message({
            showClose: true,
            message: this.$t("dns.6112hr2864k0"),
            type: "warn"
          });
          return false;
        }
      });
    },
    // 根据 tabname 查找 zone 的 ID
    getZoneIDFromTabname(tabname) {
      for (let i = 0; i < this.dnsData.length; i++) {
        if (this.dnsData[i].name === tabname) {
          // console.log(this.dnsData[i].ID);
          return this.dnsData[i].ID;
        }
      }
    },
    async deleteZone(tabname) {
      const dnsZoneId = this.getZoneIDFromTabname(tabname);
      this.$confirm(
        this.$t("dns.deleteZoneTips"),
        this.$t("dns.6112hr286740"),
        {
          confirmButtonText: this.$t("dns.6112hr285so0"),
          cancelButtonText: this.$t("dns.6112hr285q00"),
          type: "warning"
        }
      )
        .then(async() => {
          try {
            await batchDeleteDnsZoneByIds({
              ids: [dnsZoneId]
            }).then((res) => {
              this.judgeResult(res);
            });
            this.dnsZoneActiveTab = "0";
            this.getData();
          } finally {
            this.loading = false;
          }
        })
        .catch(() => {
          Message({
            showClose: true,
            type: "info",
            message: this.$t("dns.6112hr2869s0")
          });
        });
    },
    // 新增
    addRecord() {
      this.getZoneOptions();
      const dnsZoneId = this.getZoneIDFromTabname(this.dnsZoneActiveTab);
      this.dnsRecordForm.zone_id = dnsZoneId;

      this.dnsRecordFormTitle = this.$t("dns.6112hr286cc0");
      this.dnsRecordFormType = "add";
      this.dnsRecordFormVisible = true;
    },

    // 修改
    updateRecord(row) {
      this.getZoneOptions();
      this.dnsRecordForm.id = row.ID;
      this.dnsRecordForm.zone_id = row.zone_id;
      this.dnsRecordForm.type = row.type;
      this.dnsRecordForm.host = row.host;
      this.dnsRecordForm.value = row.value;
      this.dnsRecordForm.ttl = row.ttl;

      this.dnsRecordFormTitle = this.$t("dns.6112hr286f40");
      this.dnsRecordFormType = "update";
      this.dnsRecordFormVisible = true;
    },

    // 单个删除
    async deleteRecord(id) {
      this.loading = true;
      try {
        await batchDeleteDnsRecordByIds({ ids: [id] }).then((res) => {
          this.judgeResult(res);
        });
      } finally {
        this.loading = false;
      }
      this.getData();
    },
    // 批量删除
    batchDeleteRecords() {
      this.$confirm(this.$t("tips.deleteWarning"), this.$t("dns.6112hr286740"), {
        confirmButtonText: this.$t("dns.6112hr285so0"),
        cancelButtonText: this.$t("dns.6112hr285q00"),
        type: "warning"
      })
        .then(async() => {
          this.loading = true;
          const ids = [];
          this.multipleSelection.forEach((x) => {
            ids.push(x.ID);
          });
          try {
            await batchDeleteDnsRecordByIds({ ids: ids }).then((res) => {
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
            message: this.$t("dns.6112hr2869s0")
          });
        });
    },
    // 提交表单
    // https://stackoverflow.com/questions/73772552/typeerror-this-refs-is-not-a-function
    dnsRecordFormSubmit() {
      this.$refs["dnsRecordForm"].validate(async(valid) => {
        if (valid) {
          this.loading = true;
          try {
            if (this.dnsRecordFormType === "add") {
              await createDnsRecord(this.dnsRecordForm).then((res) => {
                this.judgeResult(res);
              });
            } else if (this.dnsRecordFormType === "update") {
              await updateDnsRecord(this.dnsRecordForm).then((res) => {
                this.judgeResult(res);
              });
            }
          } finally {
            this.loading = false;
          }
          this.dnsRecordFormReset();
          this.getData();
        } else {
          Message({
            showClose: true,
            message: this.$t("dns.6112hr2864k0"),
            type: "warn"
          });
          return false;
        }
      });
    },

    // 提交表单
    dnsRecordFormCancel() {
      this.dnsRecordFormReset();
    },

    dnsRecordFormReset() {
      this.dnsRecordFormVisible = false;
      this.$refs["dnsRecordForm"].resetFields();
      this.dnsRecordForm = {
        id: "",
        zone_id: "",
        type: "",
        host: "",
        value: "",
        ttl: ""
      };
    },

    // 判断结果
    judgeResult(res) {
      if (res.code === 200) {
        Message({
          showClose: true,
          message: this.$t("dns.6112hr286i40"),
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
