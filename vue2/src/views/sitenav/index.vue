<template>
  <div class="app-container">
    <el-row>
      <el-col :xs="24" :sm="12" :md="10" :lg="10" :xl="6">
        <div class="header-bar">
          <el-input
            v-model="search"
            :placeholder="$t('sitenav.searchTips')"
            class="search"
            @keyup.enter.native="searchData"
          />
          <span
            class="search-text"
          ><el-button type="primary" icon="search" @click="searchData">{{
            $t("common.search")
          }}</el-button></span>
          <el-button
            v-show="searchStatus"
            type="success"
            icon="plus-round"
            @click="resetSearch"
          >{{ $t("common.reset") }}</el-button>
        </div></el-col>
      <el-col :xs="24" :sm="10" :md="13" :lg="13" :xl="17"><Notice /></el-col>
    </el-row>

    <NavSub v-if="data.length > 0" :data="data" />
    <el-empty v-else :description="$t('common.nodata')" />
  </div>
</template>

<script>
import { Message } from "element-ui";
import NavSub from "@/views/sitenav/components/sub";
import Notice from "@/views/notice";
import { getNav } from "@/api/sitenav/sitenav";
export default {
  name: "SiteNav",
  components: {
    NavSub,
    Notice
  },
  data() {
    return {
      isCollapsed: false,
      search: "",
      searchStatus: false,
      data: [],
      sourceData: "",
      searchNum: 0
    };
  },
  created: function() {
    this._getNavJson();
  },
  methods: {
    async _getNavJson() {
      try {
        const { data } = await getNav();
        this.data = data;
      } catch (e) {
        console.log(e);
      }
    },
    jumpAnchor(name) {
      if (document.documentElement.clientWidth <= 768) {
        this.isCollapsed = true;
      }

      const offset = 66;
      const el = document.querySelector("#" + name);
      window.scroll({
        top: el.offsetTop - offset,
        left: 0,
        behavior: "smooth"
      });
    },
    searchData() {
      if (
        typeof this.search === "undefined" ||
        this.search === null ||
        this.search.trim() === ""
      ) {
        Message({
          message: this.$t("valid.pleaseInput"),
          type: "error"
        });
        return;
      }

      if (!this.searchStatus) {
        this.sourceData = JSON.parse(JSON.stringify(this.data));
      } else {
        this.data = JSON.parse(JSON.stringify(this.sourceData));
      }

      this.searchStatus = true;
      this.searchNum = 0;

      this.data = this.data
        .map((group) => {
          const filteredSites = group.sites.filter((site) => {
            const keyword = this.search.toLowerCase();
            return (
              site.name.toLowerCase().includes(keyword) ||
              site.link.toLowerCase().includes(keyword)
            );
          });

          this.searchNum += filteredSites.length;

          return {
            ...group,
            sites: filteredSites
          };
        })
        .filter((group) => group.sites.length > 0);

      if (this.searchNum === 0) {
        Message({
          message: this.$t("tips.notFoundAndRetry"),
          type: "error"
        });
      } else {
        Message({
          message: this.$t("tips.foundSome", [this.searchNum]),
          type: "success"
        });
      }
    },

    resetSearch() {
      this.searchStatus = false;
      this.search = "";
      this.searchNum = 0;
      this.data = JSON.parse(JSON.stringify(this.sourceData));
    }
  }
};
</script>

<style lang="css" scoped>
.header-bar {
  background: #fff;
  /* position: "fixed"; */
  /* width: "100%"; */
}

.search {
  /* margin-left: 10px; */
  width: 300px;
}

@media screen and (max-width: 768px) {
  .search {
    width: auto;
  }

  .search-text {
    margin: 0 3px;
  }
}
</style>
