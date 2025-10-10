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
    description?: string;
  } = $props();
</script>

<div class="page-container" class:center>
  {#if title}
    <div class="header-section">
      <h2>{title}</h2>
      {#if description}
        <p class="page-description">{description}</p>
      {/if}
    </div>
  {/if}
  {@render children()}
</div>

<style>
  .page-container {
    max-width: 700px;
    margin: 0 auto;
    padding: 0px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .page-container.center {
    align-items: center;
  }

  .header-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  h2 {
    margin: 0 0 8px 0;
    padding-left: 12px;
    color: var(--text-primary);
    font-size: 1.1em;
  }

  .page-description {
    color: var(--text-secondary);
    margin: -4px 0 0 12px;
    font-size: 0.9em;
    opacity: 0.8;
  }
</style>
