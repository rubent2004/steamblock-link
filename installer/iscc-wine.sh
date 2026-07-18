#!/usr/bin/env bash
# Envuelve el compilador de Inno Setup (ISCC.exe) instalado bajo wine, para
# generar el instalador de Windows desde Linux/CachyOS. Uso: make installer
#
# Requisito (una sola vez):
#   curl -fsSL -o /tmp/is.exe \
#     https://github.com/jrsoftware/issrc/releases/download/is-6_7_3/innosetup-6.7.3.exe
#   wine /tmp/is.exe /VERYSILENT /SUPPRESSMSGBOXES /NORESTART /SP-
set -euo pipefail

ISCC="${ISCC_EXE:-$HOME/.wine/drive_c/Program Files (x86)/Inno Setup 6/ISCC.exe}"
if [[ ! -f "$ISCC" ]]; then
  echo "iscc-wine: no encuentro ISCC.exe en '$ISCC'." >&2
  echo "Instala Inno Setup bajo wine (ver cabecera de este script) o exporta ISCC_EXE." >&2
  exit 1
fi

WINEDEBUG="${WINEDEBUG:--all}" exec wine "$ISCC" "$@"
