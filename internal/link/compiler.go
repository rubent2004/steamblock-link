package link

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Cores y librerías que el Link garantiza instalados antes de compilar. Mismo
// conjunto que el steamblock-link de Node (libraries.ts).
var (
	requiredCores     = []string{"arduino:avr"}
	requiredLibraries = []string{
		"Stepper", "Servo", "LiquidCrystal I2C", "DHT sensor library",
		"Adafruit Unified Sensor", "OneWire", "DallasTemperature",
		"Adafruit GFX Library", "Adafruit SSD1306", "Adafruit BMP280 Library",
		"Adafruit INA219", "BH1750", "IRremote", "MFRC522",
	}
)

// Compiler compila y sube sketches usando un binario arduino-cli. cliPath es la
// ruta al binario (empaquetado junto al exe o encontrado en el PATH); libsDir es
// la carpeta de librerías vendoreadas que se pasa a `compile --libraries`.
type Compiler struct {
	cliPath string
	libsDir string
}

func NewCompiler(cliPath, libsDir string) *Compiler {
	return &Compiler{cliPath: cliPath, libsDir: libsDir}
}

// Check compila el sketch sin subirlo: valida que el código compila.
func (c *Compiler) Check(ctx context.Context, code string, board Board) UploadResult {
	if board.FQBN == "" {
		return UploadResult{Message: "La placa no define FQBN."}
	}
	return c.withSketch(code, func(dir string) UploadResult {
		if out, err := c.compile(ctx, board.FQBN, dir); err != nil {
			return failed(err, out, "Error al compilar.")
		}
		return UploadResult{OK: true, Message: "Código verificado: compila sin errores."}
	})
}

// Upload compila y sube el sketch al puerto dado. Si el puerto es el simulador,
// solo compila y reporta "subido" al dispositivo virtual.
func (c *Compiler) Upload(ctx context.Context, code string, board Board, port string) UploadResult {
	if board.FQBN == "" {
		return UploadResult{Message: "La placa no define FQBN."}
	}
	sim := isSimPort(port)
	if !sim && port == "" {
		return UploadResult{Message: "No hay puerto seleccionado."}
	}
	return c.withSketch(code, func(dir string) UploadResult {
		if out, err := c.compile(ctx, board.FQBN, dir); err != nil {
			return failed(err, out, "Error al compilar.")
		}
		if sim {
			return UploadResult{OK: true, Message: `Simulación: el código compila; "subido" al dispositivo virtual.`}
		}
		if out, err := c.run(ctx, "upload", "-p", port, "--fqbn", board.FQBN, dir); err != nil {
			return failed(err, out, "Error al subir.")
		}
		return UploadResult{OK: true, Message: "Subido correctamente."}
	})
}

// Prepare asegura los cores y librerías necesarios. Se corre en segundo plano al
// arrancar; los timeouts son amplios porque la primera vez descarga de internet.
func (c *Compiler) Prepare(ctx context.Context) UploadResult {
	// El índice puede fallar sin red; no es fatal si el core ya está instalado.
	_, _ = c.run(ctx, "core", "update-index")
	for _, core := range requiredCores {
		if out, err := c.run(ctx, "core", "install", core); err != nil {
			return failed(err, out, "No se pudo instalar el core AVR.")
		}
	}
	args := append([]string{"lib", "install"}, requiredLibraries...)
	if out, err := c.run(ctx, args...); err != nil {
		return failed(err, out, "No se pudieron instalar las librerías.")
	}
	return UploadResult{OK: true, Message: "Entorno Arduino listo (core AVR + librerías)."}
}

func (c *Compiler) compile(ctx context.Context, fqbn, dir string) (string, error) {
	return c.run(ctx, "compile", "--fqbn", fqbn, "--libraries", c.libsDir, dir)
}

// run ejecuta arduino-cli con un timeout generoso y devuelve la salida combinada
// (stdout+stderr) para poder mostrar el error de compilación al usuario.
func (c *Compiler) run(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.cliPath, args...)
	hideConsole(cmd) // Windows: sin ventana de consola parpadeante (no-op en otros SO).
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// withSketch escribe el código en un sketch temporal (sketch/sketch.ino) y lo
// borra al terminar. arduino-cli exige que el .ino esté en una carpeta homónima.
func (c *Compiler) withSketch(code string, fn func(dir string) UploadResult) UploadResult {
	base, err := os.MkdirTemp("", "steamblock-")
	if err != nil {
		return UploadResult{Message: "No se pudo crear el sketch temporal."}
	}
	defer os.RemoveAll(base)
	dir := filepath.Join(base, "sketch")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return UploadResult{Message: "No se pudo crear el sketch temporal."}
	}
	if err := os.WriteFile(filepath.Join(dir, "sketch.ino"), []byte(code), 0o644); err != nil {
		return UploadResult{Message: "No se pudo escribir el sketch."}
	}
	return fn(dir)
}

// failed traduce un error de exec a un UploadResult legible. Si arduino-cli no
// existe, lo dice claramente; si compiló con errores, muestra su salida.
func failed(err error, out, fallback string) UploadResult {
	if errors.Is(err, exec.ErrNotFound) {
		return UploadResult{Message: "arduino-cli no está instalado o no está en el PATH."}
	}
	if msg := strings.TrimSpace(out); msg != "" {
		return UploadResult{Message: msg}
	}
	return UploadResult{Message: fallback}
}
