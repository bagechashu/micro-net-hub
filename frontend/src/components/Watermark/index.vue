<template>
  <div ref="watermarkContainer" class="watermark-container">
    <slot />
    <canvas ref="watermarkCanvas" class="watermark-canvas" />
  </div>
</template>

<script>
export default {
  name: "Watermark",
  props: {
    text: {
      type: String,
      required: true
    },
    fontSize: {
      type: Number,
      default: 16
    },
    color: {
      type: String,
      default: "rgba(0, 0, 0, 0.15)"
    },
    rotate: {
      type: Number,
      default: -30
    }
  },
  watch: {
    text: "addWatermark",
    fontSize: "addWatermark",
    color: "addWatermark",
    rotate: "addWatermark"
  },
  mounted() {
    this.addWatermark();
    window.addEventListener("resize", this.addWatermark);
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.addWatermark);
  },
  methods: {
    addWatermark() {
      const container = this.$refs.watermarkContainer;
      const canvas = this.$refs.watermarkCanvas;
      const ctx = canvas.getContext("2d");

      const width = container.offsetWidth;
      const height = container.offsetHeight;

      canvas.width = width;
      canvas.height = height;

      ctx.clearRect(0, 0, width, height);
      ctx.save();

      // 旋转画布
      ctx.translate(width / 2, height / 2);
      ctx.rotate((Math.PI / 180) * this.rotate);
      ctx.translate(-width / 2, -height / 2);

      ctx.font = `${this.fontSize}px Arial`;
      ctx.fillStyle = this.color;
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";

      const textWidth = ctx.measureText(this.text).width;
      const textHeight = this.fontSize;
      const diagonal = Math.sqrt(width * width + height * height);
      const stepX = textWidth + 50; // 调整水印间距
      const stepY = textHeight + 50; // 调整水印间距

      for (let x = -diagonal / 2; x < diagonal; x += stepX) {
        for (let y = -diagonal / 2; y < diagonal; y += stepY) {
          ctx.fillText(this.text, x, y);
        }
      }

      ctx.restore();
    }
  }
};
</script>

<style scoped>
.watermark-container {
  position: relative;
}
.watermark-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 30000; /* 确保 z-index 更高, sidebar 的 z-index: 20000 */
}
</style>
