import { execFile } from 'node:child_process'
import { promisify } from 'node:util'
import { arduinoCli } from './arduino-env.js'

const run = promisify(execFile)

export const REQUIRED_CORES = ['arduino:avr'] as const

export const REQUIRED_LIBRARIES = [
  'Stepper',
  'Servo',
  'LiquidCrystal I2C',
  'DHT sensor library',
  'Adafruit Unified Sensor',
  'OneWire',
  'DallasTemperature',
  'Adafruit GFX Library',
  'Adafruit SSD1306',
  'Adafruit BMP280 Library',
  'Adafruit INA219',
  'BH1750',
  'IRremote',
  'MFRC522'
] as const

export interface LibrariesResult {
  ok: boolean
  message: string
}

export async function installLibraries(): Promise<LibrariesResult> {
  try {
    await run(arduinoCli(), ['lib', 'install', ...REQUIRED_LIBRARIES], { timeout: 600_000 })
    return { ok: true, message: `Librerías Arduino listas (${REQUIRED_LIBRARIES.length}).` }
  } catch (err) {
    return toResult(err, 'No se pudieron instalar las librerías.')
  }
}

export async function ensureCore(): Promise<LibrariesResult> {
  try {
    await run(arduinoCli(), ['core', 'update-index'], { timeout: 300_000 }).catch(() => {})
    for (const core of REQUIRED_CORES) {
      await run(arduinoCli(), ['core', 'install', core], { timeout: 900_000 })
    }
    return { ok: true, message: 'Core Arduino AVR listo.' }
  } catch (err) {
    return toResult(err, 'No se pudo instalar el core AVR.')
  }
}

export async function prepareArduino(): Promise<LibrariesResult> {
  const core = await ensureCore()
  const libs = await installLibraries()
  return { ok: core.ok && libs.ok, message: `${core.message} · ${libs.message}` }
}

function toResult(err: unknown, fallback: string): LibrariesResult {
  const e = err as { code?: string; stderr?: string; message?: string }
  if (e.code === 'ENOENT') return { ok: false, message: 'arduino-cli no encontrado.' }
  return { ok: false, message: e.stderr || e.message || fallback }
}
