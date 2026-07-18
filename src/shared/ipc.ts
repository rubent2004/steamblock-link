import type { Board } from './types/board.js'

export interface SerialPortInfo {
  path: string
  manufacturer: string
  label: string
}

export interface UploadResult {
  ok: boolean
  message: string
}

export interface IpcContract {
  'ports:list': { args: []; result: SerialPortInfo[] }
  'serial:connect': { args: [{ path: string; baudRate: number }]; result: boolean }
  'serial:disconnect': { args: []; result: boolean }
  'serial:send': { args: [string]; result: void }
  'upload:sketch': { args: [{ code: string; board: Board; port: string }]; result: UploadResult }
  'compile:check': { args: [{ code: string; board: Board }]; result: UploadResult }
  'project:save': {
    args: [{ path: string | null; data: string; suggestedName?: string }]
    result: { path: string } | null
  }
  'project:open': { args: []; result: { path: string; data: string } | null }
}

export interface IpcEvents {
  'serial:data': string
  'serial:disconnected': void
}

export type IpcChannel = keyof IpcContract
export type IpcArgs<C extends IpcChannel> = IpcContract[C]['args']
export type IpcResult<C extends IpcChannel> = IpcContract[C]['result']
export type IpcEvent = keyof IpcEvents

export interface StudioApi {
  ports: {
    list: () => Promise<SerialPortInfo[]>
  }
  serial: {
    connect: (opts: IpcArgs<'serial:connect'>[0]) => Promise<boolean>
    disconnect: () => Promise<boolean>
    send: (data: string) => Promise<void>
    onData: (cb: (data: string) => void) => () => void
    onDisconnected: (cb: () => void) => () => void
  }
  upload: {
    sketch: (payload: IpcArgs<'upload:sketch'>[0]) => Promise<UploadResult>
  }
  compile: {
    check: (payload: IpcArgs<'compile:check'>[0]) => Promise<UploadResult>
  }
  project: {
    save: (payload: IpcArgs<'project:save'>[0]) => Promise<IpcResult<'project:save'>>
    open: () => Promise<IpcResult<'project:open'>>
  }
}
