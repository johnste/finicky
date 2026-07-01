# Windows Port — QC Checklist & Findings

This document records the quality-control methodology used before the Windows port was opened for community review.
It serves as a reusable checklist for future contributors, release engineers, and anyone who wants to extend or audit the Windows layer.

---

## How to run a QC pass

### 1. API parity check

For every file that has both a `_darwin.go` and a `_windows.go` counterpart, verify that all **exported** symbols (functions, types, vars) defined on one platform are also present on the other with a compatible signature. Intentional differences are fine — document them.

Files to compare:

| darwin | windows | intentional differences |
|---|---|---|
| `main.go` | `main_windows.go` | URL receiving: Apple Events vs argv+unix socket |
| `browser.go` | `browser_windows.go` | `setDefaultBrowser` opens Settings app on Windows (OS-enforced) |
| `browser/detect.go` | `browser/detect_windows.go` | Registry walk vs `mdfind`/plist scan |
| `browser/launcher.go` | `browser/launcher_windows.go` | Profile paths via `%LOCALAPPDATA%` / `%APPDATA%` |
| `util/info.go` | `util/info_windows.go` | `GetForegroundWindowTitle` present only on Windows (future opener info) |
| `util/directories.go` | `util/directories_windows.go` | Delegates entirely to stdlib — no differences |
| `window/window.go` | `window/window_windows.go` | WKWebView → WebView2; WKURLSchemeHandler → local HTTP server |

**Checklist:**
- [ ] Every exported function in `_darwin.go` has a matching counterpart in `_windows.go`
- [ ] Function signatures match (same argument types and return types)
- [ ] Any intentional difference is noted in the table above

---

### 2. Race condition analysis

The Windows event model is different from macOS. In particular:

- `window_windows.go` runs WebView2 in a goroutine (`w.Run()` blocks).
- `main_windows.go` receives URLs on the `listenForURLs()` goroutine.
- `SendMessageToWebView` is called from multiple goroutines (URL handler, config watcher, update checker).

**Shared mutable state to audit:**

| Variable | File | Guard | Risk |
|---|---|---|---|
| `wv` (WebView2 instance) | `window_windows.go` | `queueMutex` | Read/write from multiple goroutines |
| `windowReady` | `window_windows.go` | `queueMutex` | Set in goroutine launched by `ShowWindow()` |
| `messageQueue` | `window_windows.go` | `queueMutex` | Appended by `SendMessageToWebView`, drained by `NavigationComplete` |
| `vm` (JS VM) | `main_windows.go` | None — single goroutine (event loop owns it) | OK: all writes and reads are serialized through the `select` loop |
| `shouldKeepRunning` | `main_windows.go` | None — event loop only | OK: same reason |
| `pipeAddress` | `main_windows.go` | `pipeOnce` (sync.Once) | Safe |

**Checklist:**
- [ ] Every write to `wv` is under `queueMutex`
- [ ] Every write to `windowReady` is under `queueMutex`
- [ ] `sendMessageToWebViewInternal` is only called with `queueMutex` held (check all call sites)
- [ ] When `w.Run()` returns (window closed), `wv = nil` and `windowReady = false` are set inside the mutex before the close is signalled
- [ ] No goroutine calls `w.Eval()` / `w.Navigate()` directly — all WebView2 calls go through `Dispatch()`

---

### 2b. Thread-affinity audit (GUI thread ownership)

**This is the most important Windows-specific check and the easiest to get wrong.** Data-race analysis (section 2) asks "is this variable safe to touch from two goroutines?" Thread-affinity asks a different question: "is this call happening on the *one specific OS thread* that is allowed to make it?" Win32 GUI objects are thread-affine — window messages are delivered only to the thread that created the `HWND`, and the message loop must run on that same thread. A WebView2 window created on one thread and pumped on another is silently broken: it appears but never paints or accepts input, and WebView2's controller init may crash.

Go makes this trap easy to fall into, because goroutines migrate between OS threads freely unless pinned with `runtime.LockOSThread()`. The rule the port must obey (mirroring how the macOS build runs all of Cocoa on the main thread):

> The WebView2 window is **created, message-pumped, and torn down on a single OS thread**, and that thread is pinned with `runtime.LockOSThread()`. All work from other goroutines reaches it via `webview.Dispatch()`, never directly.

