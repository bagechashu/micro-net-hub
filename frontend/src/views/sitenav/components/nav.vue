<template>
  <div>
    <el-row :gutter="20">
      <el-col
        v-for="(item, index) in navData"
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
              @click="jumpLink(item)"
            >跳转</el-button>
            <!-- FIXME: 导航页 button 3 个按钮功能异常 -->
            <el-button
              v-if="item.doc"
              size="small"
              icon="el-icon-document"
              @click="openDoc(item)"
            >使用文档</el-button>
            <el-button
              size="small"
              icon="el-icon-copy-document"
              class="btn"
              :data-clipboard-text="item.link"
              @click="copyLink"
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
            <span class="desc">{{ item.desc }}</span>
          </el-card>
        </el-popover>
      </el-col>
    </el-row>

    <el-dialog
      v-model="modalDoc"
      fullscreen
      title="使用文档"
      @on-cancel="closeDoc"
    >
      <div v-if="modalDoc" class="usage-content">
        <div class="toc">
          目录
          <div id="toc" />
        </div>
        <div class="markdown">
          <vue-markdown
            :source="docData"
            :toc="true"
            toc-id="toc"
          />
        </div>
        <Spin v-if="docSpinShow" size="large" fix />
      </div>
    </el-dialog>
    <Spin v-if="spinShow" size="large" fix />
  </div>
</template>

<script>
import Clipboard from "clipboard";
import VueMarkdown from "vue-markdown";

import hljs from "highlight.js";
// import "highlight.js/styles/atom-one-dark.css";
import "highlight.js/styles/github.css";

const highlightCode = () => {
  const preEl = document.querySelectorAll("pre");
  const codeEl = document.querySelectorAll("code");
  preEl.forEach((el) => {
    hljs.highlightBlock(el);
  });
  codeEl.forEach((el) => {
    hljs.highlightBlock(el);
  });
};

export default {
  components: {
    VueMarkdown
  },
  props: {
    navData: {
      type: Array,
      default: () => []
    },
    subTitle: {
      type: String,
      default: ""
    },
    spinShow: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      modalDoc: false,
      docSpinShow: false,
      docData: ""
    };
  },
  mounted() {
    highlightCode();
    // console.log(this.navData);
  },
  updated() {
    highlightCode();
  },
  methods: {
    openDoc(item) {
      if (item.doc.startsWith("http")) {
        window.open(item.doc);
        return;
      }
      this.modalDoc = true;
      this.docSpinShow = true;
      this.$axios
        .get(item.doc)
        .then((rep) => {
          this.docData = rep.data;
        })
        .catch((e) => {
          this.$Message.error("获取数据失败!");
          window.console.log(e);
        })
        .then(() => {
          this.docSpinShow = false;
        });
    },
    closeDoc() {
      this.docData = "";
      this.modalDoc = false;
    },
    jumpLink(item) {
      item.title = this.subTitle ? this.subTitle : item.title;
      window.open(item.link);
    },
    copyLink() {
      var clipboard = new Clipboard(".btn");
      clipboard.on("success", (e) => {
        // 成功提示
        this.$Message.success("复制成功");
        // 释放内存
        clipboard.destroy();
        window.console.log(e);
      });
      clipboard.on("error", (e) => {
        // 不支持复制
        this.$Message.error("该浏览器不支持自动复制");
        // 释放内存
        clipboard.destroy();
        window.console.log(e);
      });
    }
  }
};
</script>

<style lang="css" scoped>
.site-card {
  /* margin: 0, 1rem, 1rem, 0; */
  height: 7rem;
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

.top {
  height: 36px;
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
  margin-top: 8px;
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
}

.toc {
  width: 200px;
  position: fixed;
}

@media screen and (max-width: 768px) {
  .toc {
    position: relative;
    margin-left: 0px;
  }
}

.toc a {
  word-break: break-all;
  word-wrap: break-word;
}

.markdown {
  margin-left: 210px;
}

@media screen and (max-width: 768px) {
  .markdown {
    position: relative;
    margin-left: 0px;
  }
}
</style>
