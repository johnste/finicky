<script lang="ts">
  import PageContainer from "../components/PageContainer.svelte";
  import LinkIcon from "../components/icons/Link.svelte";
  import InfoIcon from "../components/icons/Info.svelte";
  import SpinnerIcon from "../components/icons/Spinner.svelte";
  import { testUrlResult } from "../lib/testUrlStore";

  let testUrl = "";
  let loading = false;

  // Debounce timer
  let debounceTimer: ReturnType<typeof setTimeout>;
  let loadingTimer: ReturnType<typeof setTimeout>;

  function isValidUrl(url: string): boolean {
    if (!url.trim()) return false;
    try {
      // If no protocol, add https://
      const urlToTest = url.includes("://") ? url : `https://${url}`;
      new URL(urlToTest);
      return true;
    } catch {
      return false;
    }
  }

  function normalizeUrl(url: string): string {
    return url.includes("://") ? url : `https://${url}`;
  }


  const DEBOUNCE_DELAY = 300;
  const LOADING_DELAY = DEBOUNCE_DELAY + 100;

  function testUrlAutomatically() {
    clearTimeout(debounceTimer);
    clearTimeout(loadingTimer);
    loading = false;

    if (!isValidUrl(testUrl)) {
      testUrlResult.set(null);
      return;
    }

    // Only show loading indicator if it takes more than 100ms
    loadingTimer = setTimeout(() => {
      loading = true;
    }, LOADING_DELAY);

    debounceTimer = setTimeout(() => {
      const normalizedUrl = normalizeUrl(testUrl);

      // Send message to native app to test the URL
      window.finicky.sendMessage({
        type: "testUrl",
        url: normalizedUrl,
      });
    }, DEBOUNCE_DELAY);
  }

  // Subscribe to test results
  testUrlResult.subscribe(() => {
    clearTimeout(loadingTimer);
    loading = false;
  });

  // Reactive statement to test URL whenever it changes
  $: if (testUrl !== undefined) {
    testUrlAutomatically();
  }
</script>

<PageContainer
  title="Test"
  description="Test how Finicky will handle a URL based on your current configuration"
>
  <div class="test-card">
    <div class="input-section">
      <label for="url-input" class="input-label">
        Enter URL
        {#if loading}
          <div class="loading-spinner">
            <SpinnerIcon />
          </div>
        {/if}
      </label>
      <input
        id="url-input"
        type="text"
        class="url-input"
        placeholder="https://example.com"
        autocapitalize="off"
        autocorrect="off"
        required
        bind:value={testUrl}
      />
    </div>

    {#if $testUrlResult}
      <div class="result-section">
        <div class="result-header">
          <h3>Result</h3>
        </div>
        <div class="result-grid">
          <div class="result-item">
            <span class="result-label">Browser</span>
            <span class="result-value browser">{$testUrlResult.browser}</span>
          </div>
          <div class="result-item">
            <span class="result-label">Profile</span>
            <span class="result-value">{$testUrlResult.profile || "N/A"}</span>
          </div>
          {#if typeof $testUrlResult.openInBackground === "boolean"}
          <div class="result-item">
            <span class="result-label">Open in background</span>
            <span class="result-value"
              >{$testUrlResult.openInBackground ? "Yes" : "No"}</span
            >
          </div>
          {/if}
          <div class="result-item full-width">
            <span class="result-label">Final URL</span>
            <span class="result-value url">{$testUrlResult.url}</span>
          </div>
        </div>
      </div>
    {:else if testUrl.trim() && !isValidUrl(testUrl)}
      <div class="hint-message">
        <InfoIcon />
        Enter a valid URL to see how Finicky will handle it
      </div>
    {:else if !testUrl.trim()}
      <div class="empty-state">
        <LinkIcon />
        <p>Enter a URL above to test your configuration</p>
      </div>
    {/if}
  </div>
</PageContainer>

<style>
  .test-card {
    background: rgba(0, 0, 0, 0.1);
    border-radius: 12px;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .input-section {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .input-label {
    color: var(--text-primary);
    font-size: 0.9em;
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .loading-spinner {
    width: 16px;
    height: 16px;
    color: var(--accent-color);
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  .url-input {
    padding: 14px 16px;
    font-size: 1em;
    background: rgba(0, 0, 0, 0.2);
    border: 2px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    transition: all 0.2s ease;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
      monospace;
  }

  .url-input:focus {
    outline: none;
    border-color: var(--accent-color);
    background: rgba(0, 0, 0, 0.3);
  }

  .url-input::placeholder {
    color: var(--text-secondary);
    opacity: 0.4;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 48px 24px;
    gap: 16px;
    color: var(--text-secondary);
    opacity: 0.6;
  }

  .empty-state p {
    margin: 0;
    font-size: 0.95em;
  }

  .hint-message {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 16px;
    background: rgba(180, 84, 255, 0.08);
    border: 1px solid rgba(180, 84, 255, 0.2);
    border-radius: 8px;
    color: var(--text-secondary);
    font-size: 0.9em;
  }

  .result-section {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .result-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--border-color);
  }

  h3 {
    margin: 0;
    color: var(--text-primary);
    font-size: 1.1em;
  }

  .result-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
  }

  .result-item {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 12px;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 8px;
  }

  .result-item.full-width {
    grid-column: 1 / -1;
  }

  .result-label {
    color: var(--text-secondary);
    font-size: 0.8em;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 500;
  }

  .result-value {
    color: var(--text-primary);
    font-size: 1em;
    word-break: break-word;
    overflow-wrap: break-word;
    min-width: 0;
  }

  .result-value.browser {
    font-weight: 600;
    color: var(--accent-color);
  }

  .result-value.url {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
      monospace;
    font-size: 0.9em;
    word-break: break-all;
  }
</style>
