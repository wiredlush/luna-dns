import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'https://localhost:8080',
        secure: false,
      },
    },
  },
  build: {
    outDir: '../pkg/web/static',
    emptyOutDir: true,
  },
})
