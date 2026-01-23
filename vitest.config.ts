import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/__tests__/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'src/__tests__/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/dist/**',
      ],
      all: true,
      lines: 80,
      functions: 80,
      branches: 80,
      statements: 80,
    },
    include: ['src/**/*.test.ts', 'src/**/*.test.tsx'],
    exclude: ['node_modules', 'dist', '.idea', '.git'],
    testTimeout: 10000,
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@models': path.resolve(__dirname, './src/models'),
      '@parsers': path.resolve(__dirname, './src/parsers'),
      '@analyzers': path.resolve(__dirname, './src/analyzers'),
      '@visualizers': path.resolve(__dirname, './src/visualizers'),
      '@engine': path.resolve(__dirname, './src/engine'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@workers': path.resolve(__dirname, './src/workers'),
      '@formatters': path.resolve(__dirname, './src/formatters'),
    },
  },
})
