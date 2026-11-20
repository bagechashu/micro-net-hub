import { defineStore } from 'pinia';
import Cookies from 'js-cookie';

interface IAppState {
  size: 'large' | 'default' | 'small';
}

/** 全局组件尺寸（App.vue 中的 el-config-provider 使用） */
const useAppStore = defineStore('app', {
  state: (): IAppState => ({
    size: (Cookies.get('size') as IAppState['size']) || 'default'
  }),
  actions: {
    setSize(size: IAppState['size']) {
      this.size = size;
      Cookies.set('size', size);
    }
  }
});

export default useAppStore;
