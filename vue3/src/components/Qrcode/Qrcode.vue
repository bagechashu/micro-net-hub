<template>
  <div :id="id" style="height: 100%;">
    <canvas v-show="text" ref="canvasRef" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import QRCode from 'qrcode';

const props = defineProps({
  id: {
    type: String,
    required: true
  },
  // 后端返回的二维码地址
  text: {
    type: String,
    default: ''
  },
  width: {
    type: [String, Number],
    default: '128'
  },
  colorDark: {
    type: String,
    default: '#000000'
  },
  colorLight: {
    type: String,
    default: '#ffffff'
  }
});

const canvasRef = ref<HTMLCanvasElement>();

async function createQrcode() {
  if (!canvasRef.value || !props.text) return;
  try {
    await QRCode.toCanvas(canvasRef.value, props.text, {
      width: Number(props.width),
      margin: 0,
      errorCorrectionLevel: 'H',
      color: { dark: props.colorDark, light: props.colorLight }
    });
  } catch (err) {
    console.error('QRCode render failed', err);
  }
}

watch(() => [props.text, props.width, props.colorDark, props.colorLight], createQrcode);
onMounted(createQrcode);
</script>
