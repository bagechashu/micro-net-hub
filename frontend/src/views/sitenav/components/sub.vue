<template>
  <div>
    <div v-for="(item, index) in data" :id="item.title" :key="index">
      <el-card v-if="item.sites" class="node-card" shadow="hover" :body-style="{ padding: '5px' }">
        <div slot="header" class="clearfix card-header">
          <span style="float: left; padding: 3px 0">{{ item.title }}<a :href="'#' + item.title" /> </span>
          <span
            style="float: right; padding: 3px 0"
            type="text"
          >{{ $t('sitenav.totolInfo', [item.sites.length]) }}</span>
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
  font-size: 14px;
  transition: opacity 0.2s ease-in-out;
  margin-left: 5px;
  position: relative;
  top: 2px;
}
.card-header {
  font-size: 14px;
}
::v-deep .el-card__header {
  padding: 8px;
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
