import { createApp } from 'vue';

import App from './App.vue';
import router from './router';
import { setupStore } from './store';
import i18n from '@/i18n';
import useThemeStore from '@/store/modules/theme';

// ElMessage / ElMessageBox 是在 JS 里直接调用的，按需引入插件不会自动带上它们的样式
import 'element-plus/es/components/message/style/css';
import 'element-plus/es/components/message-box/style/css';
import '@/styles/index.scss';
import SvgIcon from './icons'; // icon
import './permission'; // permission control
import vPermission from './directive/permission/index'; // permission control

const app = createApp(App);
setupStore(app);
useThemeStore().init();
app.use(i18n);
app.use(router);
app.component('svg-icon', SvgIcon);
app.directive('permission', vPermission);

app.mount('#app');
