import Vue from "vue";
import VueI18n from "vue-i18n";
import Cookies from "js-cookie";
import enEle from "element-ui/lib/locale/lang/en";
import zhEle from "element-ui/lib/locale/lang/zh-CN";
import enCustom from "./locales/en.js";
import zhCustom from "./locales/zh.js";

Vue.use(VueI18n);

const messages = {
  en: {
    ...enEle,
    custom: enCustom
  },
  zh: {
    ...zhEle,
    custom: zhCustom
  }
};

const i18n = new VueI18n({
  locale: Cookies.get("locale") || "en", // set locale
  messages // set locale messages
});

export default i18n;
