import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  publicDir: 'images',
  server: {
    // Bind mounts from Windows hosts do not deliver file events into the container.
    watch: { usePolling: process.env.VITE_USE_POLLING === 'true' },
  },
})
