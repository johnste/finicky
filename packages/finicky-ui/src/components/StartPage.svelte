<script lang="ts">
  import { Link } from "svelte-routing";
  import type { UpdateInfo, ConfigInfo } from "../types";

  export let hasConfig: boolean;
  export let numErrors: number;
  export let config: ConfigInfo;
  export let updateInfo: UpdateInfo | null;
</script>


<div class="status-section">
  {#if !hasConfig}
    <div class="status-card warning">
      <h3>No Configuration Found</h3>
      <p>Create a configuration file to customize your browser behavior.<br />
        <a
          href="https://github.com/johnste/finicky/wiki/Getting-started"
          target="_blank"
        >
          Learn how to get started
        </a>
      </p>
    </div>
  {/if}

  {#if hasConfig && (config.keepRunning !== undefined || config.hideIcon !== undefined || config.logRequests !== undefined || config.checkForUpdate !== undefined)}
    <div class="config-options">
      <h3>Configuration</h3>
      <p class="options-description">Current settings from your configuration file</p>
      <div class="options-grid">
        <div class="option-row">
          <div class="option-info">
            <div class="option-text">
              <span class="option-label">Keep Running</span>
              <span class="option-hint">App stays open after handling links</span>
            </div>
            <label class="toggle">
              <input type="checkbox" checked={config.keepRunning ?? true} disabled />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="option-row">
          <div class="option-info">
            <div class="option-text">
              <span class="option-label">Hide Icon</span>
              <span class="option-hint">Hide menu bar icon</span>
            </div>
            <label class="toggle">
              <input type="checkbox" checked={config.hideIcon ?? false} disabled />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="option-row">
          <div class="option-info">
            <div class="option-text">
              <span class="option-label">Log Requests</span>
              <span class="option-hint">Log all URL handling to file</span>
            </div>
            <label class="toggle">
              <input type="checkbox" checked={config.logRequests ?? false} disabled />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="option-row">
          <div class="option-info">
            <div class="option-text">
              <span class="option-label">Check For Updates</span>
              <span class="option-hint">Automatically check for new versions</span>
            </div>
            <label class="toggle">
              <input type="checkbox" checked={config.checkForUpdate ?? true} disabled />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
      </div>
    </div>
  {/if}



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

</div>


<style>


  .status-section {
    max-width: 700px;
    margin: 0 auto 32px;
    display: flex;
    flex-direction: column;
    gap: 16px;
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

  .success h3::before {
    background: #4caf50;
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

  .config-options {
    margin: 0 0 16px 0;
  }

  .config-options h3 {
    margin: 0 0 8px 0;
    font-size: 1.1em;
    color: var(--text-primary);
    padding-left: 12px;
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
    background-color: #ccc;
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
    background-color: white;
    transition: 0.3s;
    border-radius: 50%;
  }

  .toggle input:checked + .toggle-slider {
    background-color: #4caf50;
  }

  .toggle input:checked + .toggle-slider:before {
    transform: translateX(20px);
  }
</style>