**Checklist:**
- [ ] `main()` calls `runtime.LockOSThread()` and then keeps the main goroutine as the UI owner (it runs the window loop, not the event loop)
- [ ] The event loop runs on a **background goroutine**; the **main thread** runs `runUILoop` → `window.RunWindow()`
- [ ] `window.RunWindow()` both **creates** the webview and calls `w.Run()` — on the same call stack, same thread. Creation and the message loop are never split across goroutines
- [ ] Show-window requests cross from the event-loop goroutine to the main thread via a channel (`showWindowChan`), not by calling `RunWindow` directly off-thread
- [ ] Every `SendMessageToWebView` from a non-UI goroutine ends in `wv.Dispatch(...)`, which marshals onto the pumping thread
- [ ] Re-opening a closed window creates a fresh webview on the same locked thread (no stale `HWND` reuse)

**How to catch a regression:** on a real Windows machine, trigger the config window (`Finicky.exe --window`). If it opens but is blank/frozen/unclickable, thread-affinity is broken even though `go build` and `go test` pass — this class of bug is invisible to the compiler and to macOS-hosted tests.

---

### 3. Lifecycle signal audit

The macOS port relies on OS-provided lifecycle callbacks (Apple Events, `WKNavigationDelegate`). The Windows port must manually wire equivalent signals.

| Signal | macOS mechanism | Windows mechanism | Status |
|---|---|---|---|
| URL received from OS | Apple Event `kAEGetURL` | `argv[1]` on process launch → unix socket IPC | ✅ Implemented |
| Window closed | Cocoa delegate callback | `w.Run()` returns on main thread → `windowClosed` channel | ✅ Implemented |
| Page load complete | `WKNavigationDelegate.didFinish` | `w.Init(...)` injects `window.addEventListener("load", () => __finicky_ready())` | ✅ Implemented |
| Message queue drain | Triggered by page load | `NavigationComplete()` called via `__finicky_ready` binding | ✅ Implemented |
| Single instance | NSRunningApplication check | Named mutex `Global\FinickyBrowserRouter` | ✅ Implemented |

**Checklist:**
- [ ] `__finicky_ready` binding is registered via `w.Bind()` before `w.Navigate()`
- [ ] `w.Init(...)` script fires `__finicky_ready()` on `window` load event (not DOMContentLoaded — assets must be ready)
- [ ] `NavigationComplete()` drains `messageQueue` and sets `windowReady = true` inside `queueMutex`
- [ ] The UI loop signals `windowClosed` after `RunWindow()` returns, and the event loop resets `showingWindow` so the window can be reopened
- [ ] `sendToPrimary` retries the socket dial (covering the mutex-created-before-socket startup race) and, on ultimate failure, only falls back to `cmd /c start` when Finicky is **not** the default browser (else it would relaunch itself in a loop)

---

### 4. Dead code and import audit

After any cross-platform file is created or edited, run:

```bash
cd apps/finicky/src
GOOS=windows CGO_ENABLED=0 go build ./... 2>&1 | grep -E "imported and not used|declared and not used"
GOOS=darwin CGO_ENABLED=1 go build ./... 2>&1 | grep -E "imported and not used|declared and not used"
```

**Checklist:**
- [ ] No unused imports on either platform
- [ ] No unused constants (e.g. named pipe path constants left over from earlier design iterations)
- [ ] No `var _ = ...` import silencers (these mask real unused-import issues)

---

### 5. Error handling — critical paths

Silent failures in a URL router are user-visible: the user clicks a link, nothing happens.

**Critical paths to audit:**

| Path | What can go wrong | Mitigation |
|---|---|---|
| `forwardURLToPipe` | Pipe not yet up (race at startup); stale socket | Falls back to `cmd /c start` — URL opens in system default |
| `listenForURLs` | Stale `.sock` file from crash | Removes and retries |
| `findBrowserExe` | Browser installed in non-standard path | Falls back to `cmd /c start ""` which lets Windows resolve the browser |
| `startAssetServer` | Port binding failure | Falls back to `w.SetHtml()` with embedded HTML |
| `LaunchBrowser` | Profile not found | Logs warning, proceeds without profile flag (browser opens default profile) |
| `GetInstalledBrowsers` | Registry key missing | Returns empty slice (UI shows "no browsers found") |

