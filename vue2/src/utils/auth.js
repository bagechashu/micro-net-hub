import Cookies from "js-cookie";

const TokenKey = "jwt-token";

export function getToken() {
  return Cookies.get(TokenKey);
}

export function setToken(token) {
  // console.log(token)
  // Firefox 和Safari 允许cookie 多达4097 个字节, 包括名(name)、值(value)和等号。
  // Opera 允许cookie 多达4096 个字节, 包括: 名(name)、值(value)和等号
  return Cookies.set(TokenKey, token);
}

export function removeToken() {
  return Cookies.remove(TokenKey);
}
