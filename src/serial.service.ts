import { SerialPort } from 'serialport'
import type { SerialPortInfo } from './shared/ipc.js'
import { SIM_PORT, isSimPort, startSim, sendSim, stopSim } from './serial.sim.js'

let activePort: SerialPort | null = null
let simActive = false

function simEnabled(): boolean {
  const v = process.env.STUDIO_SIM_SERIAL
  return v === '1' || v === 'true'
}

export async function listPorts(): Promise<SerialPortInfo[]> {
  const ports = await SerialPort.list()
  const real = ports.map((p) => ({
    path: p.path,
    manufacturer: p.manufacturer ?? '',
    label: p.path
  }))
  return simEnabled() ? [SIM_PORT, ...real] : real
}

export interface SerialHandlers {
  onData: (data: string) => void
  onClose: () => void
}

export function connect(
  opts: { path: string; baudRate: number },
  handlers: SerialHandlers
): Promise<boolean> {
  if (isSimPort(opts.path)) {
    return disconnect().then(() => {
      startSim(handlers)
      simActive = true
      return true
    })
  }
  return disconnect().then(
    () =>
      new Promise<boolean>((resolve, reject) => {
        const port = new SerialPort({ path: opts.path, baudRate: opts.baudRate }, (err) => {
          if (err) return reject(err)
          activePort = port
          resolve(true)
        })
        port.on('data', (chunk: Buffer) => handlers.onData(chunk.toString()))
        const onGone = (): void => {
          if (activePort === port) activePort = null
          handlers.onClose()
        }
        port.on('close', onGone)
        port.on('error', onGone)
      })
  )
}

export async function disconnect(): Promise<boolean> {
  if (simActive) {
    simActive = false
    stopSim()
    return true
  }
  const port = activePort
  activePort = null
  if (!port || !port.isOpen) return true
  await new Promise<void>((resolve) => port.close(() => resolve()))
  return true
}

export function send(data: string): void {
  if (simActive) return sendSim(data)
  if (activePort?.isOpen) activePort.write(data + '\n')
}
