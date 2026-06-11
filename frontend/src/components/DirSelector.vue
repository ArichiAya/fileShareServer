<script setup>
import { ref } from 'vue'

const props = defineProps({
  modelValue: String,
})

const emit = defineEmits(['update:modelValue', 'refresh'])

const input = ref(props.modelValue || '')

function currentMonth() {
  const now = new Date()
  const y = now.getFullYear()
  const m = String(now.getMonth() + 1).padStart(2, '0')
  return `${y}-${m}`
}

function apply() {
  emit('update:modelValue', input.value.trim())
  emit('refresh')
}

function quickSet(dir) {
  input.value = dir
  emit('update:modelValue', dir)
  emit('refresh')
}
</script>

<template>
  <div class="dir-selector">
    <label class="label">目录:</label>
    <input
      class="input"
      v-model="input"
      placeholder="例如 2026-06"
      @keyup.enter="apply"
    />
    <button class="btn btn-go" @click="apply">进入</button>
    <button class="btn btn-quick" @click="quickSet(currentMonth())">
      本月 ({{ currentMonth() }})
    </button>
  </div>
</template>

<style scoped>
.dir-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.label {
  font-weight: 600;
  white-space: nowrap;
}
.input {
  padding: 6px 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
  width: 140px;
}
.btn {
  padding: 6px 14px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}
.btn-go {
  background: #42b883;
  color: #fff;
}
.btn-quick {
  background: #eee;
  color: #333;
}
.btn-quick:hover {
  background: #ddd;
}
</style>
