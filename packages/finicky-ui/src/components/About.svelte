<script lang="ts">
  import DebugMessageToggle from "./DebugMessageToggle.svelte";

  let {
    toggleDevMode,
    version,
    isDevMode,
    addMessage,
  }: {
    toggleDevMode: () => void;
    version: string;
    isDevMode: boolean;
    addMessage: (msg: LogEntry) => void;
  } = $props();

  function handleIconClick() {
    toggleDevMode();
  }
</script>

<div class="about">
  <div class="main-section">
    <div class="icon-container">
      <button class="icon-button" ondblclick={handleIconClick}>
        <picture>
          <source
            media="(prefers-color-scheme: dark)"
            srcset="/finicky-logo.png"
          />
          <img
            alt="Finicky icon"
            src="/finicky-logo-light.png"
            class="icon {isDevMode ? 'dev-mode' : ''}"
          />
        </picture>
      </button>
    </div>
  </div>

  <div class="dev-mode-section {isDevMode ? 'dev-mode' : ''}">
    Developer mode
  </div>
  <p class="version">{version}</p>

  <div class="about-section">
    <p>
      <a href="https://github.com/johnste/finicky" target="_blank"
        >https://github.com/johnste/finicky</a
      >
    </p>
  </div>
  <div class="about-section">
    <p>
      Created by <a href="https://github.com/johnste" target="_blank"
        >John Sterling</a
      >
      <br />
      Icon designed by
      <a href="https://github.com/uetchy" target="_blank">@uetchy</a>
    </p>
  </div>
  {#if isDevMode}
    <DebugMessageToggle onAddMessage={addMessage} />
  {/if}
</div>

<style>
  .about {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px;
  }

  .main-section {
    display: flex;
    align-items: center;
  }

  .about-section {
    text-align: center;
  }

  .icon-button {
    /* Disable macOS Live Text feature on the icon */
    -webkit-user-select: none;
    user-select: none;
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
  }

  .icon {
    height: 128px;
    transition: all 1.3s cubic-bezier(0.23, 1.56, 0.64, 1);
  }

  .icon.dev-mode {
  }

  .dev-mode-section {
    transition: all 1.3s cubic-bezier(0.23, 1.56, 0.64, 1);
    background-image: linear-gradient(white, var(--accent-color));
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    font-size: 20px;
    opacity: 0;
    font-weight: bold;
    transform: perspective(1000px) rotateY(0deg);
    animation: float 3s ease-in-out infinite;
  }

  .dev-mode-section.dev-mode {
    opacity: 1;
  }

  @keyframes float {
    0% {
      transform: perspective(500px) rotateY(0deg) translateY(0px);
    }
    50% {
      transform: perspective(500px) rotateY(15deg) translateY(-5px);
    }
    100% {
      transform: perspective(500px) rotateY(0deg) translateY(0px);
    }
  }

  p {
    color: var(--text-secondary);
    line-height: 1.6;
    margin-bottom: 16px;
  }

  a {
    color: #b654ff;
    text-decoration: none;
    transition: opacity 0.2s ease;
  }

  a:hover {
    opacity: 0.8;
  }
</style>
