<script lang="ts">
  import { Link } from "svelte-routing";
  import { onDestroy } from "svelte";
  import PageContainer from "../components/PageContainer.svelte";
  import BrowserProfileSelector from "../components/BrowserProfileSelector.svelte";
  import OptionRow from "../components/OptionRow.svelte";
  import type { UpdateInfo, ConfigInfo, RulesFile } from "../types";
  import ExternalIcon from "../components/icons/External.svelte";
  import LockIcon from "../components/icons/Lock.svelte";
  import WarningIcon from "../components/icons/Warning.svelte";
  import Tooltip from "../components/Tooltip.svelte";
  import { toast } from "../lib/toast";

  function onLockedClick() {
    toast.show("JS config is active", "info", "These settings are managed by your JavaScript config file and can't be changed here.");
  }

  export let hasConfig: boolean;
  export let numErrors: number;
  export let config: ConfigInfo;
  export let updateInfo: UpdateInfo | null;
  export let rulesFile: RulesFile;
  export let isJSConfig: boolean;
  export let installedBrowsers: string[] = [];
  export let profilesByBrowser: Record<string, string[]> = {};

  const SAVE_DEBOUNCE = 500;
  let saveTimer: ReturnType<typeof setTimeout>;

  // Local editable state, seeded from rules.json overrides then JS config then defaults
  let keepRunning = rulesFile.options?.keepRunning ?? config.options?.keepRunning ?? true;
  let hideIcon = rulesFile.options?.hideIcon ?? config.options?.hideIcon ?? false;
  let logRequests = rulesFile.options?.logRequests ?? config.options?.logRequests ?? false;
  let checkForUpdates = rulesFile.options?.checkForUpdates ?? config.options?.checkForUpdates ?? true;

  const SAFARI = "Safari";

  let defaultBrowser = isJSConfig ? (config.defaultBrowser ?? "") : (rulesFile.defaultBrowser || SAFARI);
  let defaultProfile = rulesFile.defaultProfile ?? "";
  let defaultBrowserIsCustom = false;
  let defaultProfileIsCustom = false;

  // Sync when the rulesFile prop is updated from the backend.
  // Does NOT read profilesByBrowser — keeps profile fetches from resetting local edits.
  $: {
    keepRunning = rulesFile.options?.keepRunning ?? config.options?.keepRunning ?? true;
    hideIcon = rulesFile.options?.hideIcon ?? config.options?.hideIcon ?? false;
    logRequests = rulesFile.options?.logRequests ?? config.options?.logRequests ?? false;
    checkForUpdates = rulesFile.options?.checkForUpdates ?? config.options?.checkForUpdates ?? true;
    defaultBrowser = isJSConfig ? (config.defaultBrowser ?? "") : (rulesFile.defaultBrowser || SAFARI);
    defaultProfile = rulesFile.defaultProfile ?? "";
  }
  // Separate statements so browser/profile list updates only affect these derived
  // flags, not defaultBrowser/defaultProfile values above.
  $: defaultBrowserIsCustom = defaultBrowser !== "" && !installedBrowsers.includes(defaultBrowser);
  $: defaultProfileIsCustom = defaultProfile !== "" && !(profilesByBrowser[defaultBrowser] ?? []).includes(defaultProfile);

  function save() {
    if (isJSConfig) return;
    clearTimeout(saveTimer);
    window.finicky.sendMessage({
      type: "saveRules",
      payload: {
        ...rulesFile,
        defaultBrowser,
        defaultProfile,
        options: { keepRunning, hideIcon, logRequests, checkForUpdates },
      },
    });
  }

  function scheduleSave() {
    if (isJSConfig) return;
    clearTimeout(saveTimer);
    saveTimer = setTimeout(() => {
      window.finicky.sendMessage({
        type: "saveRules",
        payload: {
          ...rulesFile,
          defaultBrowser,
          defaultProfile,
          options: { keepRunning, hideIcon, logRequests, checkForUpdates },
        },
      });
    }, SAVE_DEBOUNCE);
  }

  onDestroy(() => {
    clearTimeout(saveTimer);
  });

</script>

