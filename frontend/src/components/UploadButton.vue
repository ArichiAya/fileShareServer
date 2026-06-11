<script setup>
import { ref } from 'vue'

const emit = defineEmits(['upload'])

const uploading = ref(false)
const fileInput = ref(null)

function trigger() {
  fileInput.value.click()
}

defineExpose({ trigger })

async function handleChange(e) {
  const file = e.target.files[0]
  if (!file) return
  uploading.value = true
  try {
    await emit('upload', file)
  } finally {
    uploading.value = false
    e.target.value = ''
  }
}
</script>

<template>
  <div class="upload-btn">
    <button class="btn" :disabled="uploading" @click="trigger">
      {{ uploading ? '上传中...' : '📤 上传文件' }}
    </button>
    <input
      ref="fileInput"
      type="file"
      class="hidden-input"
      @change="handleChange"
    />
  </div>
</template>

<style scoped>
.btn {
  padding: 7px 18px;
  background: #42b883;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}
.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.hidden-input {
  display: none;
}
</style>
