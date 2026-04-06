<script lang="ts">
  import XIcon from "./icons/X.svelte";
  const CUSTOM = "__custom__";

  let {
    browser,
    profile,
    isCustom,
    isProfileCustom,
    installedBrowsers = [],
    profilesByBrowser = {},
    disabled = false,
    required = false,
    browserPlaceholder = "Select browser",
    onBrowserChange,
    onProfileChange,
    onRequestProfiles,
    onSave,
    onInput,
  }: {
    browser: string;
    profile: string;
    isCustom: boolean;
    isProfileCustom: boolean;
    installedBrowsers: string[];
    profilesByBrowser: Record<string, string[]>;
    disabled?: boolean;
    required?: boolean;
    browserPlaceholder?: string;
    onBrowserChange?: (browser: string, profile: string, isCustom: boolean) => void;
    onProfileChange?: (profile: string, isProfileCustom: boolean) => void;
    onRequestProfiles?: (b: string) => void;
    onSave?: () => void;
    onInput?: () => void;
  } = $props();

  function profileOptions(b: string): string[] {
    return profilesByBrowser[b] ?? [];
  }

  function handleBrowserSelect(e: Event) {
    const val = (e.target as HTMLSelectElement).value;
    if (val === CUSTOM) {
      onBrowserChange?.("", "", true);
    } else {
      if (val && profilesByBrowser[val] === undefined) {
        onRequestProfiles?.(val);
      }
      onBrowserChange?.(val, "", false);
      onSave?.();
    }
  }

  function handleBrowserInput(e: Event) {
    onBrowserChange?.((e.target as HTMLInputElement).value, profile, true);
    onInput?.();
  }

  function handleClearBrowser() {
    onBrowserChange?.("", "", false);
    onSave?.();
  }

  function handleProfileSelect(e: Event) {
    const val = (e.target as HTMLSelectElement).value;
    if (val === CUSTOM) {
      onProfileChange?.("", true);
    } else {
      onProfileChange?.(val, false);
      onSave?.();
    }
  }

  function handleProfileInput(e: Event) {
    onProfileChange?.((e.target as HTMLInputElement).value, true);
    onInput?.();
  }

  function handleClearProfile() {
    onProfileChange?.("", false);
    onSave?.();
  }
</script>

<div class="browser-select-row">
  {#if isCustom}
    <input
      class="browser-input"
      type="text"
      placeholder="e.g. Firefox"
      value={browser}
      oninput={handleBrowserInput}
      onblur={() => onSave?.()}
      {disabled}
    />
    <button class="clear-btn" onclick={handleClearBrowser} aria-label="Clear browser"><XIcon /></button>
  {:else}
    <div class="select-wrapper">
      <select
        class="browser-dropdown"
        value={browser}
        onchange={handleBrowserSelect}
        {disabled}
      >
        <option value="">{browserPlaceholder}</option>
        {#each installedBrowsers as b}
          <option value={b}>{b}</option>
        {/each}
        <option value={CUSTOM}>Custom...</option>
      </select>
    </div>
  {/if}

  {#if !isCustom && browser && profileOptions(browser).length > 0}
    {#if isProfileCustom}
      <input
        class="browser-input"
        type="text"
        placeholder="Profile name"
        value={profile}
        oninput={handleProfileInput}
        onblur={() => onSave?.()}
        {disabled}
      />
      <button class="clear-btn" onclick={handleClearProfile} aria-label="Clear profile"><XIcon /></button>
    {:else}
    <div class="select-wrapper" class:empty={required && !browser && !isCustom}>
        <select
          class="browser-dropdown"
          value={profile}
          onchange={handleProfileSelect}
          {disabled}
        >
          <option value="">No profile</option>
          {#each profileOptions(browser) as p}
            <option value={p}>{p}</option>
          {/each}
          <option value={CUSTOM}>Custom...</option>
        </select>
      </div>
    {/if}
  {/if}
</div>

<style>
  .browser-select-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .select-wrapper {
    position: relative;
    flex: 1;
    min-width: 0;
  }

  .select-wrapper::after {
    content: '';
    position: absolute;
    right: 14px;
    top: 50%;
    transform: translateY(-60%) rotate(45deg);
    width: 5px;
    height: 5px;
    border-right: 2px solid var(--text-secondary);
    border-bottom: 2px solid var(--text-secondary);
    pointer-events: none;
  }

  .select-wrapper:focus-within::after {
    border-color: var(--accent-color);
  }

  .browser-dropdown {
    width: 100%;
    padding: 7px 32px 7px 12px;
    background: var(--input-bg);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    font-size: 0.9em;
    font-family: inherit;
    cursor: pointer;
    -webkit-appearance: none;
    appearance: none;
  }

  .browser-dropdown:focus {
    outline: none;
    border-color: var(--accent-color);
  }

  .select-wrapper.empty .browser-dropdown {
    border-color: var(--log-warning);
  }

  .select-wrapper.empty::after {
    border-color: var(--log-warning);
  }

  .browser-dropdown:disabled {
    opacity: 0.6;
    cursor: default;
  }

  .select-wrapper:has(.browser-dropdown:disabled)::after {
    opacity: 0.6;
  }

  .browser-input {
    flex: 1;
    min-width: 0;
    padding: 8px 12px;
    background: var(--input-bg);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    font-size: 0.9em;
    font-family: inherit;
  }

  .browser-input:focus {
    outline: none;
    border-color: var(--accent-color);
  }

  .browser-input:disabled {
    opacity: 0.6;
    cursor: default;
  }

  .clear-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 3px;
    border-radius: 4px;
    flex-shrink: 0;
    display: flex;
    transition: color 0.15s;
  }

  .clear-btn:hover {
    color: var(--log-error);
  }
</style>
