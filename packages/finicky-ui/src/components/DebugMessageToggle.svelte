<script lang="ts">
  import type { LogEntry } from "../types";
  import { onMount, onDestroy } from "svelte";

  export let onAddMessage: (message: LogEntry) => void;

  let isEnabled = false;
  let intervalId: ReturnType<typeof setInterval> | undefined;

  // Load state from sessionStorage
  onMount(() => {
    isEnabled = sessionStorage.getItem("debugMode") === "true";
    if (isEnabled) {
      startAddingLogs();
    }
  });

  // Clean up interval on component destroy
  onDestroy(() => {
    if (intervalId) {
      clearInterval(intervalId);
    }
  });

  function startAddingLogs() {
    // Add initial set of logs
    addDebugMessages();

    // Set up interval to add logs every 5 seconds
    intervalId = setInterval(addDebugMessages, 5000);
  }

  function stopAddingLogs() {
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = undefined;
    }
  }

  function toggleDebugMode() {
    isEnabled = !isEnabled;
    sessionStorage.setItem("debugMode", isEnabled.toString());

    if (isEnabled) {
      startAddingLogs();
    } else {
      stopAddingLogs();
    }
  }

  async function addDebugMessages() {
    // Debug message with extra properties
    onAddMessage({
      level: "debug",
      msg: "Initializing application components",
      time: new Date().toISOString(),
      component: "App",
      duration: "45ms",
      tags: ["startup", "performance"],
    });

    // Add a small delay to emulate a real-world scenario
    await new Promise((resolve) => setTimeout(resolve, 100));

    // Info message (basic)
    onAddMessage({
      level: "info",
      msg: "Application started successfully https://www.example.com/path/to/resource",
      time: new Date().toISOString(),
      url: "https://www.example.com",
    });

    await new Promise((resolve) => setTimeout(resolve, 100));

    // Warning message with one extra property
    onAddMessage({
      level: "warn",
      msg: "Configuration file not found, using defaults",
      time: new Date().toISOString(),
      file: "config.json",
    });

    await new Promise((resolve) => setTimeout(resolve, 100));

    // Error message (basic)
    onAddMessage({
      level: "error",
      msg: "Failed to connect to database",
      time: new Date().toISOString(),
    });
  }
</script>

<label class="debug-toggle">
  <input type="checkbox" checked={isEnabled} on:change={toggleDebugMode} />
  <span>Auto-add Debug Messages</span>
</label>

<style>
  .debug-toggle {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 12px;
    background: var(--button-bg);
    color: var(--text-secondary);
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9em;
    transition: all 0.2s ease;
    border: 1px solid var(--border-color);
    height: 32px;
  }

  .debug-toggle:hover {
    background: var(--button-hover);
  }

  .debug-toggle input[type="checkbox"] {
    width: 14px;
    height: 14px;
    margin: 0;
    cursor: pointer;
    accent-color: var(--log-debug);
  }

  .debug-toggle span {
    user-select: none;
  }
</style>
