//go:build !windows

package link

import "os/exec"

// En Linux/macOS no hay ventanas de consola que ocultar: no-op.
func hideConsole(cmd *exec.Cmd) {}
