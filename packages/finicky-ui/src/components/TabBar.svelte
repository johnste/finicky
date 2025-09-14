<script lang="ts">
  import { Link } from "svelte-routing";
  import General from "./icons/General.svelte";
  import Troubleshoot from "./icons/Troubleshoot.svelte";
  import About from "./icons/About.svelte";

  export let numErrors: number = 0;

  const tabs = [
    {
      path: "/",
      label: "General",
      component: General,
    },
    {
      path: "/troubleshoot",
      label: "Troubleshoot",
      showErrors: true,
      component: Troubleshoot,
    },
    {
      path: "/about",
      label: "About",
      component: About,
    },
  ];
</script>

<div class="tab-container">
  <nav class="tab-bar">
    {#each tabs as tab}
      <Link to={tab.path} let:active>
        <div class:active class="tab-content">
          <div class="icon-container">
            <svelte:component this={tab.component} />
            {#if tab.showErrors && numErrors > 0}
              <div class="error-badge">{numErrors}</div>
            {/if}
          </div>
          <span>
            {tab.label}
          </span>
        </div>
      </Link>
    {/each}
  </nav>
</div>

<style>
  @property --gradient-start {
    syntax: "<color>";
    initial-value: rgba(98, 98, 98, 0.3);
    inherits: false;
  }

  @property --gradient-end {
    syntax: "<color>";
    initial-value: rgba(111, 111, 111, 0.3);
    inherits: false;
  }

  .tab-container {
    
    width: 190px;
    background: var(--bg-primary);
    position: sticky;
    top: 0;
    z-index: 10;
    padding: 1rem 0.5rem;
    display: flex;
    flex: 0 0 auto;
    flex-direction: column;
    align-items: stretch;
    justify-content: flex-start;
    border-right: 1px solid var(--border-color);
    border-bottom: none;
  }

  .tab-bar {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    justify-content: flex-start;
    background: transparent;
    padding: 0;
    margin: 0;
    :global(a) {
      text-decoration: none;
    }
  }

  .tab-content {
    --gradient-start: rgba(98, 98, 98, 0.1);
    --gradient-end: rgba(111, 111, 111, 0.1);

    user-select: none;
    padding: 0.75rem 1rem;
    border-radius: 0.5rem;
    display: flex;
    align-items: flex-start;
    flex-direction: row;
    gap: 12px;
    color: var(--text-primary);
    background: radial-gradient(
      circle at center,
      var(--gradient-start),
      var(--gradient-end)
    );
    background-size: 200% 200%;
    transition:
      --gradient-start 0.3s ease,
      --gradient-end 0.3s ease;
    animation: gradient 3s ease 1;
    width: 100%;
    text-align: left;

    &:hover {
      --gradient-start: rgba(98, 98, 98, 0.4);
      --gradient-end: rgba(111, 111, 111, 0.4);
    }

    &.active {
      --gradient-start: rgba(180, 84, 255, 0.3);
      --gradient-end: rgba(65, 50, 255, 0.3);
    }
  }

  @keyframes gradient {
    0% {
      background-position: 0% 50%;
    }
    50% {
      background-position: 100% 50%;
    }
    100% {
      background-position: 0% 50%;
    }
  }

  .icon-container {
    position: relative;
    display: flex;
    align-items: flex-start;
    justify-content: flex-start;
  }

  .error-badge {
    position: absolute;
    top: -8px;
    right: -8px;
    background: var(--log-error);
    color: white;
    border-radius: 50%;
    min-width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: bold;
    padding: 0 4px;
    border: 2px solid var(--bg-primary);
  }
</style>
