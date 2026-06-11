<script setup>
import { ref, computed } from 'vue'
import TreeView from './components/TreeView.vue'
import UploadButton from './components/UploadButton.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import { uploadFile, deleteItem, restoreItem, cleanItem } from './api.js'

const showTrash = ref(false)
const error = ref('')
const refreshKey = ref(0)

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
    refreshKey.value++
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
      refreshKey.value++
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
      refreshKey.value++
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
      refreshKey.value++
    } catch (e) {
      error.value = '彻底删除失败: ' + e.message
    }
  })
}

function toggleTrash() {
  showTrash.value = !showTrash.value
  refreshKey.value++
}

const title = computed(() => (showTrash.value ? '🗑️ 回收站' : '📁 文件管理'))
</script>

<template>
  <div class="app">
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
        :key="refreshKey"
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
</style>
