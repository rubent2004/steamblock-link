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

## Instalación

### Descargar binario pre-compilado

Ve a [Releases](../../releases) y descarga el binario para tu plataforma:

| Plataforma | Archivo |
|-----------|---------|
| Linux x64 | `steamblock-link-linux-x64` |
| Windows x64 | `steamblock-link-win-x64.exe` |

### Desde npm (desarrollo)

```bash
npm install -g steamblock-link
```

## Uso

### Arrancar el Link

```bash
# Linux
./steamblock-link-linux-x64

# Windows
steamblock-link-win-x64.exe
```

Verás:

```
  ╔══════════════════════════════════════════╗
  ║       STEAMBLOCK Studio Link v0.1.0      ║
  ║   Companion local para el IDE web         ║
  ╚══════════════════════════════════════════╝

[steamblock-link] Escuchando en ws://127.0.0.1:20111
[steamblock-link] Esperando conexión del IDE web...
```

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

```bash
# Instalar dependencias
npm install

# Modo desarrollo (con tsx)
npm run dev

# Compilar TypeScript
npm run build

# Ejecutar
npm start
```

## Construir binarios

```bash
# Compilar + empaquetar con pkg
npm run dist
```

Los binarios se generan en `dist/`:

- `steamblock-link-linux-x64` (~40 MB)
- `steamblock-link-win-x64.exe` (~40 MB)

## Arquitectura

```
src/
├── index.ts              # Entry point (CLI, detección de arduino-cli)
├── server.ts             # Servidor WebSocket + JSON-RPC router
├── core.ts               # Barrel de servicios nativos
├── serial.service.ts     # Comunicación serie (serialport)
├── serial.sim.ts         # Dispositivo virtual (simulación)
├── compiler.service.ts   # Compilación con arduino-cli
├── libraries.ts          # Instalación de cores y librerías
├── arduino-env.ts        # Ruta del ejecutable arduino-cli
└── shared/
    ├── ipc.ts            # Contrato IPC (fuente de verdad)
    ├── jsonrpc.ts        # Tipos JSON-RPC 2.0
    └── types/
        └── board.ts      # Schema de placas (valibot)
```

## Licencia

MIT — STEAM Robotics Academy
