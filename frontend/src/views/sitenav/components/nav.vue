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
        :xl="2"
        class="site-card"
      >
        <el-popover placement="right-start" width="200" trigger="hover">
          <el-button-group class="vertical">
            <el-button
              size="small"
              icon="el-icon-position"
              @click="jumpLink(item.link)"
            >跳转</el-button>
            <el-button
              v-if="item.doc"
              size="small"
              icon="el-icon-document"
              @click="jumpLink(item.doc)"
            >相关文档</el-button>
            <el-button
              size="small"
              icon="el-icon-copy-document"
              class="clip-btn"
              @click="copyLink(item.link)"
            >
              拷贝网址</el-button>
            <el-button
              size="small"
              icon="el-icon-star-on"
              @click="addBookmarks(item.link, item.name)"
            >加入书签</el-button>
          </el-button-group>
          <el-card slot="reference" shadow="never">
            <span><img v-lazy="item.icon" class="icon" alt=""></span>
            <span>{{ item.name }}</span>
            <p />
            <div class="desc">{{ item.desc }}</div>
            <div class="desclink">{{ item.link }}</div>
          </el-card>
        </el-popover>
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
      this.$copyText(link).then(
        function(e) {
          Message({
            message: "复制成功",
            type: "success"
          });
          // console.log(e);
        },
        function(e) {
          Message({
            message: "该浏览器不支持自动复制",
            type: "error"
          });
          // console.log(e);
        }
      );
    },
    addBookmarks(url, title) {
      Message({
        message: "请按 Ctrl+D 或 Command+D 将此页面添加至书签",
        type: "info"
      });
    }
  }
};
</script>

<style lang="css" scoped>
.site-card {
  /* margin: 0, 1rem, 1rem, 0; */
  height: 8rem;
  margin-bottom: 1rem;
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
  width: 40px;
  height: 40px;
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
