import Vue from "vue";
import Cookies from "js-cookie";
import "normalize.css/normalize.css"; // a modern alternative to CSS resets
import Element from "element-ui";
import i18n from "@/i18n";
import "./styles/element-variables.scss";
// import enLang from 'element-ui/lib/locale/lang/en'// 如果使用中文语言包请默认支持，无需额外引入，请删除该依赖
import "@/styles/index.scss"; // global css

import VueLazyload from "vue-lazyload";
import VueClipboard from "vue-clipboard2";

import App from "./App";
import store from "./store";
import router from "./router";

import "./icons"; // icon
import "./permission"; // permission control
import "./utils/error-log"; // error log

import * as filters from "./filters"; // global filters

Vue.use(Element, {
  size: Cookies.get("size") || "medium", // set element-ui default size
  locale: Cookies.get("locale") || "en",
  i18n: (key, value) => i18n.t(key, value)
});

Vue.use(VueLazyload, {
  loading: require("@/assets/icon/loading.gif")
});
VueClipboard.config.autoSetContainer = true;
Vue.use(VueClipboard);

// register global utility filters
Object.keys(filters).forEach(key => {
  Vue.filter(key, filters[key]);
});

Vue.config.productionTip = false;

new Vue({
  el: "#app",
  i18n,
  router,
  store,
  render: h => h(App)
});
