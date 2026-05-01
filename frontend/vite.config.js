import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      // Backend Go (API + webhooks)
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
      '/webhook': { target: 'http://localhost:8080', changeOrigin: true },
      // Proxy legado para pruebas manuales contra n8n; Actas usa el backend.
      '/n8n': {
        target: 'http://localhost:5678',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/n8n/, ''),
      },
    },
  },
})
