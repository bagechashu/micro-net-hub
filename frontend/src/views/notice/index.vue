<template>
  <div class="message-container">
    <marquee-text
      :repeat="repeat"
      :duration="duration"
      :paused="paused"
      @mouseover="paused = true"
      @mouseleave="paused = false"
    >
      <span v-for="item in data" :key="item.ID">
        <span
          class="message"
          :style="{
            fontSize: getRandomFontSize(),
            fontWeight: getRandomWeight(),
          }"
        >
          <span
            :class="['level', getLevelColorClass(item.level)]"
          >{{ getLevelText(item.level) }}:</span>
          <span class="content">
            {{ item.content }}
          </span></span>
      </span>
    </marquee-text>
  </div>
</template>

<script>
import { getNotice } from "@/api/notice/notice";
import MarqueeText from "vue-marquee-text-component";
export default {
  name: "NoticeBoard",
  components: {
    MarqueeText
  },
  data() {
    return {
      duration: 15,
      paused: false,
      repeat: 4,
      data: []
    };
  },
  created: function() {
    this._getNotice();
  },
  mounted() {
    this.timer = setInterval(() => this._getNotice(), 60 * 60 * 1000); // 每 60 分钟执行一次
  },
  beforeDestroy() { // vue3 中 beforeDestroy 改名为 beforeUnmount.
    if (this.timer) {
      clearInterval(this.timer); // 组件销毁前清除定时器
    }
  },
  methods: {
    async _getNotice() {
      try {
        const { data } = await getNotice();
        this.data = data;
        this.duration = Math.round(JSON.stringify(this.data).length / 30);
        this.repeat = 10 - data.length;
      } catch (e) {
        console.log(e);
      }
    },
    getLevelText(level) {
      switch (level) {
        case 1:
          return "一般";
        case 2:
          return "普通";
        case 3:
          return "重要";
        case 4:
          return "紧急";
        default:
          return "其他";
      }
    },
    getLevelColorClass(level) {
      switch (level) {
        case 1:
          return "default-level";
        case 2:
          return "info-level";
        case 3:
          return "warning-level";
        case 4:
          return "critical-level";
        default:
          return "default-level";
      }
    },
    getRandomFontSize() {
      const minSize = 12; // minimum font size in pixels
      const maxSize = 20; // maximum font size in pixels
      return `${
        Math.floor(Math.random() * (maxSize - minSize + 1)) + minSize
      }px`;
    },
    getRandomWeight() {
      return Math.random() > 0.5 ? "bold" : "normal";
    }
  }
};
</script>

<style lang="css" scoped>
.message-container {
  display: flex;
  flex-direction: column;
  height: fit-content;
  padding: 8px;
  /* background-color: #fefefe;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05); */
}

.message {
  align-items: center;
  padding: 2px 10px;
  /* border-radius: 8px; */
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.info-level {
  background-color: #66bb6a;
  color: white;
}

.warning-level {
  background-color: #ffa726;
  color: white;
}

.critical-level {
  background-color: #e53935;
  color: white;
}

.default-level {
  background-color: #999999;
  color: white;
}

.level {
  margin-left: 5px;
  padding: 1px 8px;
  border-radius: 4px;
  font-weight: bold;
}

.content {
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
