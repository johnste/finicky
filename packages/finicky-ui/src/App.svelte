<script lang="ts">
  import { onMount } from "svelte";
  import { Router, Route } from "svelte-routing";
  import LogViewer from "./components/LogViewer.svelte";
  import StartPage from "./components/StartPage.svelte";
  import TabBar from "./components/TabBar.svelte";
  import About from "./components/About.svelte";
  import TestUrl from "./components/TestUrl.svelte";
  import type { LogEntry, UpdateInfo, ConfigInfo } from "./types";

  let version = "v0.0.0";
  let buildInfo = "dev";
  let isDevMode = false;

  function activateDevMode() {
    isDevMode = !isDevMode;
    localStorage.setItem("finicky-dev-mode", JSON.stringify(isDevMode));
    console.log("Dev mode toggled globally! 🚀");
  }

  // Configuration state
  let hasConfig = false;
  let config: ConfigInfo = { configPath: "" };
  // Initialize message buffer
  let messageBuffer: LogEntry[] = [];
  let updateInfo: UpdateInfo | null = null;

  // Reactive declaration to count errors in messageBuffer
  $: numErrors = messageBuffer.filter(
    (msg) => msg.level.toLowerCase() === "error"
  ).length;

  // Function to handle messages from the native app
  function handleMessage(msg: any) {
    const parsedMsg = typeof msg === "string" ? JSON.parse(msg) : msg;
    console.log("Received message:", parsedMsg.type, parsedMsg);
    switch (parsedMsg.type) {
      case "version":
        version = parsedMsg.message;
        break;
      case "buildInfo":
        buildInfo = parsedMsg.message;
        break;
      case "config":
        hasConfig = true;
        if (parsedMsg.message) {
          config = parsedMsg.message;
        }
        break;

      case "updateInfo":
        updateInfo = parsedMsg.message;
        break;
      default:
        const newMessage = parsedMsg.message
          ? JSON.parse(parsedMsg.message)
          : parsedMsg;
        messageBuffer = [...messageBuffer, newMessage];
    }
  }

  // Handle debug messages from the DebugMessageToggle
  function handleDebugMessage(message: LogEntry) {
    messageBuffer = [...messageBuffer, message];
  }

  // Clear all logs
  function clearAllLogs() {
    messageBuffer = [];
  }

  onMount(() => {
    // Load dev mode state from localStorage
    isDevMode = localStorage.getItem("finicky-dev-mode") === "true";
    if (isDevMode) {
      console.log("Dev mode restored from localStorage! 🚀");
    }

    // Set up the bridge between native app and web UI
    window.finicky = {
      sendMessage: (msg: any) => {
        window.webkit?.messageHandlers?.finicky?.postMessage(
          JSON.stringify(msg)
        );
      },
      receiveMessage: handleMessage,
    };

    // Listen for path changes
  });
</script>

<Router>
  <main>
    <div class="layout">
      <TabBar {numErrors} />
      <div class="container">
        <div class="content">
          <Route path="/">
            <StartPage
              {hasConfig}
              {updateInfo}
              {config}
              {numErrors}
            />
          </Route>

          <Route path="/troubleshoot">
            <LogViewer {messageBuffer} onClearLogs={clearAllLogs} />
          </Route>

          <Route path="/test">
            <TestUrl />
          </Route>

          <Route path="/about">
            <About
              {version}
              {isDevMode}
              toggleDevMode={activateDevMode}
              addMessage={handleDebugMessage}
            />
          </Route>
        </div>
      </div>
    </div>
    <div class="footer">
      <span class="version">{version}</span>
      {#if hasConfig}
        <span class="config-label">Loaded config:</span>
        <span class="config-path" title={config.configPath}>{config.configPath || "Not set"}</span>
      {:else}
        <span class="config-status warning">No config</span>
      {/if}
      <span class="spacer"></span>
      {#if updateInfo && updateInfo.updateCheckEnabled && !updateInfo.hasUpdate}
        <span class="up-to-date">✓ Up to date</span>
      {/if}
      {#if isDevMode}
        <span class="dev-mode">DEV MODE</span>
      {/if}
    </div>
  </main>
</Router>

<style>
  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
    position: relative;
  }

  .layout {
    display: flex;
    flex: 1 1 auto;
    min-height: 0;
  }

  .container {
    padding: 1rem;
    max-width: 100%;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    flex: 1 1 100%;
    overflow-y: auto;
  }

  .footer {
    display: flex;
    flex: 0 0 auto;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem;
    background: var(--background);
    border-top: 1px solid var(--border-color);
    overflow: hidden;
  }

  .version {
    color: var(--text-secondary);
    font-size: 0.9em;
  }

  .build-info {
    color: var(--text-secondary);
    font-size: 0.8em;
    opacity: 0.8;
  }

  .content {
    flex: 1;
  }

  .dev-mode {
    color: #b654ff;
    font-weight: bold;
    font-size: 0.8em;
    opacity: 0.8;
  }

  .spacer {
    flex: 1;
  }

  .up-to-date {
    color: #4caf50;
    font-size: 0.8em;
    opacity: 0.9;
  }

  .config-path {
    color: var(--text-secondary);
    font-size: 0.8em;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    opacity: 0.8;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 500px;
    flex-shrink: 1;
    background: rgba(0, 0, 0, 0.2);
    padding: 2px 6px;
    border-radius: 3px;
  }

  .config-label {
    color: var(--text-secondary);
    font-size: 0.8em;
    opacity: 0.8;
  }

  .config-status {
    font-size: 0.8em;
    opacity: 0.8;
  }

  .config-status.warning {
    color: #ffc107;
  }
</style>
