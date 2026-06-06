import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  base: '/admin',
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:1816',
        changeOrigin: true,
      }
    }
  },
  build: {
    outDir: 'dist/admin',
    assetsDir: 'assets',
    emptyOutDir: true,
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('/node_modules/')) return
          if (id.includes('/node_modules/react-admin/') || id.includes('/node_modules/ra-data-simple-rest/')) return 'react-admin'
          if (id.includes('/node_modules/echarts/') || id.includes('/node_modules/echarts-for-react/')) return 'echarts'
          if (id.includes('/node_modules/react/') || id.includes('/node_modules/react-dom/')) return 'react-vendor'
        }
      }
    }
  }
})
