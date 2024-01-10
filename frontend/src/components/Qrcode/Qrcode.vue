<template>
  <div :id="id" :ref="id" style="height: 100%;" />
</template>
<script>
// https://blog.csdn.net/weixin_44705979/article/details/132268988
import QRCode from "qrcodejs2";
export default {
  props: {
    id: {
      type: String,
      required: true
    },
    text: { // 后端返回的二维码地址
      type: String,
      default: ""
    },
    width: {
      type: String,
      default: "128"
    },
    height: {
      type: String,
      default: "128"
    },
    colorDark: {
      type: String,
      default: "#000000"
    },
    colorLight: {
      type: String,
      default: "#ffffff"
    }
  },
  data() {
    return {
      qrcode: ""
    };
  },
  watch: {
    text(newText) {
      this.createQrcode();
    }
  },
  mounted() {
    this.createQrcode();
  },
  methods: {
    createQrcode() {
      if (this.qrcode) { // 有新的二维码地址了，先把之前的清除掉
        this.$refs[this.id].innerHTML = "";
      }
      this.qrcode = new QRCode(this.$refs[this.id], {
        text: this.text, // 页面地址 ,如果页面需要参数传递请注意哈希模式#
        width: this.width, // 二维码宽度 （不支持100%）
        height: this.height, // 二维码高度 （不支持100%）
        colorDark: this.colorDark,
        colorLight: this.colorLight,
        correctLevel: QRCode.CorrectLevel.H
      });
    }
    // // 制作另一个二维码
    // updateCode() {
    //   this.qrcode.makeCode(this.text)
    // }
  }
};
</script>
