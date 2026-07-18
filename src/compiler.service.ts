import { execFile } from 'node:child_process'
import { promisify } from 'node:util'
import { mkdtemp, mkdir, writeFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import type { Board } from './shared/types/board.js'
import type { UploadResult } from './shared/ipc.js'
import { isSimPort } from './serial.sim.js'
import { arduinoCli } from './arduino-env.js'

const run = promisify(execFile)

export interface UploadOptions {
  librariesDir: string
}

export async function checkSketch(
  payload: { code: string; board: Board },
  options: UploadOptions
): Promise<UploadResult> {
  if (!payload.board.fqbn) return { ok: false, message: 'La placa no define FQBN.' }
  return withSketch(payload.code, async (sketchDir) => {
    await compile(payload.board.fqbn!, sketchDir, options.librariesDir)
    return { ok: true, message: 'Código verificado: compila sin errores.' }
  })
}

export async function uploadSketch(
  payload: { code: string; board: Board; port: string },
  options: UploadOptions
): Promise<UploadResult> {
  const { code, board, port } = payload
  const sim = isSimPort(port)

  if (!board.fqbn) return { ok: false, message: 'La placa no define FQBN.' }
  if (!sim && !port) return { ok: false, message: 'No hay puerto seleccionado.' }

  return withSketch(code, async (sketchDir) => {
    await compile(board.fqbn!, sketchDir, options.librariesDir)
    if (sim) {
      return { ok: true, message: 'Simulación: el código compila; "subido" al dispositivo virtual.' }
    }
    await run(arduinoCli(), ['upload', '-p', port, '--fqbn', board.fqbn!, sketchDir])
    return { ok: true, message: 'Subido correctamente.' }
  })
}

function compile(fqbn: string, sketchDir: string, librariesDir: string): Promise<unknown> {
  return run(arduinoCli(), ['compile', '--fqbn', fqbn, '--libraries', librariesDir, sketchDir])
}

async function withSketch(
  code: string,
  fn: (sketchDir: string) => Promise<UploadResult>
): Promise<UploadResult> {
  const base = await mkdtemp(join(tmpdir(), 'steamblock-'))
  const sketchDir = join(base, 'sketch')
  try {
    await mkdir(sketchDir, { recursive: true })
    await writeFile(join(sketchDir, 'sketch.ino'), code)
    return await fn(sketchDir)
  } catch (err) {
    const e = err as { code?: string; stderr?: string; message?: string }
    if (e.code === 'ENOENT') {
      return { ok: false, message: 'arduino-cli no está instalado o no está en el PATH.' }
    }
    return { ok: false, message: e.stderr || e.message || 'Error al compilar o subir.' }
  } finally {
    await rm(base, { recursive: true, force: true })
  }
}
