import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './ui'),
      '@models': path.resolve(__dirname, './ui/models'),
      '@parsers': path.resolve(__dirname, './ui/parsers'),
      '@analyzers': path.resolve(__dirname, './ui/analyzers'),
      '@visualizers': path.resolve(__dirname, './ui/visualizers'),
      '@components': path.resolve(__dirname, './ui/components'),
      '@utils': path.resolve(__dirname, './ui/utils'),
    },
  },
  build: {
    outDir: 'build',
    assetsDir: 'assets',
    minify: 'terser',
    sourcemap: false,
    rollupOptions: {
      output: {
        inlineIndex: true,
      },
    },
  },
})
