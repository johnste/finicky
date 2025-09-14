<script lang="ts">
  import { Link } from "svelte-routing";
  import type { UpdateInfo } from "../types";
  // TODO
  // import Configuration from "./Configuration.svelte";

  export let hasConfig: boolean;
  export let numErrors: number;
  export let configPath: string;
  export let updateInfo: UpdateInfo | null;
</script>


<div class="status-section">
  {#if hasConfig}
    <div class="status-card success">
      <h3>Loaded Configuration</h3>
        <p>Config Path: {configPath || "Not set"}</p>
    </div>
  {:else}
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
    {:else if updateInfo.updateCheckEnabled}
      <div class="status-card success">
        <h3>Finicky is up to date</h3>
      </div>
    {:else}
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
    max-width: 600px;
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
</style>
