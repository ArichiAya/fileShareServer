<script setup>
import { ref, computed } from 'vue'
import TreeView from './components/TreeView.vue'
import UploadButton from './components/UploadButton.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import { uploadFile, deleteItem, restoreItem, cleanItem } from './api.js'

const showTrash = ref(false)
const error = ref('')
const tick = ref(0)  // 操作后 +1，TreeView watch 来刷新已展开的节点

// 拖拽上传
const dragOver = ref(false)
let dragCount = 0

function onDragEnter(e) {
  e.preventDefault()
  dragCount++
  if (!showTrash.value) dragOver.value = true
}
function onDragLeave(e) {
  dragCount--
  if (dragCount <= 0) {
    dragCount = 0
    dragOver.value = false
  }
}
function onDragOver(e) {
  e.preventDefault()
}
async function onDrop(e) {
  e.preventDefault()
  dragOver.value = false
  dragCount = 0
  if (showTrash.value) return
  const files = e.dataTransfer?.files
  if (!files || files.length === 0) return
  for (const file of files) {
    await onUpload(file)
  }
}

// 确认对话框
const dialogVisible = ref(false)
const dialogMessage = ref('')
let pendingAction = null

function showDialog(msg, action) {
  dialogMessage.value = msg
  pendingAction = action
  dialogVisible.value = true
}

function closeDialog() {
  dialogVisible.value = false
  pendingAction = null
}

async function doConfirm() {
  if (pendingAction) await pendingAction()
  closeDialog()
}

// 上传
async function onUpload(file) {
  error.value = ''
  try {
    await uploadFile(file)
    tick.value++
  } catch (e) {
    error.value = '上传失败: ' + e.message
  }
}

const targetLabel = ({ dir, name }) => name || dir.split('/').pop()

// 删除（移至回收站）
function onDelete(payload) {
  const label = targetLabel(payload)
  showDialog(`确认将「${label}」移至回收站？`, async () => {
    try {
      await deleteItem(payload.dir, payload.name)
      tick.value++
    } catch (e) {
      error.value = '删除失败: ' + e.message
    }
  })
}

// 恢复
function onRestore(payload) {
  const label = targetLabel(payload)
  showDialog(`确认恢复「${label}」？`, async () => {
    try {
      await restoreItem(payload.dir, payload.name)
      tick.value++
    } catch (e) {
      error.value = '恢复失败: ' + e.message
    }
  })
}

// 彻底删除
function onClean(payload) {
  const label = targetLabel(payload)
  showDialog(`确认彻底删除「${label}」？此操作不可撤销！`, async () => {
    try {
      await cleanItem(payload.dir, payload.name)
      tick.value++
    } catch (e) {
      error.value = '彻底删除失败: ' + e.message
    }
  })
}

function toggleTrash() {
  showTrash.value = !showTrash.value
  tick.value++
}

const title = computed(() => (showTrash.value ? '🗑️ 回收站' : '📁 文件管理'))
</script>

<template>
  <div
    class="app"
    @dragenter="onDragEnter"
    @dragleave="onDragLeave"
    @dragover="onDragOver"
    @drop="onDrop"
  >
    <!-- 拖拽上传蒙层 -->
    <Transition name="fade">
      <div v-if="dragOver && !showTrash" class="drop-overlay">
        <div class="drop-box">
          <span class="drop-icon">📂</span>
          <span>释放文件以上传</span>
        </div>
      </div>
    </Transition>

    <header class="header">
      <h1>{{ title }}</h1>
      <div class="header-actions">
        <UploadButton v-if="!showTrash" @upload="onUpload" />
        <button class="trash-toggle" @click="toggleTrash">
          {{ showTrash ? '返回文件' : '查看回收站' }}
        </button>
      </div>
    </header>

    <div v-if="error" class="error">{{ error }}</div>

    <div class="tree-container">
      <TreeView
        :tick="tick"
        path=""
        :trash="showTrash"
        @delete="onDelete"
        @restore="onRestore"
        @clean="onClean"
      />
    </div>

    <ConfirmDialog
      :visible="dialogVisible"
      :message="dialogMessage"
      @confirm="doConfirm"
      @cancel="closeDialog"
    />
  </div>
</template>

<style scoped>
.app {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}
.header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px;
}
.header h1 { margin: 0; font-size: 24px; }
.header-actions { display: flex; align-items: center; gap: 10px; }
.trash-toggle {
  padding: 6px 16px; border: 1px solid #ccc; border-radius: 4px;
  background: #fff; cursor: pointer; font-size: 13px;
}
.trash-toggle:hover { background: #f5f5f5; }
.error {
  background: #fff0f0; color: #c62828;
  padding: 10px 16px; border-radius: 4px; margin-bottom: 12px; font-size: 14px;
}
.tree-container {
  border: 1px solid #eee; border-radius: 8px; padding: 12px 16px;
}

/* 拖拽上传蒙层 */
.drop-overlay {
  position: fixed; inset: 0; z-index: 1000;
  background: rgba(66, 184, 131, 0.12);
  display: flex; align-items: center; justify-content: center;
  pointer-events: none;
}
.drop-box {
  background: #fff;
  border: 3px dashed #42b883;
  border-radius: 16px;
  padding: 48px 64px;
  text-align: center;
  font-size: 20px;
  color: #333;
  display: flex; flex-direction: column; align-items: center; gap: 12px;
}
.drop-icon { font-size: 48px; }

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
