<script setup>
defineProps({
  visible: Boolean,
  message: String,
})

const emit = defineEmits(['confirm', 'cancel'])
</script>

<template>
  <div v-if="visible" class="overlay" @click.self="emit('cancel')">
    <div class="dialog">
      <p class="msg">{{ message }}</p>
      <div class="actions">
        <button class="btn btn-cancel" @click="emit('cancel')">取消</button>
        <button class="btn btn-confirm" @click="emit('confirm')">确认</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.dialog {
  background: #fff;
  border-radius: 8px;
  padding: 28px 32px;
  min-width: 320px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}
.msg {
  margin: 0 0 20px;
  font-size: 15px;
  line-height: 1.6;
}
.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
.btn {
  padding: 7px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}
.btn-cancel {
  background: #eee;
  color: #333;
}
.btn-confirm {
  background: #c62828;
  color: #fff;
}
</style>
