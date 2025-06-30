import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Redirige peticiones de /api a tu backend local
      '/api': {
        target: 'http://localhost:8080', // Cambia si tu backend corre en otro puerto
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
