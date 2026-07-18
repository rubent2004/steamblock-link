//go:build !windows

package main

import _ "embed"

// En Linux/macOS la bandeja usa un icono PNG.
//
//go:embed icon.png
var trayIcon []byte
