/**
 * STEAMBLOCK Studio Link — servidor WebSocket local (companion web).
 *
 * Envuelve los servicios nativos (serial + compiler) y expone el contrato
 * shared/ipc.ts como JSON-RPC 2.0, para que el IDE corriendo en el navegador
 * pueda conectar, monitorear, compilar y subir a placas reales.
 *
 * Seguridad:
 *  - Bind SOLO a loopback (127.0.0.1), nunca 0.0.0.0.
 *  - Allowlist de Origin (defensa contra CSWSH).
 *  - Token compartido opcional vía ?token=.
 *  - Dueño único del puerto serie (una pestaña a la vez).
 */

import { WebSocketServer, type WebSocket } from 'ws'
import type { IncomingMessage } from 'node:http'
import { resolve } from 'node:path'
import {
  serial,
  uploadSketch,
  checkSketch,
  JSONRPC_VERSION,
  RpcErrorCode,
  isRequest,
  type IpcChannel,
  type IpcArgs,
  type IpcEvent,
  type IpcEvents,
  type JsonRpcRequest,
  type JsonRpcResponse,
  type JsonRpcNotification
} from './core.js'

export interface LinkServerOptions {
  /** Puerto TCP. Por defecto 20111. */
  port?: number
  /** Ruta de librerías Arduino vendoreadas. */
  librariesDir?: string
  /** Orígenes extra permitidos además de localhost/127.0.0.1. */
  origins?: string[]
  /** Token de sesión opcional. */
  token?: string
}

const DEFAULT_PORT = 20111

function isAllowedOrigin(origin: string | undefined, extra: string[]): boolean {
  if (!origin || origin === 'file://') return true
  if (/^https?:\/\/(localhost|127\.0\.0\.1)(:\d+)?$/.test(origin)) return true
  return extra.includes(origin)
}

function authorize(req: IncomingMessage, origins: string[], token?: string): boolean {
  if (!isAllowedOrigin(req.headers.origin, origins)) return false
  if (token) {
    const url = new URL(req.url ?? '/', 'http://localhost')
    if (url.searchParams.get('token') !== token) return false
  }
  return true
}

interface ServerState {
  serialOwner: WebSocket | null
}

export function createLinkServer(options: LinkServerOptions = {}): WebSocketServer {
  const port = options.port ?? DEFAULT_PORT
  const librariesDir = options.librariesDir ?? resolve(process.cwd(), 'resources', 'arduino-libs')
  const origins = options.origins ?? []
  const token = options.token
  const state: ServerState = { serialOwner: null }

  const wss = new WebSocketServer({
    host: '127.0.0.1',
    port,
    verifyClient: (info: { req: IncomingMessage }) => authorize(info.req, origins, token)
  })

  wss.on('connection', (socket) => {
    console.log('[steamblock-link] Cliente conectado')

    const notify = <E extends IpcEvent>(event: E, data: IpcEvents[E]): void => {
      const msg: JsonRpcNotification<E> = { jsonrpc: JSONRPC_VERSION, method: event, params: data }
      if (socket.readyState === socket.OPEN) socket.send(JSON.stringify(msg))
    }

    const ctx: Ctx = { socket, notify, librariesDir, state }
    socket.on('message', (raw) => handleMessage(ctx, raw.toString()))
    socket.on('close', () => {
      console.log('[steamblock-link] Cliente desconectado')
      if (state.serialOwner === socket) {
        state.serialOwner = null
        void serial.disconnect()
      }
    })
  })

  return wss
}

interface Ctx {
  socket: WebSocket
  notify: <E extends IpcEvent>(event: E, data: IpcEvents[E]) => void
  librariesDir: string
  state: ServerState
}

async function handleMessage(ctx: Ctx, raw: string): Promise<void> {
  let req: JsonRpcRequest
  try {
    const parsed = JSON.parse(raw)
    if (!isRequest(parsed)) throw new Error('no es una request JSON-RPC 2.0')
    req = parsed
  } catch {
    return send(ctx.socket, {
      jsonrpc: JSONRPC_VERSION,
      id: null,
      error: err(RpcErrorCode.ParseError, 'JSON inválido')
    })
  }

  try {
    const result = await dispatch(req.method, req.params, ctx)
    send(ctx.socket, { jsonrpc: JSONRPC_VERSION, id: req.id, result } as JsonRpcResponse)
  } catch (e) {
    const message = e instanceof Error ? e.message : 'error interno'
    send(ctx.socket, { jsonrpc: JSONRPC_VERSION, id: req.id, error: err(RpcErrorCode.InternalError, message) })
  }
}

async function dispatch(method: IpcChannel, params: IpcArgs<IpcChannel>, ctx: Ctx): Promise<unknown> {
  switch (method) {
    case 'ports:list':
      return serial.listPorts()
    case 'serial:connect': {
      const owner = ctx.state.serialOwner
      if (owner && owner !== ctx.socket && owner.readyState === owner.OPEN) {
        throw new Error('Otro cliente ya tiene el puerto abierto.')
      }
      const [opts] = params as IpcArgs<'serial:connect'>
      const ok = await serial.connect(opts, {
        onData: (data) => ctx.notify('serial:data', data),
        onClose: () => ctx.notify('serial:disconnected', undefined)
      })
      ctx.state.serialOwner = ctx.socket
      return ok
    }
    case 'serial:disconnect': {
      if (ctx.state.serialOwner === ctx.socket) ctx.state.serialOwner = null
      return serial.disconnect()
    }
    case 'serial:send': {
      const [data] = params as IpcArgs<'serial:send'>
      return serial.send(data)
    }
    case 'upload:sketch': {
      const [payload] = params as IpcArgs<'upload:sketch'>
      return uploadSketch(payload, { librariesDir: ctx.librariesDir })
    }
    case 'compile:check': {
      const [payload] = params as IpcArgs<'compile:check'>
      return checkSketch(payload, { librariesDir: ctx.librariesDir })
    }
    default:
      throw new Error(`método no soportado por el Link: ${method}`)
  }
}

function send(socket: WebSocket, msg: JsonRpcResponse): void {
  if (socket.readyState === socket.OPEN) socket.send(JSON.stringify(msg))
}

function err(code: number, message: string): { code: number; message: string } {
  return { code, message }
}
