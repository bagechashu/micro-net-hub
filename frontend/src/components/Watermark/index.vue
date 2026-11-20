<template>
  <div ref="containerRef" class="watermark-container">
    <slot />
    <canvas ref="canvasRef" class="watermark-canvas" />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';

defineOptions({ name: 'Watermark' });

const props = withDefaults(defineProps<{
  text: string;
  fontSize?: number;
  color?: string;
  rotate?: number;
}>(), {
  fontSize: 16,
  color: 'rgba(0, 0, 0, 0.15)',
  rotate: -30
});

const containerRef = ref<HTMLDivElement>();
const canvasRef = ref<HTMLCanvasElement>();

function addWatermark() {
  const container = containerRef.value;
  const canvas = canvasRef.value;
  const ctx = canvas?.getContext('2d');
  if (!container || !canvas || !ctx) return;

  const width = container.offsetWidth;
  const height = container.offsetHeight;
  canvas.width = width;
  canvas.height = height;
  ctx.clearRect(0, 0, width, height);
  if (!props.text) return;

  ctx.save();
  // 旋转画布
  ctx.translate(width / 2, height / 2);
  ctx.rotate((Math.PI / 180) * props.rotate);
  ctx.translate(-width / 2, -height / 2);

  ctx.font = `${props.fontSize}px Arial`;
  ctx.fillStyle = props.color;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';

  const stepX = ctx.measureText(props.text).width + 50; // 水印间距
  const stepY = props.fontSize + 50;
  const diagonal = Math.sqrt(width * width + height * height);
  for (let x = -diagonal / 2; x < diagonal; x += stepX) {
    for (let y = -diagonal / 2; y < diagonal; y += stepY) {
      ctx.fillText(props.text, x, y);
    }
  }
  ctx.restore();
}

// 容器高度会随页面内容变化，用 ResizeObserver 重新绘制
let observer: ResizeObserver | undefined;

onMounted(() => {
  addWatermark();
  observer = new ResizeObserver(addWatermark);
  if (containerRef.value) observer.observe(containerRef.value);
});

onBeforeUnmount(() => observer?.disconnect());

watch(() => [props.text, props.fontSize, props.color, props.rotate], addWatermark);
</script>

<style scoped>
.watermark-container {
  position: relative;
  width: 100%;
  height: 100%;
}

.watermark-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 30000; /* 高于侧边栏的 z-index: 20000 */
}
</style>
