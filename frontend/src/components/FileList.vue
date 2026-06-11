<script setup>
import { computed } from 'vue'

const props = defineProps({
  files: Array,
  loading: Boolean,
  trashMode: Boolean,
})

const emit = defineEmits(['download', 'delete', 'restore', 'clean'])

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

const columns = computed(() => {
  if (props.trashMode) {
    return ['名称', '大小', '时间', '恢复', '彻底删除']
  }
  return ['名称', '大小', '时间', '下载', '删除']
})
</script>

<template>
  <div class="file-list">
    <div v-if="loading" class="status">加载中...</div>
    <div v-else-if="!files || files.length === 0" class="status">暂无文件</div>
    <table v-else>
      <thead>
        <tr>
          <th
            v-for="col in columns"
            :key="col"
            :class="{ 'col-action': ['下载', '删除', '恢复', '彻底删除'].includes(col) }"
          >
            {{ col }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="f in files" :key="f.name">
          <td>
            <span v-if="f.isDir" class="icon">📁</span>
            <span v-else class="icon">📄</span>
            {{ f.name }}
          </td>
          <td class="col-size">{{ f.isDir ? '-' : formatSize(f.size) }}</td>
          <td class="col-time">{{ formatTime(f.time) }}</td>

          <!-- 普通模式 -->
          <template v-if="!trashMode">
            <td class="col-action">
              <button
                v-if="!f.isDir"
                class="btn btn-sm btn-download"
                @click="emit('download', f)"
              >
                下载
              </button>
            </td>
            <td class="col-action">
              <button
                class="btn btn-sm btn-delete"
                @click="emit('delete', f)"
              >
                删除
              </button>
            </td>
          </template>

          <!-- 回收站模式 -->
          <template v-else>
            <td class="col-action">
              <button class="btn btn-sm btn-restore" @click="emit('restore', f)">
                恢复
              </button>
            </td>
            <td class="col-action">
              <button class="btn btn-sm btn-clean" @click="emit('clean', f)">
                彻底删除
              </button>
            </td>
          </template>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.file-list {
  width: 100%;
}
.status {
  text-align: center;
  color: #999;
  padding: 40px;
  font-size: 15px;
}
table {
  width: 100%;
  border-collapse: collapse;
}
th, td {
  padding: 10px 12px;
  border-bottom: 1px solid #eee;
  text-align: left;
}
th {
  font-weight: 600;
  color: #555;
  font-size: 13px;
}
td {
  font-size: 14px;
}
.icon {
  margin-right: 6px;
}
.col-size {
  white-space: nowrap;
  color: #666;
  width: 100px;
}
.col-time {
  white-space: nowrap;
  color: #888;
  font-size: 13px;
}
.col-action {
  width: 80px;
  text-align: center;
}
.btn-sm {
  padding: 4px 10px;
  border: none;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
}
.btn-download {
  background: #e8f4fd;
  color: #1976d2;
}
.btn-download:hover {
  background: #d0ebfb;
}
.btn-delete {
  background: #fde8e8;
  color: #c62828;
}
.btn-delete:hover {
  background: #fbd0d0;
}
.btn-restore {
  background: #e8f5e9;
  color: #2e7d32;
}
.btn-restore:hover {
  background: #d0edd1;
}
.btn-clean {
  background: #fff3e0;
  color: #e65100;
}
.btn-clean:hover {
  background: #ffe0b2;
}
</style>
