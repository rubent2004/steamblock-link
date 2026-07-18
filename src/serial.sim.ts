import type { SerialPortInfo } from './shared/ipc.js'

export const SIM_PORT: SerialPortInfo = {
  path: 'SIM://steamblock',
  manufacturer: 'STEAMBLOCK',
  label: 'Simulador STEAMBLOCK (virtual)'
}

export function isSimPort(path: string): boolean {
  return path === SIM_PORT.path
}

interface SimHandlers {
  onData: (data: string) => void
  onClose: () => void
}

let timer: ReturnType<typeof setInterval> | null = null
let onData: ((data: string) => void) | null = null
let startedAt = 0

export function startSim(handlers: SimHandlers): void {
  stopSim()
  onData = handlers.onData
  startedAt = Date.now()
  onData('SIM: dispositivo virtual STEAMBLOCK conectado\n')
  timer = setInterval(() => {
    const s = Math.round((Date.now() - startedAt) / 1000)
    const temp = (20 + Math.random() * 8).toFixed(1)
    const dist = Math.floor(Math.random() * 200)
    onData?.(`t=${s}s temp=${temp}C dist=${dist}cm\n`)
  }, 1000)
  timer.unref?.()
}

export function sendSim(data: string): void {
  onData?.(`echo: ${data}\n`)
}

export function stopSim(): void {
  if (timer) clearInterval(timer)
  timer = null
  onData = null
}
