import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3010,
    proxy: {
      '/api': { target: 'http://localhost:8090', changeOrigin: true },
      '/webhook': { target: 'http://localhost:8090', changeOrigin: true },
      '/n8n': { target: 'http://localhost:5690', changeOrigin: true, rewrite: path => path.replace(/^\/n8n/, '') },
    },
  },
})