<PageContainer title={hasConfig ? "Configuration" : "No Configuration Found"}>
  {#if hasConfig}
    {#snippet description()}Current settings from your configuration file{/snippet}
  {/if}
  {#if !hasConfig}
    <div class="no-config-message">
      <a
        href="https://github.com/johnste/finicky/wiki/Getting-started"
        target="_blank"
        rel="noopener noreferrer"
        class="external-link"
      >
        Learn how to get started
        <ExternalIcon />
      </a>
    </div>
  {/if}

  <!-- Default browser -->
  <div class="section" class:readonly={isJSConfig}>
    <div class="section-header">
      <span class="section-label">Default browser</span>
      {#if isJSConfig}
        <Tooltip text="JS config is active — these settings can't be changed here">
          <span class="lock-inline"><LockIcon /></span>
        </Tooltip>
      {:else}
        <span class="section-hint">Used when no rule matches</span>
      {/if}
    </div>
    <BrowserProfileSelector
      browser={defaultBrowser}
      profile={defaultProfile}
      isCustom={defaultBrowserIsCustom}
      isProfileCustom={defaultProfileIsCustom}
      {installedBrowsers}
      {profilesByBrowser}
      disabled={isJSConfig}
      browserPlaceholder="System default"
      onBrowserChange={(browser, profile, isCustom) => {
        defaultBrowser = browser;
        defaultProfile = profile;
        defaultBrowserIsCustom = isCustom;
        defaultProfileIsCustom = false;
      }}
      onProfileChange={(profile, isProfileCustom) => {
        defaultProfile = profile;
        defaultProfileIsCustom = isProfileCustom;
      }}
      onRequestProfiles={(b) => window.finicky.sendMessage({ type: "getBrowserProfiles", browser: b })}
      onSave={save}
      onInput={scheduleSave}
    />
  </div>

  <div class="config-options">
    <div class="options-grid">
      <OptionRow
        label="Keep running"
        hint="App stays open in the background"
        bind:checked={keepRunning}
        locked={isJSConfig}
        onLockedClick={onLockedClick}
        onchange={scheduleSave}
      />
      <OptionRow
        label="Hide icon"
        hint="Hide menu bar icon"
        bind:checked={hideIcon}
        locked={isJSConfig}
        onLockedClick={onLockedClick}
        onchange={scheduleSave}
      />
      <OptionRow
        label="Log requests"
        hint="Log all URL handling to file"
        bind:checked={logRequests}
        locked={isJSConfig}
        onLockedClick={onLockedClick}
        onchange={scheduleSave}
      />
      <OptionRow
        label="Check for updates"
        hint="Automatically check for new versions"
        bind:checked={checkForUpdates}
        locked={isJSConfig}
        onLockedClick={onLockedClick}
        onchange={scheduleSave}
      />
    </div>
  </div>

  {#if numErrors > 0}
    <div class="status-card error">
      <h3>Errors</h3>
      <p>
        Finicky found {numErrors} errors while evaluating your configuration.
      </p>
      <Link to="/troubleshoot">Troubleshooting</Link>
    </div>
  {/if}

  {#if updateInfo}
    {#if updateInfo.hasUpdate}
      <div class="status-card info">
        <div class="update-header">
          <h3>New Version Available</h3>
          <span class="update-version">{updateInfo.version}</span>
        </div>
        <div class="update-actions">
          <a href={updateInfo.downloadUrl} target="_blank" rel="noopener noreferrer" class="download-btn">
            Download {updateInfo.version}
          </a>
          <a href={updateInfo.releaseUrl} target="_blank" rel="noopener noreferrer" class="release-link">
            Release notes
          </a>
        </div>
      </div>
    {:else if !updateInfo.updateCheckEnabled}
      <div class="status-card info">
        <h3>Update check is disabled</h3>
        <a href="https://github.com/johnste/finicky/releases" target="_blank">Check releases</a>
      </div>
    {/if}
  {/if}
</PageContainer>

<style>
  .section {
    display: flex;
    flex-direction: column;
    gap: 10px;
    background: var(--card-bg);
    border-radius: 12px;
    padding: 16px;
    border: 1px solid var(--card-border);
  }

  .section.readonly {
    cursor: default;
  }

.section-header {
    display: flex;
    align-items: baseline;
    gap: 10px;
  }

  .section-label {
    color: var(--text-primary);
    font-size: 0.9em;
    font-weight: 600;
  }

  .section-hint {
    color: var(--text-secondary);
    font-size: 0.82em;
    opacity: 0.7;
  }

  .lock-inline {
    display: inline-flex;
    align-items: center;
    opacity: 0.6;
    color: var(--text-secondary);
    flex-shrink: 0;
    cursor: help;
  }

  .status-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 20px;
    border-radius: 12px;
    text-align: left;
    background: var(--log-bg);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
    transition: transform 0.2s ease, box-shadow 0.2s ease;
  }

  .status-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .status-card h3 {
    margin: 0;
    font-size: 1.1em;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-card h3::before {
    content: "";
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .error h3::before { background: #f44336; }
  .info h3::before { background: #2196f3; }

  .update-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .update-version {
    font-size: 0.78em;
    color: var(--text-secondary);
    background: var(--inset-bg);
    border-radius: 4px;
    padding: 2px 6px;
  }

  .update-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .download-btn {
    display: inline-flex;
    align-items: center;
    padding: 7px 16px;
    background: var(--accent-color);
    color: #fff;
    border-radius: 8px;
    font-size: 0.88em;
    font-weight: 500;
    text-decoration: none;
    transition: opacity 0.15s;
  }

  .download-btn:hover {
    opacity: 0.85;
  }

  .release-link {
    font-size: 0.85em;
    color: var(--text-secondary);
    text-decoration: none;
    opacity: 0.7;
  }

  .release-link:hover {
    opacity: 1;
  }

  .external-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .no-config-message {
    color: var(--text-secondary);
    font-size: 0.9em;
  }

  .options-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

</style>
