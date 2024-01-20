<template>
  <div class="app-container">
    <div class="header-bar">
      <el-input
        v-model="search"
        placeholder="内网地址搜索"
        class="search"
        @keyup.enter.native="searchData"
      />
      <span
        class="search-text"
      ><el-button
        type="primary"
        icon="search"
        @click="searchData"
      >搜索</el-button></span>
      <el-button
        v-show="searchStatus"
        type="success"
        icon="plus-round"
        @click="resetSearch"
      >重置</el-button>
    </div>
    <NavSub :data="data" />
  </div>
</template>

<script>
import { Message } from "element-ui";
import NavSub from "@/views/sitenav/components/sub";
import { getNav } from "@/api/sitenav/sitenav";
export default {
  name: "SiteNav",
  components: {
    NavSub
  },
  data() {
    return {
      isCollapsed: false,
      search: "",
      searchStatus: false,
      data: null,
      sourceData: "",
      serarchNum: 0
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
        this.search === ""
      ) {
        Message({
          message: "请输入内容",
          type: "error"
        });
        return true;
      }
      if (!this.searchStatus) {
        this.sourceData = JSON.parse(JSON.stringify(this.data));
      } else {
        this.data = JSON.parse(JSON.stringify(this.sourceData));
      }
      this.searchStatus = true;
      this.serarchNum = 0;
      for (const d in this.data) {
        if (!Object.prototype.hasOwnProperty.call(this.data[d], "nav")) {
          continue;
        }
        for (let i = 0; i < this.data[d]["nav"].length; i++) {
          if (
            this.data[d]["nav"][i]["name"]
              .toLowerCase()
              .indexOf(this.search.toLowerCase()) === -1
          ) {
            if (
              this.data[d]["nav"][i]["link"]
                .toLowerCase()
                .indexOf(this.search.toLowerCase()) === -1
            ) {
              this.data[d]["nav"].splice(i--, 1);
            } else {
              this.serarchNum++;
            }
          } else {
            this.serarchNum++;
          }
        }
      }
      if (this.serarchNum === 0) {
        Message({
          message: "没找到，请重试!",
          type: "error"
        });
      } else {
        Message({
          message: `查找到了 ${this.serarchNum}个相近的.`,
          type: "success"
        });
      }
    },
    resetSearch() {
      this.searchStatus = false;
      this.search = "";
      this.serarchNum = 0;
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
