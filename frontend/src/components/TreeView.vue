<script setup>
import { ref, computed, watch } from 'vue'
import { listFiles, downloadFile } from '../api.js'

const props = defineProps({
  tick: { type: Number, default: 0 },
  path: { type: String, default: '' },
  trash: { type: Boolean, default: false },
})

const emit = defineEmits(['delete', 'restore', 'clean'])

const expanded = ref(props.path === '')
const children = ref([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    children.value = await listFiles(props.path, props.trash)
  } finally {
    loading.value = false
  }
}

// 监听 tick，展开的节点在操作后刷新
watch(() => props.tick, () => {
  if (expanded.value) load()
})

async function toggle() {
  expanded.value = !expanded.value
  if (expanded.value) await load()
}

const dirs = computed(() => children.value.filter((f) => f.isDir))
const files = computed(() => children.value.filter((f) => !f.isDir))

function childPath(name) {
  return props.path ? `${props.path}/${name}` : name
}

function onDownload(f) {
  downloadFile(props.path, f.name)
}

function onDelete(f) {
  emit('delete', { dir: props.path, name: f.name })
}

function onRestore(f) {
  emit('restore', { dir: props.path, name: f.name })
}

function onClean(f) {
  emit('clean', { dir: props.path, name: f.name })
}

// 目录操作 —— 直接传当前路径作为 dir
function onDeleteDir() {
  emit('delete', { dir: props.path, name: '' })
}

function onRestoreDir() {
  emit('restore', { dir: props.path, name: '' })
}

function onCleanDir() {
  emit('clean', { dir: props.path, name: '' })
}

function formatSize(bytes) {
  if (bytes == null) return '-'
  if (bytes >= 1 << 30) return (bytes / (1 << 30)).toFixed(2) + ' GB'
  if (bytes >= 1 << 20) return (bytes / (1 << 20)).toFixed(2) + ' MB'
  if (bytes >= 1 << 10) return (bytes / (1 << 10)).toFixed(2) + ' KB'
  return bytes + ' B'
}

function formatTime(t) {
  if (!t) return '-'
  return new Date(t).toLocaleString()
}

load()
</script>

<template>
  <div class="tree-node">
    <div v-if="path" class="node-header" @click="toggle">
      <span class="arrow">{{ expanded ? '▼' : '▶' }}</span>
      <span class="icon">📁</span>
      <span class="name">{{ path.split('/').pop() }}</span>
      <span class="node-actions" @click.stop>
        <template v-if="!trash">
          <button class="btn btn-rm" @click="onDeleteDir">删除</button>
        </template>
        <template v-else>
          <button class="btn btn-restore" @click="onRestoreDir">恢复</button>
          <button class="btn btn-clean" @click="onCleanDir">彻底删除</button>
        </template>
      </span>
    </div>

    <div v-if="expanded" class="node-children" :class="{ root: !path }">
      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="children.length === 0" class="empty">(空)</div>
      <template v-else>
        <!-- 子目录 -->
        <TreeView
          v-for="d in dirs"
          :key="d.name"
          :tick="tick"
          :path="childPath(d.name)"
          :trash="trash"
          @delete="emit('delete', $event)"
          @restore="emit('restore', $event)"
          @clean="emit('clean', $event)"
        />
        <!-- 文件 -->
        <div v-for="f in files" :key="f.name" class="file-row">
          <span class="icon">📄</span>
          <span class="name">{{ f.name }}</span>
          <span class="size">{{ formatSize(f.size) }}</span>
          <span class="time">{{ formatTime(f.time) }}</span>

          <template v-if="!trash">
            <button class="btn btn-dl" @click="onDownload(f)">下载</button>
            <button class="btn btn-rm" @click="onDelete(f)">删除</button>
          </template>
          <template v-else>
            <button class="btn btn-restore" @click="onRestore(f)">恢复</button>
            <button class="btn btn-clean" @click="onClean(f)">彻底删除</button>
          </template>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.tree-node { user-select: none; }
.node-header {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 4px; cursor: pointer; border-radius: 4px;
}
.node-header:hover { background: #f0f4ff; }
.arrow { font-size: 10px; width: 14px; color: #888; flex-shrink: 0; }
.icon { flex-shrink: 0; }
.name { flex: 1; font-size: 14px; }
.node-actions { opacity: 0; transition: opacity 0.15s; display: flex; gap: 4px; }
.node-header:hover .node-actions { opacity: 1; }
.node-children { padding-left: 20px; }
.node-children.root { padding-left: 0; }
.loading, .empty {
  padding: 4px 4px 4px 24px; color: #999; font-size: 13px;
}
.file-row {
  display: flex; align-items: center; gap: 8px;
  padding: 5px 4px; font-size: 13px; border-radius: 3px;
}
.file-row:hover { background: #f9fafb; }
.size { color: #888; font-size: 12px; white-space: nowrap; min-width: 70px; }
.time { color: #aaa; font-size: 12px; white-space: nowrap; flex: 1; }
.btn {
  border: none; padding: 2px 10px; border-radius: 3px; cursor: pointer; font-size: 12px;
}
.btn-dl { background: #e8f4fd; color: #1976d2; }
.btn-dl:hover { background: #d0ebfb; }
.btn-rm { background: #fde8e8; color: #c62828; }
.btn-rm:hover { background: #fbd0d0; }
.btn-restore { background: #e8f5e9; color: #2e7d32; }
.btn-restore:hover { background: #d0edd1; }
.btn-clean { background: #fff3e0; color: #e65100; }
.btn-clean:hover { background: #ffe0b2; }
</style>
