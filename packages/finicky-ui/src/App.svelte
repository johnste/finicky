<script lang="ts">
  import { onMount } from 'svelte';
  import LogViewer from './components/LogViewer.svelte';
  import DebugMessageToggle from './components/DebugMessageToggle.svelte';
  import type { LogEntry } from './types';

  let version = 'v0.0.0';
  let buildInfo = 'dev';
  const isDevMode = import.meta.env.DEV;

  // Initialize message buffer
  let messageBuffer: LogEntry[] = [];

  // Function to handle messages from the native app
  function handleMessage(msg: any) {
    const parsedMsg = typeof msg === 'string' ? JSON.parse(msg) : msg;

    switch (parsedMsg.type) {
      case 'version':
        version = parsedMsg.message;
        break;
      case 'buildInfo':
        buildInfo = parsedMsg.message;
        break;
      default:
        if (parsedMsg.message) {
          messageBuffer = [...messageBuffer, JSON.parse(parsedMsg.message)];
        } else {
          messageBuffer = [...messageBuffer, parsedMsg];
        }
    }
  }

  // Handle debug messages from the DebugMessageToggle
  function handleDebugMessage(message: LogEntry) {
    messageBuffer = [...messageBuffer, message];
  }

  // Set up the bridge between native app and web UI
  onMount(() => {
    window.finicky = {
      sendMessage: (msg: any) => {
        window.webkit?.messageHandlers?.finicky?.postMessage(JSON.stringify(msg));
      },
      receiveMessage: handleMessage
    };
  });
</script>

<main>
  <div class="container">
    <h1>
      <img src="/finicky-icon.png" alt="Finicky" />
      <span>Finicky</span>
    </h1>

    <div class="header-meta">
      <a href="https://github.com/johnste/finicky" class="github-link" target="_blank">
        View on GitHub
      </a>
      <span class="version">{version}</span>
      <span class="build-info">{buildInfo}</span>
    </div>

    {#if isDevMode}
      <DebugMessageToggle onAddMessage={handleDebugMessage} />
    {/if}

    <LogViewer {messageBuffer} />
  </div>
</main>

<style>
  .container {
    padding: 16px;
    max-width: 100%;
    max-height: 100%;
    box-sizing: border-box;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
  }

  h1 {
    color: var(--text-primary);
    margin: 0 0 12px 0;
    font-size: 1.5em;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  h1 img {
    width: 32px;
    height: 32px;
  }

  h1 span {
    background: linear-gradient(135deg, #b654ff, #b654ff);
    background-clip: text;
    -webkit-background-clip: text;
    color: transparent;
    position: relative;
  }

  .header-meta {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
  }

  .github-link {
    color: var(--text-secondary);
    text-decoration: none;
    font-size: 0.9em;
    transition: color 0.2s ease;
  }

  .github-link:hover {
    color: var(--text-primary);
  }

  .version {
    color: var(--text-secondary);
    font-size: 0.9em;
    margin-left: 8px;
  }

  .build-info {
    color: var(--text-secondary);
    font-size: 0.8em;
    margin-left: 8px;
    opacity: 0.8;
  }
</style>
