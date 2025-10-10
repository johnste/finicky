<script lang="ts">
  let {
    version,
  }: {
    version: string;
  } = $props();
</script>

<div class="page-container">
  <div class="icon-section">
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
    <p class="version">{version}</p>
  </div>

  <div class="info-card">
    <a href="https://github.com/johnste/finicky" target="_blank" rel="noopener noreferrer" class="external-link">
      github.com/johnste/finicky
      <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
        <polyline points="15 3 21 3 21 9"/>
        <line x1="10" x2="21" y1="14" y2="3"/>
      </svg>
    </a>
  </div>

  <div class="info-card credits">
    <p>
      Created by <a href="https://github.com/johnste" target="_blank" rel="noopener noreferrer">John Sterling</a>
    </p>
    <p>
      Icon designed by <a href="https://github.com/uetchy" target="_blank" rel="noopener noreferrer">@uetchy</a>
    </p>
  </div>
</div>

<style>
  .page-container {
    max-width: 700px;
    margin: 0 auto;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 24px;
    align-items: center;
  }

  .icon-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  .icon {
    height: 80px;
    opacity: 0.9;
  }

  .version {
    color: var(--text-secondary);
    font-size: 1.1em;
    margin: 0;
  }

  .info-card {
    width: 100%;
    padding: 20px;
    background: rgba(0, 0, 0, 0.1);
    border-radius: 8px;
    text-align: center;
  }

  .info-card.credits p {
    margin: 8px 0;
    color: var(--text-secondary);
    line-height: 1.6;
  }

  .external-link {
    color: var(--accent-color);
    text-decoration: none;
    font-size: 1.1em;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    transition: opacity 0.2s ease;
  }

  .external-link:hover {
    opacity: 0.8;
  }

  .external-link svg {
    opacity: 0.7;
  }

  a {
    color: var(--accent-color);
    text-decoration: none;
    transition: opacity 0.2s ease;
  }

  a:hover {
    opacity: 0.8;
  }
</style>
