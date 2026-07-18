# STEAMBLOCK Link

Companion local para [STEAMBLOCK Studio](https://github.com/rubent2004/steamblock-studio) — el IDE de bloques educativo de STEAM Robotics Academy.

STEAMBLOCK Link es un **servicio WebSocket** que corre en la PC del usuario y permite al IDE web **compilar y subir** sketches a placas Arduino reales. Es el análogo de [Scratch Link](https://en.scratch-wiki.org/wiki/Scratch_Link) y [OpenBlock Link](https://github.com/openblockcc/openblock-link).

## ¿Por qué existe?

El navegador **no puede** ejecutar `arduino-cli` ni acceder al puerto serie directamente. STEAMBLOCK Link resuelve esto corriendo un pequeño servicio local que:

- Expone un **WebSocket en `127.0.0.1:20111`** (JSON-RPC 2.0)
- Gestiona la **comunicación serie** con la placa
- Ejecuta `arduino-cli` para **compilar y subir** sketches
- Instala automáticamente el **core AVR** y las **librerías** necesarias

```
┌──────────────────────┐     WebSocket (localhost)     ┌─────────────────────────┐
│  Navegador            │  ─────────────────────────►  │  STEAMBLOCK Link         │
│  studio.steamblock    │   { method, params }          │  (servicio local)        │
│                      │  ◄─────────────────────────   │                          │
│  Vue + Blockly       │   { result } / push events    │  serial + arduino-cli    │
└──────────────────────┘                                └─────────────────────────┘
```

> **Implementado en Go** (binario estático único). Antes fue TypeScript/Node
> empaquetado con `pkg`; se migró a Go para bajar el tamaño del instalador (~10 MB
> vs ~50 MB de runtime), correr **sin ventana de consola** y poner un **icono en la
> bandeja** como mBlock/OpenBlock. El contrato JSON-RPC es idéntico.

## Instalación

### Windows (instalador clásico)

Descarga `STEAMBLOCK-Link-Setup-x.y.z.exe` de [Releases](../../releases) y sigue
el wizard (siguiente-siguiente). Al terminar, el Link queda como **icono en la
bandeja** (junto al reloj) y, si marcaste la casilla, **arranca solo con Windows**.
No hay que abrir ninguna consola.

### Linux

Descarga el binario `steamblock-link` y ejecútalo: aparece en la bandeja del
escritorio. (Requiere un panel con soporte de AppIndicator/StatusNotifier.)

## Uso

Una vez instalado, el Link corre en segundo plano en la **bandeja del sistema**.
El menú del icono muestra la dirección (`ws://127.0.0.1:20111`), el estado del
entorno Arduino y la opción **Salir**.

### Abrir el IDE web

Abre [studio.steamblock.web](https://studio.steamblock.web) (o tu instancia local) y el IDE detectará automáticamente el Link.

## Variables de entorno

| Variable | Descripción | Default |
|----------|-------------|---------|
| `STUDIO_LINK_PORT` | Puerto del WebSocket | `20111` |
| `STUDIO_LINK_TOKEN` | Token de sesión (opcional) | — |
| `STUDIO_LINK_ORIGINS` | Orígenes extra permitidos (comma-separated) | — |
| `STUDIO_SIM_SERIAL` | `1` para modo simulación (sin placa) | — |

### Ejemplos

```bash
# Con token de sesión
STUDIO_LINK_TOKEN=misecretoken ./steamblock-link-linux-x64

# Modo simulación (para desarrollo sin placa)
STUDIO_SIM_SERIAL=1 ./steamblock-link-linux-x64

# Puerto personalizado
STUDIO_LINK_PORT=20222 ./steamblock-link-linux-x64
```

## Seguridad

- **Loopback only**: escucha solo en `127.0.0.1`, nunca en `0.0.0.0`
- **Allowlist de Origin**: solo acepta conexiones de orígenes permitidos
- **Token opcional**: gate extra para despliegues controlados
- **Dueño único del puerto serie**: una sola pestaña maneja la placa a la vez

## Protocolo

El Link usa **JSON-RPC 2.0 sobre WebSocket**. Los canales derivan del contrato `shared/ipc.ts`:

| Canal | Dirección | Descripción |
|-------|-----------|-------------|
| `ports:list` | request | Lista puertos serie disponibles |
| `serial:connect` | request | Conecta a un puerto |
| `serial:disconnect` | request | Desconecta |
| `serial:send` | request | Envía datos al puerto |
| `upload:sketch` | request | Compila y sube un sketch |
| `compile:check` | request | Verifica sin subir |
| `serial:data` | push | Datos recibidos del puerto |
| `serial:disconnected` | push | Puerto desconectado |

## Desarrollo

Requiere **Go 1.23+**. Para la bandeja en Linux/macOS hace falta cgo (gcc y
DBus/Cocoa del sistema); en Windows es Go puro.

```bash
# Tests de integración (WebSocket + JSON-RPC + simulador, sin hardware)
make test

# Binario nativo para probar la bandeja en esta máquina
make linux
STUDIO_SIM_SERIAL=1 ./dist/steamblock-link   # con placa virtual
```

## Construir el instalador de Windows

Desde Linux/CachyOS, **sin mingw ni wine** (en Windows `systray` usa syscalls
Win32 puros, así que el cross-build es Go puro):

```bash
# 1) exe (7 MB, GUI sin consola) + arduino-cli.exe + librerías → dist/staging
make stage

# 2) wizard Inno Setup → dist/STEAMBLOCK-Link-Setup-x.y.z.exe
make installer            # necesita iscc (Inno Setup nativo, o `wine iscc.exe`)
```

`make windows` a solas hace el cross-build:

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -ldflags="-H=windowsgui -s -w" -o steamblock-link.exe ./cmd/steamblock-link
```

El instalador final pesa ~20-25 MB (arduino-cli comprime bien con lzma2).

## Arquitectura

```
cmd/steamblock-link/
├── main.go             # entrypoint: bandeja (systray) + arranque del servidor
├── icon_windows.go     # icono ICO (bandeja Windows)
└── icon_other.go       # icono PNG (bandeja Linux/macOS)
internal/link/
├── server.go           # servidor WebSocket + origen/token (loopback only)
├── dispatch.go         # router JSON-RPC (canales del contrato)
├── protocol.go         # tipos JSON-RPC 2.0 + dominio (Board, puertos, resultado)
├── serial.go           # comunicación serie (go.bug.st/serial), dueño único
├── sim.go              # dispositivo virtual (STUDIO_SIM_SERIAL)
├── compiler.go         # compilar/subir y preparar cores/librerías (arduino-cli)
└── server_test.go      # tests de integración end-to-end
installer/
└── steamblock-link.iss # script del wizard (Inno Setup)
Makefile                # test · linux · windows · stage · installer
```

> El código TypeScript original queda en `src/` como referencia durante la
> transición; una vez validado el binario Go se puede retirar junto a `node_modules/`.

## Licencia

MIT — STEAM Robotics Academy
