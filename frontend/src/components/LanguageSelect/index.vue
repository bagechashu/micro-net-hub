<template>
  <el-dropdown trigger="click" @command="handleSetLocale">
    <div>
      <svg-icon class-name="language-icon" icon-class="language" />
    </div>
    <el-dropdown-menu slot="dropdown">
      <el-dropdown-item
        v-for="item of localeOptions"
        :key="item.value"
        :disabled="locale === item.value"
        :command="item.value"
      >
        {{ item.label }}
      </el-dropdown-item>
    </el-dropdown-menu>
  </el-dropdown>
</template>

<script>
export default {
  data() {
    return {
      localeOptions: [
        { label: "中文", value: "zh" },
        { label: "English", value: "en" }
      ]
    };
  },
  computed: {
    locale() {
      return this.$store.getters.locale;
    }
  },
  methods: {
    handleSetLocale(locale) {
      this.$ELEMENT.locale = locale;
      this.$store.dispatch("app/setLocale", locale);
      this.refreshView();
    },
    refreshView() {
      location.reload(true);
    }
  }
};
</script>
