<template>
  <div class="app-container">
    <div v-if="user">
      <el-row :gutter="20">

        <el-col :span="6" :xs="24">
          <user-card :user="user" />
        </el-col>
        <el-col :span="6" :xs="24">
          <TotpCard :user="user" />
        </el-col>
        <el-col :span="12" :xs="24">
          <Account />
        </el-col>

      </el-row>
    </div>
  </div>
</template>

<script>
import { mapGetters } from "vuex";
import UserCard from "./components/UserCard";
import Account from "./components/Account";
import TotpCard from "./components/TotpCard";

export default {
  name: "Profile",
  components: { UserCard, Account, TotpCard },
  data() {
    return {
      user: {},
      activeTab: "activity"
    };
  },
  computed: {
    ...mapGetters([
      "name",
      "mail",
      "avatar",
      "roles",
      "totp"
    ])
  },
  created() {
    this.getUser();
  },
  methods: {
    getUser() {
      this.user = {
        name: this.name,
        role: this.roles.join(" | "),
        mail: this.mail,
        avatar: this.avatar,
        totp: this.totp
      };
    }
  }
};
</script>
