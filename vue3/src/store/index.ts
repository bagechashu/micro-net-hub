// 经过这次修改, 已经被架空的, 保留以备不时之需
import { createPinia, acceptHMRUpdate } from 'pinia';
import type { StoreDefinition } from 'pinia';
// https://webpack.js.org/guides/dependency-management/#requirecontext
const modulesFiles = import.meta.glob('./modules/*.ts', { eager: true });
// console.log('modulesFiles=', modulesFiles);

// you do not need `import app from './modules/app'`
// it will auto require all vuex module from modules file
const modules:Record<string, StoreDefinition> = {};
Object.keys(modulesFiles).forEach((modulePath) => {
  // console.log('modulePath=', modulePath);
  // set './modules/app.js' => 'app'
  const moduleName = modulePath.replace(/^\.\/modules\/(.*)\.\w+$/, '$1');
  const value = modulesFiles[modulePath] as any;
  modules[moduleName] = value.default;
}, {});

export const setupStore = (app) => {
  app.use(createPinia());
  Object.values(modules).forEach(fn => {
    fn();
    // console.log('import.meta.hot=', import.meta.hot);
    if (import.meta.hot) {
      import.meta.hot.accept(acceptHMRUpdate(fn, import.meta.hot));
    }
  });
};

// console.log('modules=', modules);

export default modules;