**Checklist:**
- [ ] `forwardURLToPipe` has `cmd /c start` fallback and `HideWindow: true` on SysProcAttr
- [ ] `listenForURLs` removes stale socket before retrying bind
- [ ] `startAssetServer` falls back to `SetHtml` on listener error
- [ ] `LaunchBrowser` uses `cmd /c start ""` fallback when `findBrowserExe` returns empty
- [ ] Profile resolution failures log a warning and continue (not fatal)

---

### 6. Cross-compile verification

Run this before every release candidate:

```bash
cd apps/finicky/src

# Windows cross-compile (the primary deliverable)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -v \
  -ldflags="-H windowsgui -X finicky/version.GitCommit=$(git rev-parse --short HEAD) -X finicky/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o /tmp/Finicky.exe \
  .

# macOS native build (regression check — must not break)
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
  go build -v \
  -o /tmp/Finicky-mac \
  .

# Test suite (both platforms share the same test files)
go test ./...
```

**Expected results:**
- Windows binary: PE32+ executable (GUI), ~28-35 MB depending on assets
- macOS binary: Mach-O 64-bit, arm64
- Tests: 4 suites pass (resolver, rules, util, version)

---

## Findings from initial QC pass (2026-06-30)

Six bugs were found and fixed before community release.

| # | Severity | File | Bug | Fix |
|---|---|---|---|---|
| 1 | **Critical** | `window_windows.go` | `NavigationComplete()` was defined but never called — the Svelte UI would receive a blank screen because the message queue never drained | Added `w.Init(...)` script that fires `window.__finicky_ready()` on page load; bound via `w.Bind("__finicky_ready", NavigationComplete)` |
| 2 | **Critical** | `window_windows.go` | `wv` read/written from multiple goroutines without synchronization — race condition that could panic or corrupt state | All `wv` accesses now under `queueMutex`; `sendMessageToWebViewInternal` documented as requiring lock held |
| 3 | **High** | `window_windows.go` | `wv = nil` and `windowReady = false` set outside `queueMutex` in `Run()` callback goroutine | Wrapped in `queueMutex.Lock()/Unlock()` |
| 4 | **Medium** | `main_windows.go` | `forwardURLToPipe` silently dropped URLs when the unix socket dial failed (e.g. main instance still starting) — user click lost | Added `exec.Command("cmd", "/c", "start", "", url)` fallback with `HideWindow: true` |
| 5 | **Low** | `main_windows.go` | Unused constant `pipeName = "\\\\.\pipe\finicky-url"` — leftover from Named Pipe design iteration; code uses unix sockets | Removed |
| 6 | **Low** | `browser/launcher_windows.go`, `window_windows.go` | Unused import silencers (`var _ = util.ShortenPath`, `var _ = filepath.Base`) | Removed; cleaned imports |

## Findings from second QC pass (2026-07-01)

A deeper architectural review — focused on threading, Win32 API contracts, and cross-referencing the installer against the runtime code — found five more issues. These are the kind that pass every compiler and macOS-hosted test but fail on real Windows, so they were the highest-value catches before community testing.

