<script lang="ts">
  import type { LogEntry } from "../types";
  import { formatLogEntry } from "../utils/text";

  export let messageBuffer: LogEntry[] = [];
  export let showDebug = true;

  let logContent: HTMLElement;
</script>

<ol class="log-content" bind:this={logContent}>
  {#each messageBuffer as entry}
    {#if showDebug || entry.level.toLowerCase() !== "debug"}
      <li class="log-entry">
        <span class="log-time" title={entry.time}>
          {new Date(entry.time).toLocaleTimeString()}
        </span>
        <span class="log-level-icon log-level-{entry.level.toLowerCase()}"
        ></span>
        <div class="log-message log-level-{entry.level.toLowerCase()}">
          {#each formatLogEntry(entry) as part}
            {#if part.type === "url"}
              <a href={part.content} target="_blank" rel="noopener noreferrer"
                >{part.content}</a
              >
            {:else}
              {part.content}
            {/if}
          {/each}
        </div>
      </li>
    {/if}
  {/each}
</ol>

<style>
  .log-content {
    list-style: none;
    margin: 0;
    padding: 12px;
    overflow-y: auto;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
      monospace;
    line-height: 1.4;
    color: var(--text-primary);
    flex: 1;
    overflow: auto;
  }

  .log-entry {
    display: flex;
    justify-content: flex-start;
    gap: 1em;
    margin-bottom: 4px;
    align-items: flex-start;
  }

  .log-time {
    color: var(--text-secondary);
    white-space: nowrap;
    font-size: 0.9em;
  }

  .log-level-icon {
    width: 1.5em;
    text-align: center;
  }

  .log-message {
    flex-grow: 1;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .log-level-error {
    color: var(--log-error);
  }

  .log-level-warn {
    color: var(--log-warning);
  }

  .log-level-debug {
    color: var(--log-debug);
  }

  .log-level-icon.log-level-error::before {
    content: "❌";
  }

  .log-level-icon.log-level-warn::before {
    content: "⚠️";
  }

  .log-level-icon.log-level-debug::before {
    content: "🔍";
  }

  :global(.log-message a) {
    color: inherit;
    text-decoration: underline;
    text-decoration-style: dotted;
    opacity: 0.9;
  }

  :global(.log-message a:hover) {
    text-decoration-style: solid;
    opacity: 1;
  }
</style>
