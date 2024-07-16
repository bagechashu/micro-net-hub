<template>
  <div id="app">
    <watermark :text="name" :font-size="16" color="rgba(0, 0, 0, 0.025)" :rotate="-30">
      <router-view />
    </watermark>
  </div>
</template>

<script>
import Watermark from "@/components/Watermark";
import { mapState } from "vuex";

export default {
  name: "App",
  components: {
    Watermark
  },
  computed: {
    ...mapState({
      name: state => state.user.name
    })
  }
};
</script>
