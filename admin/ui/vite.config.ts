import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))

// In dev (`make ui-dev` or `npm run dev`) Vite serves on :5173 and proxies
// /nucleus.admin.v1.* (Connect-RPC routes) and /healthz to the admin server
// running on :8080. In prod the UI is built and embedded into the admin
// server binary via //go:embed all:ui_dist (admin/server/ui/embed.go in
// Phase 4), so the proxy is only relevant for development.
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.join(here, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      // Connect-RPC paths are /<package>.<Service>/<Method>. We proxy the
      // whole admin namespace.
      '/nucleus.admin.v1.': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: false,
      },
      '/healthz': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: false,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    emptyOutDir: true,
  },
})
