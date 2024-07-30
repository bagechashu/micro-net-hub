/**
 * Created by PanJiaChen on 16/11/18.
 */
import i18n from "@/i18n";

/**
 * @param {string} path
 * @returns {Boolean}
 */
export function isExternal(path) {
  return /^(https?:|mailto:|tel:)/.test(path);
}

/**
 * @param {string} str
 * @returns {Boolean}
 */
export function validUsername(str) {
  return str.length >= 2;
}

/**
 * @param {string} url
 * @returns {Boolean}
 */
export function validURL(url) {
  const reg = /^(https?|ftp):\/\/([a-zA-Z0-9.-]+(:[a-zA-Z0-9.&%$-]+)*@)*((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])){3}|([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.(com|edu|gov|int|mil|net|org|biz|arpa|info|name|pro|aero|coop|museum|[a-zA-Z]{2}))(:[0-9]+)*(\/($|[a-zA-Z0-9.,?'\\+&%$#=~_-]+))*$/;
  return reg.test(url);
}

/**
 * @param {string} str
 * @returns {Boolean}
 */
export function validLowerCase(str) {
  const reg = /^[a-z]+$/;
  return reg.test(str);
}

/**
 * @param {string} str
 * @returns {Boolean}
 */
export function validUpperCase(str) {
  const reg = /^[A-Z]+$/;
  return reg.test(str);
}

/**
 * @param {string} str
 * @returns {Boolean}
 */
export function validAlphabets(str) {
  const reg = /^[A-Za-z]+$/;
  return reg.test(str);
}

/**
 * @param {string} email
 * @returns {Boolean}
 */
export function validEmail(email) {
  const reg = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
  return reg.test(email);
}

/**
 * @param {string} str
 * @returns {Boolean}
 */
export function isString(str) {
  if (typeof str === "string" || str instanceof String) {
    return true;
  }
  return false;
}

/**
 * @param {Array} arg
 * @returns {Boolean}
 */
export function isArray(arg) {
  if (typeof Array.isArray === "undefined") {
    return Object.prototype.toString.call(arg) === "[object Array]";
  }
  return Array.isArray(arg);
}

function checkPasswordComplexity(value) {
  let matches = 0;

  if (value.match(/[a-z]/)) matches++;
  if (value.match(/[A-Z]/)) matches++;
  if (value.match(/\d/)) matches++;
  if (value.match(/[!@#$%^&*,.?\_\-]/)) matches++;

  return matches;
}

export function validatePassword(rule, value, callback) {
  if (value === "") {
    return callback(new Error(i18n.t("loginform.newpassTips")));
  }
  if (value.length < 8) {
    return callback(new Error(i18n.t("loginform.newpassValidErrLenMustThan8")));
  }
  if (checkPasswordComplexity(value) < 3) {
    return callback(new Error(i18n.t("loginform.newpassValidErrMustComplex")));
  }
  callback();
}

export function validatePasswordCanEnpty(rule, value, callback) {
  if (value === "") {
    return callback();
  }
  if (value.length < 8) {
    return callback(new Error(i18n.t("loginform.newpassValidErrLenMustThan8")));
  }
  if (checkPasswordComplexity(value) < 3) {
    return callback(new Error(i18n.t("loginform.newpassValidErrMustComplex")));
  }
  callback();
}

export function validateName(rule, value, callback) {
  if (value === "") {
    return callback(new Error(i18n.t("valid.notAllowEmpty")));
  }

  const regex = /^[a-zA-Z0-9_-]+$/;
  const isValid = regex.test(value);

  if (!isValid) {
    return callback(new Error(i18n.t("valid.InvalidName")));
  }

  // if callback(null);
  // vue.runtime.esm.js:620 [Vue warn]: Error in v-on handler: "TypeError: Cannot read properties of null (reading 'field')"
  callback();
}
