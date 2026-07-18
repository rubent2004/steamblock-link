//go:build windows

package main

import _ "embed"

// En Windows la bandeja exige un icono en formato ICO.
//
//go:embed icon.ico
var trayIcon []byte
