import axios from "axios";
import { MessageBox, Message } from "element-ui";
import store from "@/store";
import { getToken } from "@/utils/auth";
import router from "@/router";

// create an axios instance
const service = axios.create({
  baseURL: process.env.NODE_ENV === "production" ? process.env.API_BASE_PATH : "/" // api 的 base_url
  // withCredentials: true, // send cookies when cross-domain requests
  // timeout: 5000 // request timeout
});

// request interceptor
service.interceptors.request.use(
  config => {
    // do something before request is sent
    if (store.getters.token) {
      // let each request carry token
      // ['X-Token'] is a custom headers key
      // please modify it according to the actual situation
      config.headers["Authorization"] = "Bearer " + getToken();
      // config.headers['Content-Type'] = 'application/json'
    }
    return config;
  },
  error => {
    // do something with request error
    return Promise.reject(error);
  }
);

// response interceptor
service.interceptors.response.use(
  /**
   * If you want to get http information such as headers or status
   * Please return  response => response
  */

  /**
   * Determine the request status by custom code
   * Here is just an example
   * You can also judge the status by HTTP Status Code
   */
  response => {
    const res = response.data;
    if (res.code === 200 || res.code === 0) {
      return res;
    } else {
      Message({
        message: res.msg,
        type: "error",
        duration: 5 * 1000
      });
      return false;
    }
  },
  error => {
    if (error.response.status === 401) {
      if (error.response.data.message.includes("Token is expired")) {
        MessageBox.alert(
          "登录状态过期, 请重新登录.",
          "登录状态过期",
          {
            confirmButtonText: " 确定",
            type: "warning",
            callback: action => {
              store.dispatch("user/logout").then(() => {
                router.push("/");
              });
            }
          }
        );
      } else {
        Message({
          showClose: true,
          message: error.response.msg || error.response.data.message || error.message,
          type: "error",
          duration: 5 * 1000
        });
        return Promise.reject(error);
      }
    } else if (error.response.status === 403) {
      router.push({ path: "/401" });
    } else {
      Message({
        showClose: true,
        message: error.response.msg || error.response.data.message || error.message,
        type: "error",
        duration: 5 * 1000
      });
      return Promise.reject(error);
    }
  }
);

export default service;
