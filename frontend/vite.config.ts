import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'

// https://vite.dev/config/
export default defineConfig(({ command, mode, isPreview }) => ({
  plugins: [react()],
  publicDir: 'images',
  server: command === 'serve' && !isPreview ? {
    proxy: {
      '/api': {
        target: loadEnv(mode, process.cwd(), 'API_PROXY_').API_PROXY_TARGET || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  } : undefined,
}))
