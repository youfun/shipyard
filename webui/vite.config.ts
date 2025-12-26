import { defineConfig } from 'vite'
import solidPlugin from 'vite-plugin-solid'
import tailwindcss from '@tailwindcss/vite'
import { resolve, dirname, join } from 'path'
import { fileURLToPath } from 'url'
const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

export default defineConfig({
  plugins: [solidPlugin(), tailwindcss()],
  resolve: {
    alias: {
      '@api': resolve(__dirname, 'src/api'),
      '@i18n': resolve(__dirname, 'src/i18n'),
      '@contexts': resolve(__dirname, 'src/contexts'),
      '@components': resolve(__dirname, 'src/components'),
      '@libs': resolve(__dirname, 'src/libs'),
      '@routes': resolve(__dirname, 'src/routes'),
      '@router': resolve(__dirname, 'src/router'),
      '@types': resolve(__dirname, 'src/types')

    }
  },
  build: {
    // Output to static directory for Go embed
     outDir: process.env.BUILD_OUTPUT_DIR
      ? process.env.BUILD_OUTPUT_DIR
      : join(__dirname, "..", "internal", "api", "webui"),
    // Optional: Static assets subdirectory (default "assets")
    assetsDir: "assets",
    // Optional: Clear output directory before build (default true)
    emptyOutDir: true,
    target: 'esnext',
    minify: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html')
      }
    }
  },
  server: {
    port: 3000,
    // Bind dev server explicitly to IPv4 to avoid mixed IPv6/IPv4 issues
    host: '127.0.0.1',
    // Proxy API requests to the Go backend during development
    proxy: {
      '/api': {
        // Use IPv4 loopback (127.0.0.1) instead of 'localhost' to avoid ::1/127.0.0.1 resolution mismatch
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        secure: false,
        ws: true,
        // log proxy errors so it's easier to debug connection issues
        configure: (proxy) => {
          proxy.on('error', (err: any, req: any) => {
            // default Vite logging will surface this, keeping minimal here
            console.error('[vite proxy] error', err?.message || err)
          })
        }
      }
    }
  }
})
