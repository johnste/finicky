<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    children,
    text,
    block = false,
  }: {
    children: Snippet;
    text: string;
    block?: boolean;
  } = $props();
</script>

<span class="tooltip-anchor" class:block>
  {@render children()}
  <span class="tooltip-text">{text}</span>
</span>

<style>
  .tooltip-anchor {
    position: relative;
    display: inline-flex;
  }

  .tooltip-anchor.block {
    display: flex;
  }

  .tooltip-text {
    display: none;
    position: absolute;
    bottom: calc(100% + 6px);
    right: 0;
    background: var(--card-bg, #222);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    font-size: 0.78rem;
    font-family: inherit;
    white-space: normal;
    min-width: 160px;
    max-width: 200px;
    line-height: 1.4;
    padding: 5px 8px;
    border-radius: 5px;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
    pointer-events: none;
    z-index: 100;
  }

  .tooltip-anchor:hover .tooltip-text {
    display: block;
  }
</style>
