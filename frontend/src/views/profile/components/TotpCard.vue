<template>
  <el-card class="profile-card">
    <div slot="header" class="clearfix">
      <span>{{ $t('custom.profile.changeTotp') }}</span>
    </div>
    <div class="box-center">
      <div :style="{ display: showDiscribe ? 'block' : 'none' }">
        <div class="warning-message">
          <p><b>{{ $t('custom.totpNotice.notice') }}</b></p>
          <p>{{ $t('custom.totpNotice.content[0]') }}</p>
          <p>{{ $t('custom.totpNotice.content[1]') }}</p>
        </div>
        <el-form
          ref="resetTotpSecretForm"
          size="small"
          :model="resetTotpSecretFormData"
          :rules="resetTotpSecretFormRules"
          label-width="auto"
        >
          <el-form-item label="OTP" prop="totp">
            <el-row type="flex" justify="space-between" align="middle">
              <el-col :xs="16" :sm="14" :md="16" :lg="16" :xl="16">
                <el-input
                  v-model.trim="resetTotpSecretFormData.totp"
                  :placeholder="$t('custom.profile.changeTotpTips')"
                />
              </el-col>

              <el-col :xs="8" :sm="10" :md="8" :lg="8" :xl="8">
                <el-popconfirm
                  :title="$t('custom.common.areyousure')"
                  @confirm="resetTotpSecret"
                >
                  <el-button
                    slot="reference"
                    class="right"
                    size="mini"
                    icon="el-icon-refresh"
                    type="danger"
                  >{{ $t('custom.common.reset') }}
                  </el-button>
                </el-popconfirm>
              </el-col>
            </el-row>
          </el-form-item>
        </el-form>
      </div>
      <!-- <div class="user-name">{{ qrCodeStr }}</div> -->
      <QrCode :id="'QrCode'" class="mt-30" :text="qrCodeStr" />
    </div>
  </el-card>
</template>

<script>
import QrCode from "@/components/Qrcode/Qrcode.vue";
import { resetTotpSecret } from "@/api/personnel/user";

export default {
  components: { QrCode },
  data() {
    return {
      showDiscribe: true,
      qrCodeStr: "",
      resetTotpSecretFormData: {
        totp: ""
      },
      resetTotpSecretFormRules: {
        totp: [
          {
            required: true,
            message: this.$i18n.t("custom.profile.changeTotpTips"),
            trigger: "blur"
          },
          {
            validator: (rule, value, callback) => {
              if (!/^\d{6}$/.test(value)) {
                return callback(new Error(this.$i18n.t("custom.profile.changeTotpTipsValid")));
              }
              callback();
            },
            trigger: "blur"
          }
        ]
      }
    };
  },
  methods: {
    async resetTotpSecret() {
      this.$refs["resetTotpSecretForm"].validate(async(valid) => {
        if (valid) {
          const { code, data } = await resetTotpSecret(
            this.resetTotpSecretFormData
          );
          if (code === 200) {
            this.qrCodeStr = data;
            // console.log(this.qrCodeStr)
            this.showDiscribe = false;
          }
        } else {
          this.$message({
            showClose: true,
            message: this.$i18n.t("custom.profile.changeTotpTipsValidErr"),
            type: "warn"
          });
          return false;
        }
      });
    }
  }
};
</script>

<style lang="scss" scoped>
.profile-card {
  min-height: 18rem;
  height: 18rem;
}
.box-center {
  margin: auto;
  display: table;
}
.right {
  float: right;
}
.mt-30 {
  margin-top: 30px;
}
.warning-message {
  background-color: #fef0f0;
  border: 1px solid #f86c6b;
  border-radius: 5px;
  padding: 10px;
  margin: 0 0 20px 0;
  font-size: smaller;
  p {
    margin: 10px 0;
  }

  b {
    color: #f86c6b;
  }
}
</style>
