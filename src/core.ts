/**
 * core.ts — barrel que re-exporta los servicios nativos.
 *
 * Equivale a @steamblock/studio-core pero empaquetado directamente en el Link
 * (sin dependencia externa al workspace pnpm).
 */

export * as serial from './serial.service.js'
export { uploadSketch, checkSketch, type UploadOptions } from './compiler.service.js'
export { SIM_PORT, isSimPort } from './serial.sim.js'
export {
  installLibraries,
  ensureCore,
  prepareArduino,
  REQUIRED_LIBRARIES,
  REQUIRED_CORES,
  type LibrariesResult
} from './libraries.js'
export { setArduinoCli, arduinoCli } from './arduino-env.js'

// Contrato
export * from './shared/ipc.js'
export * from './shared/jsonrpc.js'
export * from './shared/types/board.js'
