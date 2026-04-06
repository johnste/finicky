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

  async function copyLogs() {
    const logEntries = messageBuffer
      .filter((entry) => showDebug || entry.level.toLowerCase() !== "debug")
      .map((entry) => {
        const time = new Date(entry.time).toISOString();
        const baseMessage = `[${time}] [${entry.level.padEnd(5)}] ${entry.msg}`;
        const extraFields = Object.entries(entry)
          .filter(([key]: [string, any]) => !["level", "msg", "time", "error"].includes(key))
          .map(([key, value]: [string, any]) => `${key}: ${value}`)
          .join(" | ");
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
      toast.error("Failed to copy logs", err instanceof Error ? err.message : (err?.toString() ?? "unknown error"));
    }
  }
</script>

<PageContainer title="Logs">
  {#snippet description()}
    <div class="log-controls">
      <label class="debug-toggle">
        <input
          type="checkbox"
          bind:checked={showDebug}
          onchange={() => localStorage.setItem("showDebugLogs", String(showDebug))}
        />
        Show debug
      </label>
      <button onclick={onClearLogs}>Clear</button>
      <button onclick={copyLogs}>Copy</button>
    </div>
  {/snippet}

  <LogContent {messageBuffer} {showDebug} />
</PageContainer>

<style>
  .log-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .debug-toggle {
    display: flex;
    align-items: center;
    gap: 5px;
    cursor: pointer;
    font-size: 0.85em;
    color: var(--text-secondary);
    user-select: none;
  }

  .debug-toggle input[type="checkbox"] {
    cursor: pointer;
    accent-color: var(--accent-color);
  }

  button {
    padding: 3px 10px;
    border: 1px solid var(--border-color);
    border-radius: 5px;
    background: var(--button-bg);
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 0.82em;
    font-family: inherit;
    transition: color 0.15s;
  }

  button:hover {
    color: var(--text-primary);
  }
</style>
