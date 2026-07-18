//go:build windows

package link

import (
	"os/exec"
	"syscall"
)

// CREATE_NO_WINDOW: evita que Windows abra una ventana de consola al lanzar
// arduino-cli desde una app sin consola (-H=windowsgui). Sin esto, cada llamada
// (update-index, core install, cada lib) hace parpadear una consola negra.
const createNoWindow = 0x08000000

// hideConsole configura el subproceso para que corra sin ventana visible.
func hideConsole(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
