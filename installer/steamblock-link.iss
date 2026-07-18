; Instalador clásico (wizard "siguiente-siguiente") de STEAMBLOCK Link.
; Empaqueta el binario Go + arduino-cli.exe + librerías. Estilo mBlock/OpenBlock.
;
; Compilar:  iscc /DAppVersion=0.1.0 installer\steamblock-link.iss
;   (Inno Setup en Windows, o `wine iscc.exe ...` desde Linux).
; Requiere que `make stage` haya dejado los archivos en dist\staging.

#ifndef AppVersion
  #define AppVersion "0.1.0"
#endif
#define AppName "STEAMBLOCK Link"
#define AppExe "steamblock-link.exe"
#define Publisher "STEAM Robotics Academy"

[Setup]
; AppId fijo: identifica el producto entre versiones (no cambiar).
AppId={{62A48511-659B-4CD9-866D-4F4862E80EA5}}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#Publisher}
DefaultDirName={autopf}\STEAMBLOCK Link
DefaultGroupName=STEAMBLOCK Link
DisableProgramGroupPage=yes
OutputDir=..\dist
OutputBaseFilename=STEAMBLOCK-Link-Setup-{#AppVersion}
SetupIconFile=..\cmd\steamblock-link\icon.ico
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
ArchitecturesInstallIn64BitMode=x64compatible
; Necesita admin para instalar en Program Files y escribir el autoarranque HKLM.
PrivilegesRequired=admin

[Languages]
Name: "es"; MessagesFile: "compiler:Languages\Spanish.isl"

[Tasks]
Name: "autostart"; Description: "Iniciar STEAMBLOCK Link automáticamente con Windows"; GroupDescription: "Opciones:"

[Files]
Source: "..\dist\staging\{#AppExe}"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\dist\staging\resources\*"; DestDir: "{app}\resources"; Flags: recursesubdirs createallsubdirs ignoreversion

[Icons]
Name: "{group}\STEAMBLOCK Link"; Filename: "{app}\{#AppExe}"
Name: "{group}\Desinstalar STEAMBLOCK Link"; Filename: "{uninstallexe}"

[Registry]
; Autoarranque por máquina (arranca en la sesión de cualquier alumno que entre).
Root: HKLM; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; \
  ValueType: string; ValueName: "STEAMBLOCK Link"; ValueData: """{app}\{#AppExe}"""; \
  Flags: uninsdeletevalue; Tasks: autostart

[Run]
; Lanzarlo al terminar la instalación (queda en la bandeja, sin ventana).
Filename: "{app}\{#AppExe}"; Description: "Iniciar STEAMBLOCK Link ahora"; \
  Flags: nowait postinstall skipifsilent

[UninstallRun]
; Cerrar el proceso antes de borrar los archivos.
Filename: "{sys}\taskkill.exe"; Parameters: "/IM {#AppExe} /F"; Flags: runhidden; RunOnceId: "KillLink"
