<script lang="ts">
  import { Link } from "svelte-routing";
  import PageContainer from "../components/PageContainer.svelte";
  import type { UpdateInfo, ConfigInfo } from "../types";
  import ExternalIcon from "../components/icons/External.svelte";

  export let hasConfig: boolean;
  export let numErrors: number;
  export let config: ConfigInfo;
  export let updateInfo: UpdateInfo | null;
</script>

<PageContainer
  title={hasConfig ? "Configuration" : "No Configuration Found"}
  description={hasConfig
    ? "Current settings from your configuration file"
    : undefined}
>
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

  <div class="config-options">
    <div class="options-grid">
      <div class="option-row">
        <div class="option-info">
          <div class="option-text">
            <span class="option-label">Keep running</span>
            <span class="option-hint">App stays open in the background</span>
          </div>
          <label class="toggle">
            <input
              type="checkbox"
              checked={config.options?.keepRunning ?? true}
              disabled
            />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>
      <div class="option-row">
        <div class="option-info">
          <div class="option-text">
            <span class="option-label">Hide icon</span>
            <span class="option-hint">Hide menu bar icon</span>
          </div>
          <label class="toggle">
            <input
              type="checkbox"
              checked={config.options?.hideIcon ?? false}
              disabled
            />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>
      <div class="option-row">
        <div class="option-info">
          <div class="option-text">
            <span class="option-label">Log requests</span>
            <span class="option-hint">Log all URL handling to file</span>
          </div>
          <label class="toggle">
            <input
              type="checkbox"
              checked={config.options?.logRequests ?? false}
              disabled
            />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>
      <div class="option-row">
        <div class="option-info">
          <div class="option-text">
            <span class="option-label">Check for updates</span>
            <span class="option-hint">Automatically check for new versions</span
            >
          </div>
          <label class="toggle">
            <input
              type="checkbox"
              checked={config.options?.checkForUpdates ?? true}
              disabled
            />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>
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
        <h3>New Version Available</h3>
        <p>
          A new version "{updateInfo.version}" of Finicky is available to
          download.
        </p>
        <p>
          <a href={updateInfo.releaseUrl} target="_blank">
            View release notes
          </a>
          <br />
          <a href={updateInfo.downloadUrl} target="_blank">
            Download the latest version
          </a>
        </p>
      </div>
    {:else if !updateInfo.updateCheckEnabled}
      <div class="status-card info">
        <h3>Update check is disabled</h3>
        <a href="https://github.com/johnste/finicky/releases" target="_blank">
          Check releases
        </a>
      </div>
    {/if}
  {/if}

  <!-- <Configuration /> -->
</PageContainer>

<style>
  .status-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 20px;
    border-radius: 12px;
    text-align: left;
    background: var(--log-bg);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
    transition:
      transform 0.2s ease,
      box-shadow 0.2s ease;
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

  .warning h3::before {
    background: #ffc107;
  }

  .error h3::before {
    background: #f44336;
  }

  .info h3::before {
    background: #2196f3;
  }

  .external-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .external-link svg {
    opacity: 0.7;
  }

  .no-config-message {
    margin: -12px 0 0 12px;
    color: var(--text-secondary);
    font-size: 0.9em;
    opacity: 0.8;
  }

  .config-options {
    margin: 0;
  }

  .options-description {
    color: var(--text-secondary);
    font-size: 0.9em;
    margin: -4px 0 12px 12px;
    opacity: 0.8;
  }

  .options-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .option-row {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px;
    background: rgba(0, 0, 0, 0.1);
    border-radius: 8px;
    transition: background 0.2s ease;
  }

  .option-row:hover {
    background: rgba(0, 0, 0, 0.15);
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

  /* Toggle Switch Styles */
  .toggle {
    position: relative;
    display: inline-block;
    width: 44px;
    height: 24px;
    cursor: not-allowed;
  }

  .toggle input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .toggle-slider {
    position: absolute;
    cursor: not-allowed;
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
    background-color: var(--log-success);
  }

  .toggle input:checked + .toggle-slider:before {
    transform: translateX(20px);
    background-color: white;
  }
</style>
