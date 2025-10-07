<script lang="ts">
  let testUrl = "";
  let result: any = null;
  let loading = false;

  // Debounce timer
  let debounceTimer: ReturnType<typeof setTimeout>;
  let loadingTimer: ReturnType<typeof setTimeout>;

  function isValidUrl(url: string): boolean {
    if (!url.trim()) return false;
    try {
      // If no protocol, add https://
      const urlToTest = url.includes('://') ? url : `https://${url}`;
      new URL(urlToTest);
      return true;
    } catch {
      return false;
    }
  }

  function normalizeUrl(url: string): string {
    return url.includes('://') ? url : `https://${url}`;
  }

  function testUrlAutomatically() {
    clearTimeout(debounceTimer);
    clearTimeout(loadingTimer);
    loading = false;
    
    if (!isValidUrl(testUrl)) {
      result = null;
      return;
    }

    // Only show loading indicator if it takes more than 100ms
    loadingTimer = setTimeout(() => {
      loading = true;
    }, 100);
    
    debounceTimer = setTimeout(() => {
      // TODO: Send message to native app to test the URL
      // For now, just simulate a result
      const normalizedUrl = normalizeUrl(testUrl);
      result = {
        url: normalizedUrl,
        browser: "Safari",
        openInBackground: false,
        profile: "",
      };
      clearTimeout(loadingTimer);
      loading = false;
    }, 0);
  }

  // Reactive statement to test URL whenever it changes
  $: if (testUrl !== undefined) {
    testUrlAutomatically();
  }
</script>

<div class="test-container">
  <div class="test-section">
    <h2>Test URL Handler</h2>
    <p class="description">
      Test how Finicky will handle a URL based on your current configuration
    </p>

    <div class="input-section">
      <input
        type="text"
        class="url-input"
        placeholder="Enter a URL to test (e.g., https://example.com)"
        bind:value={testUrl}
      />
      {#if loading}
        <div class="loading-indicator">Testing...</div>
      {/if}
    </div>

    {#if result}
      <div class="result-section">
        <h3>Result</h3>
        <div class="result-card">
          <div class="result-row">
            <span class="result-label">URL:</span>
            <span class="result-value">{result.url}</span>
          </div>
          <div class="result-row">
            <span class="result-label">Browser:</span>
            <span class="result-value">{result.browser}</span>
          </div>
          <div class="result-row">
            <span class="result-label">Open in Background:</span>
            <span class="result-value">{result.openInBackground ? "Yes" : "No"}</span>
          </div>
          {#if result.profile}
            <div class="result-row">
              <span class="result-label">Profile:</span>
              <span class="result-value">{result.profile}</span>
            </div>
          {/if}
        </div>
      </div>
    {:else if testUrl.trim() && !isValidUrl(testUrl)}
      <div class="hint-message">
        Enter a valid URL to see how Finicky will handle it
      </div>
    {/if}
  </div>
</div>

<style>
  .test-container {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
  }

  .test-section {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  h2 {
    margin: 0;
    color: var(--text-primary);
    font-size: 1.5em;
  }

  .description {
    color: var(--text-secondary);
    margin: 0;
    font-size: 0.95em;
  }

  .input-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .url-input {
    padding: 14px 16px;
    font-size: 1em;
    background: var(--log-bg);
    border: 2px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    transition: border-color 0.2s ease;
  }

  .url-input:focus {
    outline: none;
    border-color: #b654ff;
  }

  .url-input::placeholder {
    color: var(--text-secondary);
    opacity: 0.5;
  }

  .loading-indicator {
    color: var(--text-secondary);
    font-size: 0.9em;
    opacity: 0.8;
    padding: 0 4px;
  }

  .hint-message {
    padding: 12px 16px;
    background: rgba(180, 84, 255, 0.05);
    border: 1px solid rgba(180, 84, 255, 0.2);
    border-radius: 8px;
    color: var(--text-secondary);
    font-size: 0.9em;
    opacity: 0.8;
  }

  .result-section {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  h3 {
    margin: 0;
    color: var(--text-primary);
    font-size: 1.2em;
  }

  .result-card {
    background: var(--log-bg);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .result-row {
    display: flex;
    gap: 12px;
    align-items: baseline;
  }

  .result-label {
    color: var(--text-secondary);
    font-weight: 500;
    min-width: 160px;
  }

  .result-value {
    color: var(--text-primary);
    word-break: break-all;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }
</style>
