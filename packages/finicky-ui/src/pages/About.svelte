<script lang="ts">
  import PageContainer from "../components/PageContainer.svelte";
  import ExternalIcon from "../components/icons/External.svelte";

  let {
    version,
  }: {
    version: string;
  } = $props();
</script>

<PageContainer>
  <div class="logo-section">
    <div class="app-info">
      <span class="app-name">Finicky</span>
      <span class="version">Version {version}</span>
    </div>
  </div>

  <div class="section">
    <p>
      Finicky is a macOS application that lets you set up rules to decide which
      browser to open for every link.
    </p>
    <a
      href="https://github.com/johnste/finicky"
      target="_blank"
      rel="noopener noreferrer"
      class="repo-link"
    >
      View on GitHub <ExternalIcon />
    </a>
  </div>

  <div class="section">
    <h2>Credits</h2>
    <p>
      Created by <a
        href="https://github.com/johnste"
        target="_blank"
        rel="noopener noreferrer">John Sterling</a
      >
    </p>
    <p>
      Icon designed by <a
        href="https://github.com/uetchy"
        target="_blank"
        rel="noopener noreferrer">@uetchy</a
      >
    </p>
    <p>
      <a
        href="https://github.com/johnste/finicky/graphs/contributors"
        target="_blank"
        rel="noopener noreferrer">View all contributors</a
      >
    </p>
  </div>

  <div class="section">
    <p>
      If you find Finicky useful, consider <a
        href="https://github.com/sponsors/johnste"
        target="_blank"
        rel="noopener noreferrer">sponsoring the project</a
      >.
    </p>
  </div>
</PageContainer>

<style>
  .logo-section {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 4px 0;
  }

  .app-info {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .app-name {
    font-size: 1.4em;
    font-weight: 600;
    color: var(--text-primary);
    line-height: 1.2;
  }

  .version {
    color: var(--text-secondary);
    font-size: 0.85em;
  }

  .section {
    padding: 16px 20px;
    background: var(--card-bg);
    border: 1px solid var(--card-border);
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .section h2 {
    font-size: 0.8em;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-secondary);
    margin: 0 0 6px 0;
  }

  .section p {
    margin: 0;
    line-height: 1.6;
    color: var(--text-secondary);
    font-size: 0.9em;
  }

  .repo-link {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    color: var(--accent-color);
    text-decoration: none;
    font-size: 0.9em;
    margin-top: 6px;
  }

  .repo-link:hover {
    opacity: 0.8;
  }

  a {
    color: var(--accent-color);
    text-decoration: none;
  }

  a:hover {
    opacity: 0.8;
  }
</style>
