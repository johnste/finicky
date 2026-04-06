<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    children,
    center = false,
    title,
    description,
  }: {
    children: Snippet;
    center?: boolean;
    title?: string;
    description?: Snippet;
  } = $props();
</script>

<div class="page-container" class:center>
  {#if title}
    <div class="header-section">
      <h2>{title}</h2>
      {#if description}
        <p class="page-description">{@render description()}</p>
      {/if}
    </div>
  {/if}
  {@render children()}
</div>

<style>
  .page-container {
    max-width: 680px;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .page-container.center {
    align-items: center;
  }

  .header-section {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-bottom: 4px;
  }

  h2 {
    margin: 0;
    color: var(--text-primary);
    font-size: 1.1em;
    font-weight: 600;
    letter-spacing: -0.01em;
  }

  .page-description {
    color: var(--text-secondary);
    margin: 0;
    font-size: 0.85em;
  }

  .page-description :global(code) {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.9em;
    background: var(--inset-bg);
    padding: 1px 4px;
    border-radius: 3px;
  }
</style>
