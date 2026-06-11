import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/upload': 'http://localhost:8666',
      '/download': 'http://localhost:8666',
      '/listfiles': 'http://localhost:8666',
      '/delete': 'http://localhost:8666',
      '/restore': 'http://localhost:8666',
      '/clean': 'http://localhost:8666',
    },
  },
})
