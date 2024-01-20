<template>
  <div>
    <div v-for="(item, index) in data" :id="item.title" :key="index">
      <el-card v-if="item.sites" class="node-card" shadow="hover">
        <div slot="header" class="clearfix">
          <span>{{ item.title }}<a :href="'#' + item.title" /> </span>
          <span
            style="float: right; padding: 3px 0"
            type="text"
          >共计：{{ item.sites.length }} 个项目</span>
        </div>
        <Nav
          :sites="item.sites"
        />
      </el-card>
    </div>
  </div>
</template>

<script>
import Nav from "./nav";
export default {
  components: {
    Nav
  },
  props: {
    data: {
      type: Array,
      default: () => []
    }
  }
};
</script>

<style scoped>
.anchor a {
  opacity: 0;
  font-size: 24px;
  transition: opacity 0.2s ease-in-out;
  margin-left: 5px;
  position: relative;
  top: 2px;
}
.node-card {
  margin-top: 1rem;
}
.clearfix:before,
.clearfix:after {
  display: table;
  content: "";
}
.clearfix:after {
  clear: both;
}
</style>
