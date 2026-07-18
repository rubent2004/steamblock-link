# STEAMBLOCK Link — build del companion en Go.
#
# Windows: binario estático puro (sin cgo → sin mingw/wine), sin consola.
# Linux/macOS: la bandeja usa cgo (dbus/Cocoa), así que se compila en su propio SO.
#
#   make test              # tests de integración (WebSocket + JSON-RPC + simulador)
#   make windows           # dist/staging/steamblock-link.exe (cross desde Linux)
#   make stage             # + arduino-cli.exe y librerías junto al exe
#   make installer         # wizard Inno Setup (necesita iscc, nativo o vía wine)

VERSION      ?= 0.1.0
LDFLAGS_WIN  := -H=windowsgui -s -w
STAGE        := dist/staging

# Recursos a empaquetar. Por defecto se toman del Studio (donde `fetch:cli` ya
# bajó el arduino-cli.exe); sobreescribibles: make stage ARDUINO_CLI_WIN=/ruta.
ARDUINO_CLI_WIN ?= ../ide-bloques/studio/resources/arduino-cli/arduino-cli.exe
LIBS_DIR        ?= ../ide-bloques/studio/resources/arduino-libs

.PHONY: test linux windows stage installer clean

test:
	go test ./...

# Binario nativo (para probar la bandeja en esta máquina).
linux:
	go build -o dist/steamblock-link ./cmd/steamblock-link

# Cross-build a Windows: un comando, sin toolchain externo.
windows:
	mkdir -p $(STAGE)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		go build -ldflags="$(LDFLAGS_WIN)" -o $(STAGE)/steamblock-link.exe ./cmd/steamblock-link

# Ensambla el exe + arduino-cli.exe + librerías tal como quedarán instalados.
stage: windows
	mkdir -p $(STAGE)/resources/arduino-cli $(STAGE)/resources/arduino-libs
	cp "$(ARDUINO_CLI_WIN)" $(STAGE)/resources/arduino-cli/arduino-cli.exe
	cp -r "$(LIBS_DIR)/." $(STAGE)/resources/arduino-libs/
	@echo "Staging listo en $(STAGE)"

# Compila el instalador con Inno Setup. ISCC apunta al compilador: `iscc` nativo,
# o vía wine (ver installer/iscc-wine.sh, que envuelve ISCC.exe instalado en wine).
ISCC ?= ./installer/iscc-wine.sh

installer: stage
	$(ISCC) /DAppVersion=$(VERSION) installer/steamblock-link.iss

clean:
	rm -rf dist
