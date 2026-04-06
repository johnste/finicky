<script lang="ts">
  import LockIcon from "./icons/Lock.svelte";
  import Tooltip from "./Tooltip.svelte";

  const LOCKED_TOOLTIP = "JS config is active — these settings can't be changed here";

  let {
    label,
    hint,
    checked = $bindable(false),
    locked = false,
    onLockedClick,
    onchange,
  }: {
    label: string;
    hint: string;
    checked: boolean;
    locked?: boolean;
    onLockedClick?: () => void;
    onchange?: () => void;
  } = $props();
</script>

{#snippet inner()}
  <div class="option-info">
    <div class="option-text">
      <span class="option-label">{label}</span>
      <span class="option-hint">{hint}</span>
    </div>
    <label class="toggle" class:locked>
      <input type="checkbox" bind:checked disabled={locked} {onchange} />
      <span class="toggle-slider"></span>
      {#if locked}<span class="nub-lock"><LockIcon /></span>{/if}
    </label>
  </div>
{/snippet}

{#if locked}
  <Tooltip text={LOCKED_TOOLTIP} block>
    <button type="button" class="option-row locked" onclick={onLockedClick}>
      {@render inner()}
    </button>
  </Tooltip>
{:else}
  <div class="option-row">
    {@render inner()}
  </div>
{/if}

<style>
  .option-row {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px;
    background: var(--inset-bg);
    border-radius: 8px;
    transition: background 0.2s ease;
    width: 100%;
    text-align: left;
    border: none;
    font: inherit;
    color: inherit;
  }

  .option-row:hover {
    background: var(--bg-hover);
  }

  .option-row.locked {
    cursor: default;
  }

  .option-info {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px;
  }

  .option-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .option-label {
    color: var(--text-primary);
    font-size: 0.95em;
    font-weight: 500;
  }

  .option-hint {
    color: var(--text-secondary);
    font-size: 0.85em;
    opacity: 0.7;
  }

  .toggle {
    position: relative;
    display: inline-block;
    width: 44px;
    height: 24px;
    cursor: pointer;
    flex-shrink: 0;
  }

  .toggle.locked {
    cursor: default;
  }

  .toggle input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .toggle-slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: #666;
    transition: 0.3s;
    border-radius: 24px;
  }

  .toggle-slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: #ddd;
    transition: 0.3s;
    border-radius: 50%;
  }

  .toggle input:checked + .toggle-slider {
    background-color: var(--toggle-active);
  }

  .toggle input:checked + .toggle-slider:before {
    transform: translateX(20px);
    background-color: #111111;
  }

  .nub-lock {
    position: absolute;
    left: 3px;
    bottom: 3px;
    width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none;
    transition: transform 0.3s;
  }

  .toggle.locked input:checked ~ .nub-lock {
    transform: translateX(20px);
  }

  .nub-lock :global(svg) {
    width: 10px;
    height: 10px;
    color: #444;
    stroke-width: 2.5;
  }

  .toggle.locked input:checked ~ .nub-lock :global(svg) {
    color: #aaa;
  }
</style>
