const BASE = ''

async function request(url, options = {}) {
  const res = await fetch(BASE + url, options)
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `请求失败 (${res.status})`)
  }
  return res
}

export function uploadFile(file) {
  const form = new FormData()
  form.append('file', file)
  return request('/upload', { method: 'POST', body: form })
}

export function downloadFile(dir, name) {
  const url = `/download?dir=${encodeURIComponent(dir)}&name=${encodeURIComponent(name)}`
  const a = document.createElement('a')
  a.href = url
  a.download = name
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

export async function listFiles(dir, trash = false) {
  let url = `/listfiles?dir=${encodeURIComponent(dir)}`
  if (trash) url += '&trash=1'
  const res = await request(url)
  return res.json()
}

export function deleteItem(dir, name) {
  let url = `/delete?dir=${encodeURIComponent(dir)}`
  if (name) url += `&name=${encodeURIComponent(name)}`
  return request(url, { method: 'DELETE' })
}

export function restoreItem(dir, name) {
  let url = `/restore?dir=${encodeURIComponent(dir)}`
  if (name) url += `&name=${encodeURIComponent(name)}`
  return request(url, { method: 'POST' })
}

export function cleanItem(dir, name) {
  let url = `/clean?dir=${encodeURIComponent(dir)}`
  if (name) url += `&name=${encodeURIComponent(name)}`
  return request(url, { method: 'DELETE' })
}
