; Finicky for Windows — Inno Setup installer
; Registers Finicky as a browser so Windows routes http/https URLs to it.
;
; Install mode: admin (all-users) by default, per-user via /CURRENTUSER.
; Rationale (smoke-test finding F10): on current Windows 11 (observed on
; build 26200) the Default Apps picker only offers browsers registered
; machine-wide (HKLM) — a per-user (HKCU-only) registration is complete and
; correct but never appears as a selectable default browser. HKA roots below
; resolve to HKLM in admin mode and HKCU in per-user mode, so /CURRENTUSER
; still yields a working install for routing/protocol use on systems where
; per-user registration is honored.

#define MyAppName "Finicky"
; Overridable from CI: iscc /DMyAppVersion=<git describe> installer.iss
#ifndef MyAppVersion
  #define MyAppVersion "4.4.0-alpha"
#endif
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
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
OutputBaseFilename=FinickySetup-{#MyAppVersion}
Compression=lzma
SolidCompression=yes
; commandline-only override: interactive installs go straight to UAC (target
; audience is personal, unmanaged machines - product owner call, 2026-07-04);
; /CURRENTUSER remains available for the rare no-admin case, unadvertised.
PrivilegesRequired=admin
PrivilegesRequiredOverridesAllowed=commandline
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
; --window: clicking the Start Menu entry must open the config UI. A bare
; launch is the (windowless) resident-router mode used by autostart, and with
; a primary already running it exits silently - i.e. the shortcut appears dead.
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Parameters: "--window"
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"

[Registry]
; ---- ProgID: FinickyURL ----
; This is the handler that runs when Windows invokes a URL assigned to Finicky.
Root: HKA; Subkey: "SOFTWARE\Classes\FinickyURL"; ValueType: string; ValueData: "Finicky URL"; Flags: uninsdeletekey
Root: HKA; Subkey: "SOFTWARE\Classes\FinickyURL"; ValueName: "URL Protocol"; ValueType: string; ValueData: ""
Root: HKA; Subkey: "SOFTWARE\Classes\FinickyURL\DefaultIcon"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"",0"
Root: HKA; Subkey: "SOFTWARE\Classes\FinickyURL\shell\open\command"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"" ""%1"""

; ---- finicky:// custom protocol (always active, independent of default browser choice) ----
Root: HKA; Subkey: "SOFTWARE\Classes\finicky"; ValueType: string; ValueData: "Finicky Protocol"; Flags: uninsdeletekey
Root: HKA; Subkey: "SOFTWARE\Classes\finicky"; ValueName: "URL Protocol"; ValueType: string; ValueData: ""
Root: HKA; Subkey: "SOFTWARE\Classes\finicky\DefaultIcon"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"",0"
Root: HKA; Subkey: "SOFTWARE\Classes\finicky\shell\open\command"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"" ""%1"""

; ---- Browser registration under StartMenuInternet ----
; This makes Finicky appear in Settings > Default Apps > Web browser.
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky"; ValueType: string; ValueData: "Finicky"; Flags: uninsdeletekey
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\DefaultIcon"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"",0"
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\shell\open\command"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"" --window"

; ---- Capabilities ----
; Tells Windows which URL schemes Finicky can handle.
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities"; ValueName: "ApplicationName"; ValueType: string; ValueData: "Finicky"
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities"; ValueName: "ApplicationDescription"; ValueType: string; ValueData: "A rule-based browser routing utility. Define rules to open URLs in different browsers based on the link domain, path, or source application."
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities"; ValueName: "ApplicationIcon"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"",0"
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities\URLAssociations"; ValueName: "http"; ValueType: string; ValueData: "FinickyURL"
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities\URLAssociations"; ValueName: "https"; ValueType: string; ValueData: "FinickyURL"
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities\URLAssociations"; ValueName: "finicky"; ValueType: string; ValueData: "FinickyURL"

; ---- Browser-shape extras (required for the Win11 Default Apps picker) ----
; Win11 filters the http/https picker to apps that look like full browsers:
; FileAssociations + Startmenu + InstallInfo must exist alongside
; URLAssociations, or Finicky never appears as a selectable option.
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities\FileAssociations"; ValueName: ".htm"; ValueType: string; ValueData: "FinickyURL"
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities\FileAssociations"; ValueName: ".html"; ValueType: string; ValueData: "FinickyURL"
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities\Startmenu"; ValueName: "StartMenuInternet"; ValueType: string; ValueData: "Finicky"
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\InstallInfo"; ValueName: "ReinstallCommand"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"" --window"
Root: HKA; Subkey: "SOFTWARE\Clients\StartMenuInternet\Finicky\InstallInfo"; ValueName: "IconsVisible"; ValueType: dword; ValueData: 1

; ---- RegisteredApplications ----
; The master index Windows checks to discover which apps offer URL handling.
Root: HKA; Subkey: "SOFTWARE\RegisteredApplications"; ValueName: "Finicky"; ValueType: string; ValueData: "SOFTWARE\Clients\StartMenuInternet\Finicky\Capabilities"; Flags: uninsdeletevalue

; ---- Autostart (optional) ----
; Launch WITHOUT --window: autostart makes Finicky the resident URL router at
; login (useful with keepRunning), it must not pop the config window on boot.
Root: HKA; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Run"; ValueName: "Finicky"; ValueType: string; ValueData: """{app}\{#MyAppExeName}"""; Flags: uninsdeletevalue; Tasks: autostart

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
// picker refreshes. NOT optional: without this (verified on Win11 26200),
// Settings does not see the new registration until a shell restart/logon.
procedure SHChangeNotify(wEventID, uFlags, dwItem1, dwItem2: Integer);
  external 'SHChangeNotify@shell32.dll stdcall';

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
    SHChangeNotify($08000000, 0, 0, 0); // SHCNE_ASSOCCHANGED, SHCNF_IDLIST
end;
