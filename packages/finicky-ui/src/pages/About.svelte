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
  <div class="pixel-bg"></div>

  <div class="about-content">
    <div class="logo-container">
      <picture>
        <source
          media="(prefers-color-scheme: dark)"
          srcset="/finicky-logo.png"
        />
        <img
          alt="Finicky icon"
          src="/finicky-logo-light.png"
          class="icon"
        />
      </picture>
      <p class="version">Version {version}</p>
    </div>

    <div class="section">
      <h2>About</h2>
      <p>Finicky is a macOS application that lets you set up rules to decide which browser to open for every link.</p>
      <a href="https://github.com/johnste/finicky" target="_blank" rel="noopener noreferrer" class="repo-link">
        View on GitHub <ExternalIcon />
      </a>
    </div>

    <div class="section">
      <h2>Credits</h2>
      <p>Created by <a href="https://github.com/johnste" target="_blank" rel="noopener noreferrer">John Sterling</a></p>
      <p>Icon designed by <a href="https://github.com/uetchy" target="_blank" rel="noopener noreferrer">@uetchy</a></p>
      <p><a href="https://github.com/johnste/finicky/graphs/contributors" target="_blank" rel="noopener noreferrer">View all contributors</a></p>
    </div>

    <div class="section">
      <h2>Support Development</h2>
      <p>If you find Finicky useful, consider <a href="https://github.com/sponsors/johnste" target="_blank" rel="noopener noreferrer" class="sponsor-link">sponsoring the project</a>.</p>
    </div>
  </div>
</PageContainer>

<style>
  .pixel-bg {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    z-index: 0;
    opacity: 0.03;
    background-image:
      repeating-linear-gradient(
        0deg,
        transparent,
        transparent 8px,
        rgba(0, 0, 0, 0.1) 8px,
        rgba(0, 0, 0, 0.1) 9px
      ),
      repeating-linear-gradient(
        90deg,
        transparent,
        transparent 8px,
        rgba(0, 0, 0, 0.1) 8px,
        rgba(0, 0, 0, 0.1) 9px
      );
    background-size: 9px 9px;
  }

  .about-content {
    position: relative;
    z-index: 1;
    max-width: 600px;
  }

  .logo-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 32px;
  }

  .icon {
    height: 64px;
    opacity: 0.9;
    image-rendering: pixelated;
    image-rendering: -moz-crisp-edges;
    image-rendering: crisp-edges;
    margin-bottom: 8px;
  }

  .version {
    color: var(--text-secondary);
    font-size: 0.95em;
    margin: 0;
  }

  .section {
    margin-bottom: 32px;
    padding: 20px;
    background: rgba(0, 0, 0, 0.1);
    border-radius: 8px;
    transition: background 0.2s ease;
  }

  .section:hover {
    background: rgba(0, 0, 0, 0.15);
  }

  .section h2 {
    font-size: 1.2em;
    font-weight: 500;
    margin: 0 0 12px 0;
    color: var(--text-primary);
  }

  .section p {
    margin: 8px 0;
    line-height: 1.6;
    color: var(--text-secondary);
  }

  .repo-link {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--accent-color);
    text-decoration: none;
    font-size: 0.95em;
    margin-top: 8px;
    transition: opacity 0.2s ease;
  }

  .repo-link:hover {
    opacity: 0.8;
  }

  .repo-link svg {
    opacity: 0.7;
    width: 14px;
    height: 14px;
  }

  a {
    color: var(--accent-color);
    text-decoration: none;
    transition: opacity 0.2s ease;
  }

  a:hover {
    opacity: 0.8;
  }

  .sponsor-link {
    font-weight: 500;
  }
</style>
