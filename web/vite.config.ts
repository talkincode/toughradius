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
        manualChunks: {
          'react-vendor': ['react', 'react-dom'],
          'react-admin': ['react-admin', 'ra-data-simple-rest'],
          'echarts': ['echarts', 'echarts-for-react']
        }
      }
    }
  }
})
