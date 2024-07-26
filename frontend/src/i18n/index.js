import enEle from "element-ui/lib/locale/lang/en";
import zhEle from "element-ui/lib/locale/lang/zh-CN";
import enCustom from "./locales/en.json";
import zhCustom from "./locales/zh.json";

export const messages = {
  en: {
    ...enEle,
    custom: enCustom
  },
  zh: {
    ...zhEle,
    custom: zhCustom
  }
};