| # | Severity | File | Bug | Fix |
|---|---|---|---|---|
| 7 | **Critical** | `main_windows.go`, `window_windows.go` | **Broken GUI thread affinity.** The webview was *created* on a `go ShowConfigWindow()` goroutine and *message-pumped* on a separate `go func(){ w.Run() }()` goroutine — neither being the `LockOSThread`-pinned main thread. Win32 delivers window messages only to the creating thread, so the config window would open blank/frozen and WebView2 init could crash. | Restructured so the **main (locked) thread owns the window**: event loop moved to a background goroutine; `runUILoop` on the main thread runs `window.RunWindow()`, which creates *and* pumps the webview on one thread. Cross-thread show requests go through `showWindowChan`. Mirrors the macOS "all UI on main thread" model. |
| 8 | **High** | `main_windows.go` | **Racy single-instance detection.** `GetLastError` was read via a *separate* `LazyProc.Call`, which can be clobbered by the Go runtime's own syscalls between it and `CreateMutexW` — so `ERROR_ALREADY_EXISTS` could be missed and two instances both act as primary. | Use the `err` value returned directly by `procCreateMutexW.Call(...)` (captured by the runtime immediately after the syscall). Removed the separate `getLastError` proc. |
| 9 | **High** | `browser_windows.go` | **`isDefaultBrowser()` always returned false.** It compared the UserChoice ProgID to the literal `"Finicky"`, but the installer registers the http/https association as ProgID `"FinickyURL"`. The two never matched. | Compare against a shared `progID = "FinickyURL"` constant that matches `installer.iss`. |
| 10 | **High** | `browser_windows.go`, `main_windows.go` | **Settings popped open on every link click.** `setDefaultBrowser()` responded to "not default" by launching `ms-settings:defaultapps`, and it ran at every startup — so (amplified by #9) the Default Apps page opened on every URL invocation. | Startup now only *observes* default status (`logDefaultBrowserStatus`) and never prompts, matching the fact that Windows 10+ requires the user to set defaults manually (the installer already offers this opt-in). Removed the unusable programmatic-set path. |
| 11 | **Medium** | `main_windows.go` | **URL loss / relaunch-loop risk in IPC handoff.** `forwardURLToPipe` dialed the primary's socket exactly once; during the primary's startup the mutex exists before the socket is listening, so the URL was dropped. Its `cmd /c start` fallback could also re-invoke Finicky if Finicky was the default browser, risking a launch loop. | `sendToPrimary` retries the dial for ~1s (covers the startup race); the `start` fallback runs **only when Finicky is not the default browser**, otherwise the URL is dropped with a clear log instead of looping. Also added a `--window` handoff so a second launch focuses the existing window (Start Menu parity). |

**Also fixed:** `installer.iss` autostart entry launched `Finicky.exe --window`, which would pop the config window on every login. Changed to launch with no arguments (resident router only). And the asset HTTP server is now closed when the window closes (was leaked on every open).

---

## Known limitations (not bugs)

These are documented design decisions, not defects:

- **Finicky cannot make itself the default browser programmatically on Windows 10+.** Microsoft removed programmatic default-browser setting after Windows 8. So, unlike the macOS build (which sets the default via `LSSetDefaultHandlerForURLScheme` behind an OS consent dialog), the Windows build only *detects* the current default and logs it — it never prompts on launch. Setting Finicky as default is a user action, offered by the installer's opt-in "Open Default Apps settings" task. A future in-app "Set as default" button could call the same `ms-settings:defaultapps` deep link.

- **Opener info is not available on Windows.** On macOS, `OpenerInfo` captures the frontmost app when a URL is clicked. On Windows, `GetForegroundWindowTitle` is implemented but the Windows process launch model (new process per URL) makes reliable opener detection impractical — by the time Finicky launches, the original foreground window may have changed. `OpenerInfo` is passed as an empty struct.

- **Browser detection is hardcoded + registry-based.** `findBrowserExe` uses a hardcoded path table for the 10 most common browsers. `detect_windows.go` supplements this via the `StartMenuInternet` registry key. Exotic browsers in non-standard install locations will fall through to the `cmd /c start` fallback (which uses Windows' own default browser association).

- **WebView2 runtime must be present.** `go-webview2` embeds `WebView2Loader.dll` via `go-winloader` but still requires the WebView2 runtime to be installed. It ships with Windows 11 and is available via Windows Update on Windows 10. The installer (`scripts/installer.iss`) does not yet bundle or auto-install the runtime — this is a known gap for fresh Windows 10 installs.

---

## Pre-release checklist (summary)

```
[ ] go build (windows) — clean
[ ] go build (darwin)  — clean, no regressions  
[ ] go test ./...      — all suites pass
[ ] API parity check   — all exports present on both platforms
[ ] Race audit         — queueMutex guards all wv/windowReady/messageQueue writes
[ ] Thread-affinity    — webview created + pumped on the one LockOSThread'd main thread
[ ] Lifecycle audit    — __finicky_ready bound, NavigationComplete drains queue
[ ] Window smoke test  — on real Windows, `--window` opens a live, clickable config UI
[ ] Dead code audit    — no unused imports, constants, or silencers
[ ] Critical paths     — all error paths have fallback or warning log
[ ] installer.iss      — InstallerSha256 updated for this build
[ ] winget manifest    — version and SHA updated
[ ] LAUNCH_MATERIALS   — demo GIF recorded, posting schedule set
```
