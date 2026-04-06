<script lang="ts">
  import { Link } from "svelte-routing";
  import PreferencesIcon from "./icons/Preferences.svelte";
  import TestIcon from "./icons/Test.svelte";
  import LogsIcon from "./icons/Logs.svelte";
  import AboutIcon from "./icons/About.svelte";
  import RulesIcon from "./icons/Rules.svelte";

  export let numErrors: number = 0;

  const tabs = [
    {
      path: "/",
      label: "Preferences",
      component: PreferencesIcon,
    },
    {
      path: "/rules",
      label: "Rules",
      component: RulesIcon,
    },
    {
      path: "/test",
      label: "Test",
      component: TestIcon,
    },
  ];

  const bottomTabs = [
    {
      path: "/troubleshoot",
      label: "Logs",
      showErrors: true,
      component: LogsIcon,
    },
    {
      path: "/about",
      label: "About",
      component: AboutIcon,
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
          </div>
          <span>{tab.label}</span>
        </div>
      </Link>
    {/each}
    <div class="spacer"></div>
    <div class="bottom-tabs">
      {#each bottomTabs as tab}
        <Link to={tab.path} let:active>
          <div class:active class="tab-content tab-small">
            <div class="icon-container">
              <svelte:component this={tab.component} />
              {#if tab.showErrors && numErrors > 0}
                <div class="error-badge">{numErrors}</div>
              {/if}
            </div>
            <span>{tab.label}</span>
          </div>
        </Link>
      {/each}
    </div>
  </nav>
</div>

<style>
  .tab-container {
    width: 210px;
    background: var(--bg-nav);
    position: sticky;
    top: 0;
    z-index: 10;
    padding: 0.75rem 0.625rem;
    display: flex;
    flex: 0 0 auto;
    flex-direction: column;
    align-items: stretch;
    justify-content: flex-start;
    border-right: 1px solid var(--border-color);
    height: 100%;
    box-sizing: border-box;
  }

  .tab-bar {
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex: 1;
    background: transparent;
    padding: 0;
    margin: 0;
    :global(a) {
      text-decoration: none;
    }
  }

  .spacer {
    flex: 1;
  }

  .bottom-tabs {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding-top: 6px;
    border-top: 1px solid var(--border-color);
  }

  .tab-content {
    user-select: none;
    padding: 0.6rem 0.875rem;
    border-radius: 0.5rem;
    display: flex;
    align-items: center;
    flex-direction: row;
    gap: 10px;
    color: var(--nav-text);
    background: transparent;
    transition: background 0.1s ease;
    width: 100%;
    text-align: left;
    font-size: 0.925em;
    font-weight: 400;

    &:hover {
      background: var(--nav-hover);
    }

    &.active {
      background: var(--nav-active);
      color: var(--nav-active-text);
      font-weight: 500;
    }

    &.tab-small {
      font-size: 0.78em;
      padding: 0.35rem 0.75rem;
      color: var(--nav-text-secondary);

      &.active {
        color: var(--nav-active-text);
      }
    }
  }

  .icon-container {
    position: relative;
    display: flex;
    align-items: center;
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
    border: 2px solid var(--bg-nav);
  }
</style>
