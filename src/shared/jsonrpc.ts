import type { IpcChannel, IpcArgs, IpcResult, IpcEvent, IpcEvents } from './ipc.js'

export const JSONRPC_VERSION = '2.0' as const

export interface JsonRpcRequest<C extends IpcChannel = IpcChannel> {
  jsonrpc: typeof JSONRPC_VERSION
  id: number
  method: C
  params: IpcArgs<C>
}

export interface JsonRpcSuccess<C extends IpcChannel = IpcChannel> {
  jsonrpc: typeof JSONRPC_VERSION
  id: number
  result: IpcResult<C>
}

export interface JsonRpcErrorResponse {
  jsonrpc: typeof JSONRPC_VERSION
  id: number | null
  error: { code: number; message: string }
}

export interface JsonRpcNotification<E extends IpcEvent = IpcEvent> {
  jsonrpc: typeof JSONRPC_VERSION
  method: E
  params: IpcEvents[E]
}

export type JsonRpcResponse = JsonRpcSuccess | JsonRpcErrorResponse

export const RpcErrorCode = {
  ParseError: -32700,
  InvalidRequest: -32600,
  MethodNotFound: -32601,
  InvalidParams: -32602,
  InternalError: -32603
} as const

export function isRequest(msg: unknown): msg is JsonRpcRequest {
  return (
    typeof msg === 'object' &&
    msg !== null &&
    (msg as JsonRpcRequest).jsonrpc === JSONRPC_VERSION &&
    typeof (msg as JsonRpcRequest).method === 'string' &&
    typeof (msg as JsonRpcRequest).id === 'number'
  )
}
