<script lang="ts">
  import PageContainer from "../components/PageContainer.svelte";
  import type { LogEntry } from "../types";
  import LogContent from "../components/LogContent.svelte";
  import { toast } from "../lib/toast";

  let {
    messageBuffer = [],
    onClearLogs,
  }: {
    messageBuffer: LogEntry[];
    onClearLogs: () => void;
  } = $props();

  let showDebug = $state(localStorage.getItem("showDebugLogs") === "true");

  // Copy logs to clipboard
  async function copyLogs() {
    const logEntries = messageBuffer
      .filter((entry) => showDebug || entry.level.toLowerCase() !== "debug")
      .map((entry) => {
        const time = new Date(entry.time).toISOString();
        const baseMessage = `[${time}] [${entry.level.padEnd(5)}] ${entry.msg}`;

        // Get all extra fields (excluding level, msg, time, error)
        const extraFields = Object.entries(entry)
          .filter(
            ([key]: [string, any]) =>
              !["level", "msg", "time", "error"].includes(key)
          )
          .map(([key, value]: [string, any]) => `${key}: ${value}`)
          .join(" | ");

        // Combine base message with extra fields and error if present
        const parts = [baseMessage];
        if (extraFields) parts.push(extraFields);
        if (entry.error) parts.push(`Error: ${entry.error}`);

        return parts.join(" | ");
      })
      .join("\n");

    try {
      await navigator.clipboard.writeText(logEntries);
      toast.success("Logs copied to clipboard");
    } catch (err) {
      console.error("Failed to copy logs:", err);
      toast.error("Failed to copy logs", err instanceof Error ? err.message : (err?.toString() ?? 'unknown error'));
    }
  }

  // Clear logs
  function clearLogs() {
    onClearLogs();
  }
</script>

<PageContainer>
  <div class="log-header">
    <h2>Logs</h2>
    <div class="log-header-buttons">
      <label class="checkbox-label">
        <input type="checkbox" bind:checked={showDebug} onchange={() => localStorage.setItem("showDebugLogs", String(showDebug))} />
        <span>Show debug logs</span>
      </label>
      <button onclick={clearLogs}>Clear</button>
      <button onclick={copyLogs}>Copy logs</button>
    </div>
  </div>

  <LogContent {messageBuffer} {showDebug} />
</PageContainer>

<style>
  .log-header {
    position: sticky;
    top: 2px;
    z-index: 10;
    background: var(--log-header-bg);
    border-radius: 8px;
    padding: 12px;
    margin: -18px 0 0 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  h2 {
    color: var(--text-primary);
    font-size: 1.1em;
  }

  .log-header-buttons {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: flex-end;
  }

  button {
    padding: 8px 16px;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    background: var(--button-bg);
    color: var(--text-primary);
    cursor: pointer;
    transition: all 0.2s ease;
    font-size: 0.9em;
    opacity: 0.9;
    font-weight: 500;
  }

  button:hover {
    background: var(--button-hover);
    transform: translateY(-1px);
    opacity: 1;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  button.active {
    background: var(--log-debug);
    color: white;
    border-color: var(--log-debug);
    opacity: 1;
  }

  button.copied {
    background: var(--log-success);
    color: white;
    border-color: var(--log-success);
    opacity: 1;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    padding: 8px 12px;
    border-radius: 6px;
    transition: background 0.2s ease;
    user-select: none;
  }

  .checkbox-label:hover {
    background: var(--button-hover);
  }

  .checkbox-label input[type="checkbox"] {
    cursor: pointer;
    width: 16px;
    height: 16px;
  }

  .checkbox-label span {
    color: var(--text-primary);
    font-size: 0.9em;
    font-weight: 500;
  }
</style>
