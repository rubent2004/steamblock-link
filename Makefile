# STEAMBLOCK Link — build del companion en Go.
#
# Windows: binario estático puro (sin cgo → sin mingw/wine), sin consola.
# Linux/macOS: la bandeja usa cgo (dbus/Cocoa), así que se compila en su propio SO.
#
#   make test              # tests de integración (WebSocket + JSON-RPC + simulador)
#   make windows           # dist/staging/steamblock-link.exe (cross desde Linux)
#   make stage             # + arduino-cli.exe y librerías junto al exe
#   make installer         # wizard Inno Setup (necesita iscc, nativo o vía wine)
#   make linux-stage       # binario Linux + resources (cgo, este SO)
#   make tarball           # dist/STEAMBLOCK-Link-linux-x86_64.tar.gz (universal/Arch)
#   make deb               # dist/steamblock-link_<ver>_amd64.deb (necesita nfpm)

VERSION      ?= 0.1.0
LDFLAGS_WIN  := -H=windowsgui -s -w
STAGE        := dist/staging
STAGE_LINUX  := dist/staging-linux

# Recursos a empaquetar. Por defecto se toman del Studio (donde `fetch:cli` ya
# bajó el arduino-cli.exe); sobreescribibles: make stage ARDUINO_CLI_WIN=/ruta.
ARDUINO_CLI_WIN ?= ../ide-bloques/studio/resources/arduino-cli/arduino-cli.exe
LIBS_DIR        ?= ../ide-bloques/studio/resources/arduino-libs

.PHONY: test linux windows stage installer linux-stage tarball deb clean

test:
	go test ./...

# Binario nativo (para probar la bandeja en esta máquina).
linux:
	go build -o dist/steamblock-link ./cmd/steamblock-link

# Ensambla el binario Linux + arduino-cli + carpeta de librerías, tal como
# quedarán en /opt/steamblock-link. Requiere GTK3 + appindicator (cgo).
linux-stage:
	mkdir -p $(STAGE_LINUX)/resources/arduino-cli $(STAGE_LINUX)/resources/arduino-libs
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(STAGE_LINUX)/steamblock-link ./cmd/steamblock-link
	curl -fSL https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_64bit.tar.gz -o dist/arduino-cli.tar.gz
	tar -xzf dist/arduino-cli.tar.gz -C $(STAGE_LINUX)/resources/arduino-cli arduino-cli
	@echo "Staging Linux listo en $(STAGE_LINUX)"

# Tarball universal: descomprimir y ejecutar. Funciona en Arch y cualquier distro.
tarball: linux-stage
	tar -czf dist/STEAMBLOCK-Link-linux-x86_64.tar.gz -C $(STAGE_LINUX) .

# Paquete .deb (Debian/Ubuntu/Mint...). Necesita nfpm en el PATH.
deb: linux-stage
	VERSION=$(VERSION) nfpm pkg --packager deb --config installer/nfpm.yaml --target dist/

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
