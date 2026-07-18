// Command steamblock-link es el companion local de STEAMBLOCK Studio: corre en
// la bandeja del sistema (sin ventana de consola) un servidor WebSocket que el
// IDE web usa para compilar y subir a placas Arduino.
//
// Distribución: binario estático único. En Windows se compila con
// `-ldflags -H=windowsgui` para que NO abra consola, y se instala con el wizard
// de installer/steamblock-link.iss. Análogo a Scratch Link / OpenBlock Link.
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/systray"
	"github.com/SARA-Robotics/steamblock-link/internal/link"
)

const defaultPort = "20111"

func main() {
	cfg := loadConfig()
	port := envOr("STUDIO_LINK_PORT", defaultPort)
	addr := "127.0.0.1:" + port
	url := "ws://" + addr

	// Instancia única: si el puerto ya está tomado, otro Link corre → salimos.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("[steamblock-link] el puerto %s ya está en uso (¿otro Link abierto?)", port)
		os.Exit(1)
	}

	srv := link.NewServer(cfg)
	httpSrv := &http.Server{Handler: http.HandlerFunc(srv.Handler)}

	app := &trayApp{url: url, httpSrv: httpSrv, ln: ln, srv: srv}
	systray.Run(app.onReady, app.onExit)
}

type trayApp struct {
	url     string
	httpSrv *http.Server
	ln      net.Listener
	srv     *link.Server
	status  *systray.MenuItem
}

func (a *trayApp) onReady() {
	systray.SetIcon(trayIcon)
	systray.SetTitle("")
	systray.SetTooltip("STEAMBLOCK Link — " + a.url)

	systray.AddMenuItem("STEAMBLOCK Link — activo", "").Disable()
	systray.AddMenuItem(a.url, "Dirección del servidor local").Disable()
	a.status = systray.AddMenuItem("Preparando entorno Arduino…", "")
	a.status.Disable()
	systray.AddSeparator()
	quit := systray.AddMenuItem("Salir", "Detiene el Link")

	// Servidor WebSocket (ya escuchando en loopback vía a.ln).
	go func() {
		if err := a.httpSrv.Serve(a.ln); err != nil && err != http.ErrServerClosed {
			log.Printf("[steamblock-link] servidor: %v", err)
		}
	}()
	log.Printf("[steamblock-link] escuchando en %s", a.url)

	// Preparar cores/librerías en segundo plano; refleja el resultado en el menú.
	go func() {
		r := a.srv.Compiler().Prepare(context.Background())
		a.setStatus(r.Message)
	}()

	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()
}

func (a *trayApp) onExit() {
	_ = a.httpSrv.Close()
}

func (a *trayApp) setStatus(msg string) {
	if a.status != nil {
		a.status.SetTitle(msg)
	}
}

// loadConfig arma la configuración del servidor desde el entorno y los recursos
// empaquetados junto al ejecutable.
func loadConfig() link.Config {
	return link.Config{
		SimEnabled: envTrue("STUDIO_SIM_SERIAL"),
		Token:      os.Getenv("STUDIO_LINK_TOKEN"),
		Origins:    splitCSV(os.Getenv("STUDIO_LINK_ORIGINS")),
		LibsDir:    resolveLibsDir(),
		CLIPath:    resolveArduinoCLI(),
	}
}

// resolveArduinoCLI busca el arduino-cli empaquetado junto al exe; si no está,
// cae al del PATH (nombre pelado, que exec resuelve).
func resolveArduinoCLI() string {
	bin := "arduino-cli"
	if isWindows() {
		bin = "arduino-cli.exe"
	}
	if p := filepath.Join(exeDir(), "resources", "arduino-cli", bin); fileExists(p) {
		return p
	}
	return bin
}

func resolveLibsDir() string {
	if d := os.Getenv("STUDIO_LINK_LIBRARIES_DIR"); d != "" {
		return d
	}
	return filepath.Join(exeDir(), "resources", "arduino-libs")
}

// --- helpers ---

func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func isWindows() bool { return os.PathSeparator == '\\' }

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envTrue(key string) bool {
	v := os.Getenv(key)
	return v == "1" || v == "true"
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
