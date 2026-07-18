let cliPath = 'arduino-cli'

export function setArduinoCli(path: string): void {
  cliPath = path
}

export function arduinoCli(): string {
  return cliPath
}
