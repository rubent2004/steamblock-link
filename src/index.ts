/**
 * STEAMBLOCK Link — servicio companion local (standalone, sin Electron).
 *
 * Servicio WebSocket que corre en la PC del usuario y permite al IDE web
 * compilar y subir sketches a placas Arduino. Análogo a Scratch Link /
 * OpenBlock Link.
 *
 * Uso:
 *   steamblock-link              # arranca en ws://127.0.0.1:20111
 *   STUDIO_LINK_TOKEN=abc123 steamblock-link  # con token de sesión
 *   STUDIO_SIM_SERIAL=1 steamblock-link       # modo simulación (sin placa)
 */

import { createLinkServer } from './server.js'
import { prepareArduino, setArduinoCli } from './arduino-env.js'
import { existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const PORT = Number(process.env.STUDIO_LINK_PORT) || 20111
const URL = `ws://127.0.0.1:${PORT}`

/** Detectar arduino-cli empaquetado (pkg) o en el PATH. */
function resolveArduinoCli(): void {
  // Cuando se empaqueta con pkg, los recursos están junto al ejecutable.
  const __dirname = dirname(fileURLToPath(import.meta.url))
  const bin = process.platform === 'win32' ? 'arduino-cli.exe' : 'arduino-cli'

  // Buscar en resources/ (junto al binario empaquetado)
  const bundled = join(__dirname, '..', 'resources', 'arduino-cli', bin)
  if (existsSync(bundled)) {
    setArduinoCli(bundled)
    console.log(`[steamblock-link] arduino-cli empaquetado: ${bundled}`)
    return
  }

  // Buscar en el directorio del ejecutable (para pkg)
  const nextToExe = join(process.cwd(), 'resources', 'arduino-cli', bin)
  if (existsSync(nextToExe)) {
    setArduinoCli(nextToExe)
    console.log(`[steamblock-link] arduino-cli local: ${nextToExe}`)
    return
  }

  console.log('[steamblock-link] arduino-cli: usando PATH')
}

function main(): void {
  console.log('')
  console.log('  ╔══════════════════════════════════════════╗')
  console.log('  ║       STEAMBLOCK Studio Link v0.1.0      ║')
  console.log('  ║   Companion local para el IDE web         ║')
  console.log('  ╚══════════════════════════════════════════╝')
  console.log('')

  resolveArduinoCli()

  const token = process.env.STUDIO_LINK_TOKEN || undefined
  const origins = (process.env.STUDIO_LINK_ORIGINS ?? '')
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)

  if (token) {
    console.log(`[steamblock-link] Token requerido: ${token.slice(0, 4)}...`)
  }

  const wss = createLinkServer({ port: PORT, token, origins })

  wss.on('listening', () => {
    console.log(`[steamblock-link] Escuchando en ${URL}`)
    console.log('[steamblock-link] Esperando conexión del IDE web...')
    console.log('')
    console.log('  Presiona Ctrl+C para detener.')
    console.log('')

    // Preparar Arduino en segundo plano (core + librerías).
    prepareArduino().then((r) => {
      console.log(`[steamblock-link] ${r.message}`)
    })
  })

  wss.on('error', (e: NodeJS.ErrnoException) => {
    if (e.code === 'EADDRINUSE') {
      console.error(`[steamblock-link] Puerto ${PORT} ya en uso. ¿Ya hay otro Link corriendo?`)
      process.exit(1)
    }
    console.error('[steamblock-link] Error:', e.message)
  })

  // Cierre limpio.
  process.on('SIGINT', () => {
    console.log('\n[steamblock-link] Cerrando...')
    wss.close()
    process.exit(0)
  })
  process.on('SIGTERM', () => {
    wss.close()
    process.exit(0)
  })
}

main()
