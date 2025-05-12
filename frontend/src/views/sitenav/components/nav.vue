<template>
  <div>
    <el-row :gutter="20">
      <el-col
        v-for="(item, index) in sites"
        :key="index"
        :xs="12"
        :sm="8"
        :md="6"
        :lg="4"
        :xl="4"
        class="site-card"
      >
        <div @click="jumpLink(item.link)">
          <el-popover placement="top" width="200" trigger="hover" :open-delay="500" :close-delay="100">
            <el-button-group class="vertical">
              <el-button
                size="small"
                icon="el-icon-position"
                @click="jumpLink(item.link)"
              >{{ $t("sitenav.visit") }}</el-button>
              <el-button
                v-if="item.doc"
                size="small"
                icon="el-icon-document"
                @click="jumpLink(item.doc)"
              >{{ $t("sitenav.doc") }}</el-button>
              <el-button
                size="small"
                icon="el-icon-copy-document"
                class="clip-btn"
                @click="copyLink(item.link)"
              >
                {{ $t("sitenav.copyURL") }}</el-button>
              <el-button
                size="small"
                icon="el-icon-star-on"
                @click="addBookmarks(item.link, item.name)"
              >{{ $t("sitenav.addBookmark") }}</el-button>
            </el-button-group>
            <el-card slot="reference" shadow="hover" :body-style="{ padding: '8px' }">
              <span><img v-lazy="item.icon" class="icon" alt=""></span>
              <span>{{ item.name }}</span>
              <p />
              <div class="desc">{{ item.desc }}</div>
              <div class="desclink">{{ item.link }}</div>
            </el-card>
          </el-popover>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { Message } from "element-ui";

export default {
  components: {},
  props: {
    sites: {
      type: Array,
      default: () => []
    }
  },
  methods: {
    jumpLink(link) {
      window.open(link);
    },
    copyLink(link) {
      this.$copyText(link)
        .then(() => {
          Message({
            message: this.$t("tips.copySuccess"),
            type: "success"
          });
        })
        .catch((error) => {
          Message({
            message: this.$t("tips.copyFailed", [error]),
            type: "error"
          });
        });
    },
    addBookmarks(url, title) {
      Message({
        message: this.$t("tips.addBookmarkInfo"),
        type: "info"
      });
    }
  }
};
</script>

<style lang="css" scoped>
.site-card {
  margin: 5px;
  /* height: 8rem; */
}
.el-button-group.vertical {
  display: flex;
  flex-direction: column;
}

.el-button-group.vertical .el-button {
  width: 100%;
  /* margin-bottom: 8px; */
}

.icon {
  width: 30px;
  height: 30px;
  vertical-align: middle;
  border-radius: 50%;
  pointer-events: none;
  margin-right: 1rem;
}

@media screen and (max-width: 768px) {
  .nav-li {
    width: 100%;
  }
}

.desc {
  padding-top: 5px;
  border-top: 1px solid #eee;
  margin-top: 5px;
  font-size: 0.6rem;
  color: rgba(0, 0, 0, 0.45);
}
.desclink {
  margin-top: 5px;
  font-size: 0.6rem;
  width: 12rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: rgba(0, 0, 0, 0.45);
}
</style>
