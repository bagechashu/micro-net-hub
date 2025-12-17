import { defineStore } from 'pinia';

interface IErrorLogState {
  logs: unknown[];
}

const useErrorLogStore = defineStore('errorLog', {
  state: ():IErrorLogState => ({
    logs: []
  }),
  getters: {},
  actions: {
    addErrorLog(log) {
      this.logs.push(log);
    },
    clearErrorLog() {
      this.logs.splice(0);
    }
  }
});

export default useErrorLogStore;
