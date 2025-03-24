<script lang="ts">
  import type { LogEntry } from '../types';
  import LogContent from './LogContent.svelte';

  export let messageBuffer: LogEntry[] = [];

  let showDebug = true;
  let copyButtonText = 'Copy Logs';
  let isCopied = false;

  // Copy logs to clipboard
  async function copyLogs() {
    const logEntries = messageBuffer
      .filter(entry => showDebug || entry.level.toLowerCase() !== 'debug')
      .map(entry => {
        const time = new Date(entry.time).toISOString();
        const baseMessage = `[${time}] [${entry.level.padEnd(5)}] ${entry.msg}`;

        // Get all extra fields (excluding level, msg, time, error)
        const extraFields = Object.entries(entry)
          .filter(([key]: [string, any]) => !['level', 'msg', 'time', 'error'].includes(key))
          .map(([key, value]: [string, any]) => `${key}: ${value}`)
          .join(' | ');

        // Combine base message with extra fields and error if present
        const parts = [baseMessage];
        if (extraFields) parts.push(extraFields);
        if (entry.error) parts.push(`Error: ${entry.error}`);

        return parts.join(' | ');
      })
      .join('\n');

    try {
      await navigator.clipboard.writeText(logEntries);
      copyButtonText = 'Copied!';
      isCopied = true;
      setTimeout(() => {
        copyButtonText = 'Copy Logs';
        isCopied = false;
      }, 2000);
    } catch (err) {
      console.error('Failed to copy logs:', err);
    }
  }

  // Clear logs
  function clearLogs() {
    messageBuffer = [];
  }

  // Toggle debug logs
  function toggleDebug() {
    showDebug = !showDebug;
  }
</script>

<div class="log-window">
  <div class="log-header">
    <h2>Logs</h2>
    <div class="log-header-buttons">
      <button class:active={showDebug} on:click={toggleDebug}>
        {showDebug ? 'Hide Debug' : 'Show Debug'}
      </button>
      <button on:click={clearLogs}>Clear</button>
      <button class:copied={isCopied} on:click={copyLogs}>{copyButtonText}</button>
    </div>
  </div>

  <LogContent {messageBuffer} {showDebug}/>
</div>

<style>
  .log-window {
    border: 1px solid var(--border-color);
    border-radius: 8px;
    background: var(--log-bg);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .log-header {
    padding: 10px 12px;
    background: var(--log-header-bg);
    border-bottom: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;

    h2 {
      margin: 0;
      font-size: 1em;
      color: var(--text-primary);
    }

    .log-header-buttons {
      display: flex;
      gap: 8px;
    }

    button {
      padding: 6px 12px;
      border: 1px solid var(--border-color);
      border-radius: 4px;
      background: var(--button-bg);
      color: var(--text-primary);
      cursor: pointer;
      transition: all 0.2s ease;
      font-size: 0.9em;
      opacity: 0.9;

      &:hover {
        background: var(--button-hover);
        transform: translateY(-1px);
        opacity: 1;
      }

      &.active {
        background: var(--log-debug);
        color: white;
        border-color: var(--log-debug);
        opacity: 1;
      }

      &.copied {
        background: #4caf50;
        color: white;
        border-color: #43a047;
        opacity: 1;
      }
    }
  }
</style>