; Finicky for Windows — Inno Setup installer
; Registers Finicky as a browser so Windows routes http/https URLs to it.
; Per-user install (no admin required).

#define MyAppName "Finicky"
#define MyAppVersion "4.2.2"
#define MyAppPublisher "John Sterling"
#define MyAppURL "https://github.com/johnste/finicky"
#define MyAppExeName "Finicky.exe"

[Setup]
AppId={{7F3E2A1B-4C5D-6E7F-8A9B-0C1D2E3F4A5B}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}/issues
DefaultDirName={localappdata}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
OutputBaseFilename=FinickySetup-{#MyAppVersion}
Compression=lzma
SolidCompression=yes
PrivilegesRequired=lowest
SetupIconFile=..\apps\finicky\assets\Resources\finicky.ico
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
WizardStyle=modern
UninstallDisplayName={#MyAppName}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "autostart"; Description: "Start Finicky when Windows starts"; GroupDescription: "Additional options:"
Name: "setdefault"; Description: "Open Default Apps settings after install"; GroupDescription: "Additional options:"; Flags: unchecked

[Files]
Source: "..\apps\finicky\build\windows\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"

[Registry]
; ---- ProgID: FinickyURL ----
; This is the handler that runs when Windows invokes a URL assigned to Finicky.
Root: HKCU; Subkey: "SOFTWARE\Classes\FinickyURL"; ValueType: string; ValueData: "Finicky URL"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Classes\FinickyURL"; ValueName: "URL Protocol"; ValueType: string; ValueData: ""
Root: HKCU; Subkey: "SOFTWARE\Classes\FinickyURL\DefaultIcon"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"",0"
Root: HKCU; Subkey: "SOFTWARE\Classes\FinickyURL\shell\open\command"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"" ""%1"""

; ---- finicky:// custom protocol (always active, independent of default browser choice) ----
Root: HKCU; Subkey: "SOFTWARE\Classes\finicky"; ValueType: string; ValueData: "Finicky Protocol"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Classes\finicky"; ValueName: "URL Protocol"; ValueType: string; ValueData: ""
Root: HKCU; Subkey: "SOFTWARE\Classes\finicky\DefaultIcon"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"",0"
Root: HKCU; Subkey: "SOFTWARE\Classes\finicky\shell\open\command"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"" ""%1"""

; ---- Browser registration under StartMenuInternet ----
; This makes Finicky appear in Settings > Default Apps > Web browser.
Root: HKCU; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky"; ValueType: string; ValueData: "Finicky"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\DefaultIcon"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"",0"
Root: HKCU; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\shell\open\command"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"" --window"

; ---- Capabilities ----
; Tells Windows which URL schemes Finicky can handle.
Root: HKCU; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities"; ValueName: "ApplicationName"; ValueType: string; ValueData: "Finicky"
Root: HKCU; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities"; ValueName: "ApplicationDescription"; ValueType: string; ValueData: "A rule-based browser routing utility. Define rules to open URLs in different browsers based on the link domain, path, or source application."
Root: HKCU; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities"; ValueName: "ApplicationIcon"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"",0"
Root: HKCU; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities\URLAssociations"; ValueName: "http"; ValueType: string; ValueData: "FinickyURL"
Root: HKCU; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities\URLAssociations"; ValueName: "https"; ValueType: string; ValueData: "FinickyURL"
Root: HKCU; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities\URLAssociations"; ValueName: "finicky"; ValueType: string; ValueData: "FinickyURL"

; ---- RegisteredApplications ----
; The master index Windows checks to discover which apps offer URL handling.
Root: HKCU; Subkey: "SOFTWARE\RegisteredApplications"; ValueName: "Finicky"; ValueType: string; ValueData: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities"; Flags: uninsdeletevalue

; ---- Autostart (optional) ----
; Launch WITHOUT --window: autostart makes Finicky the resident URL router at
; login (useful with keepRunning), it must not pop the config window on boot.
Root: HKCU; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Run"; ValueName: "Finicky"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"""; Flags: uninsdeletevalue; Tasks: autostart

[Run]
; After install, optionally open Default Apps so the user can select Finicky
Filename: "cmd"; Parameters: "/c start ms-settings:defaultapps"; Description: "Open Default Apps settings"; Flags: postinstall shellexec skipifsilent nowait; Tasks: setdefault
; Launch Finicky after install
Filename: "{app}\{#MyAppExeName}"; Parameters: "--window"; Description: "Launch Finicky"; Flags: postinstall skipifsilent nowait

[UninstallRun]
; If Finicky is running during uninstall, terminate it
Filename: "taskkill"; Parameters: "/F /IM {#MyAppExeName}"; Flags: runhidden; RunOnceId: "KillFinicky"

[UninstallDelete]
; Clean up the IPC socket and cache
Type: filesandordirs; Name: "{localappdata}\Finicky"

[Code]
// Notify Windows that registered applications changed so the Default Apps
// list refreshes without requiring a reboot.
procedure CurStepChanged(CurStep: TSetupStep);
var
  Res: Integer;
begin
  if CurStep = ssPostInstall then
  begin
    // SHChangeNotify(SHCNE_ASSOCCHANGED, SHCNF_IDLIST, 0, 0)
    // Inno doesn't expose SHChangeNotify directly, but we can call it via
    // a small helper. The registry changes above are sufficient — Windows
    // picks them up on next Settings open. For immediate refresh we'd need
    // a DLL call, which is optional polish.
  end;
end;
