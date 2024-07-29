<template>
  <el-card class="profile-card">
    <div slot="header" class="clearfix">
      <span>{{ $t('profile.aboutme') }}</span>
    </div>

    <div class="user-profile">
      <div class="box-center">
        <img :src="user.avatar" :height="'80px'" :width="'80px'" style="border-radius: 10px;" :hoverable="false">
      </div>
      <div class="box-center">
        <div class="user-name text-center">{{ user.name }}</div>
        <div class="user-name text-center">{{ user.mail }}</div>
        <div class="user-role text-center text-muted">{{ user.role | uppercaseFirst }}</div>
      </div>
    </div>
  </el-card>
</template>

<script>

export default {
  components: {},
  props: {
    user: {
      type: Object,
      default: () => {
        return {
          name: "",
          mail: "",
          avatar: "",
          role: ""
        };
      }
    }
  }
};
</script>

<style lang="scss" scoped>
.profile-card {
  min-height: 18rem;
  height: 18rem;
};
.box-center {
  margin: 0 auto;
  display: table;
}

.text-muted {
  color: #777;
}

.user-profile {
  .user-name {
    font-weight: bold;
  }

  .box-center {
    padding-top: 10px;
  }

  .user-role {
    padding-top: 10px;
    font-weight: 400;
    font-size: 14px;
  }

  .box-social {
    padding-top: 30px;

    .el-table {
      border-top: 1px solid #dfe6ec;
    }
  }

  .user-follow {
    padding-top: 20px;
  }
}

.user-bio {
  margin-top: 20px;
  color: #606266;

  span {
    padding-left: 4px;
  }

  .user-bio-section {
    font-size: 14px;
    padding: 15px 0;

    .user-bio-section-header {
      border-bottom: 1px solid #dfe6ec;
      padding-bottom: 10px;
      margin-bottom: 10px;
      font-weight: bold;
    }
  }
}
</style>
