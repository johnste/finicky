<script lang="ts">
  import { toast } from "../lib/toast";
  import type { Toast } from "../lib/toast";
  import { fade, fly } from "svelte/transition";
  import { flip } from "svelte/animate";

  let toasts: Toast[] = [];

  toast.subscribe((value) => {
    toasts = value;
  });
</script>

<div class="toast-container">
  {#each toasts as t (t.id)}
    <div
      class="toast toast-{t.type}"
      role="alert"
      in:fly={{ y: 50, duration: 200 }}
      out:fade={{ duration: 200 }}
      animate:flip={{ duration: 300 }}
    >
      <div class="toast-content">
        <svg class="toast-timer" viewBox="0 0 24 24">
          <circle class="timer-background" cx="12" cy="12" r="10" />
          {#key t.key}
            <circle
              class="timer-progress"
              cx="12"
              cy="12"
              r="10"
              style="animation-duration: {t.duration}ms"
            />
          {/key}
        </svg>
        <span class="toast-message">{t.message}</span>
      </div>
      {#if t.extra}
        <pre class="toast-extra">{t.extra}</pre>
      {/if}
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    bottom: 60px;
    right: 24px;
    z-index: 9999;
    display: flex;
    flex-direction: column-reverse;
    gap: 12px;
    pointer-events: none;
  }

  .toast {
    display: flex;
    flex-direction: column;
    align-items: left;

    gap: 12px;
    padding: 12px 16px;
    border-radius: 8px;
    background: var(--log-bg);
    border: 1px solid var(--border-color);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    min-width: 250px;
    max-width: 400px;
    pointer-events: auto;
    position: relative;
  }

  .toast-content {
    display: flex;
    align-items: center;
    gap: 12px;
    flex: 1;
  }

  @keyframes slideIn {
    from {
      transform: translateY(100%);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }

  .toast-timer {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
    transform: rotate(-90deg);
  }

  .timer-background {
    fill: none;
    stroke: var(--border-color);
    stroke-width: 2;
    opacity: 0.3;
  }

  .timer-progress {
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-dasharray: 62.83;
    stroke-dashoffset: 0;
    animation: circleProgress linear forwards;
    opacity: 0.6;
  }

  @keyframes circleProgress {
    to {
      stroke-dashoffset: 62.83;
    }
  }

  .toast-message {
    flex: 1;
    font-size: 0.9em;
    line-height: 1.4;
  }

  .toast-extra {
    font-family: var(--font-mono, monospace);
    background: rgba(30, 32, 36, 0.9);
    color: #e0e0e0;
    padding: 8px 10px;
    border-radius: 6px;
    font-size: 0.85em;
    overflow-x: auto;
    white-space: pre-wrap;
  }

  .toast-success {
    color: var(--log-success);
  }

  .toast-success .toast-message {
    color: var(--log-success);
  }

  .toast-error {
    color: var(--log-error);
  }

  .toast-error .toast-message {
    color: var(--log-error);
  }

  .toast-warning {
    color: var(--log-warning);
  }

  .toast-warning .toast-message {
    color: var(--log-warning);
  }

  .toast-info {
    color: var(--accent-color);
  }

  .toast-info .toast-message {
    color: var(--accent-color);
  }
</style>
