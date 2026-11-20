import { createI18n } from 'vue-i18n';
import Cookies from 'js-cookie';
import en from './locales/en.json';
import zh from './locales/zh.json';

const messages = {
  en,
  zh
};

const i18n = createI18n({
  legacy: false,
  locale: (Cookies.get('locale') as string) || 'en',
  fallbackLocale: 'en',
  globalInjection: true,
  messages
});

export default i18n;
