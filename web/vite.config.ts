import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath, URL } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { loadEnv, type Plugin } from 'vite'
import { defineConfig } from 'vitest/config'

const coreDir = fileURLToPath(new URL('../core', import.meta.url))
const outDir = join(coreDir, 'internal/presentation/http/ui/dist')

/** The embed directory must exist in git even when empty. */
function keepOutDir(): Plugin {
  return {
    name: 'keep-out-dir',
    closeBundle() {
      mkdirSync(outDir, { recursive: true })
      writeFileSync(join(outDir, '.gitkeep'), '')
    },
  }
}

export default defineConfig(({ mode }) => {
  const core = loadEnv(mode, coreDir, 'HTTP_SERVER_')
  return {
    plugins: [react(), tailwindcss(), keepOutDir()],
    build: { outDir, emptyOutDir: true },
    server: {
      proxy: { '/v1': `http://localhost:${core.HTTP_SERVER_PORT || '8000'}` },
    },
    resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
    test: {
      environment: 'jsdom',
      setupFiles: ['./src/test/setup.ts'],
      include: ['src/**/*.test.{ts,tsx}'],
      coverage: {
        provider: 'v8',
        reporter: ['text'],
        include: ['src/**/*.{ts,tsx}'],
        exclude: ['src/main.tsx', 'src/test/**', 'src/components/ui/**'],
      },
    },
  }
})
