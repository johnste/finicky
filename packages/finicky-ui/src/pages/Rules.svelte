<script lang="ts">
  import PageContainer from "../components/PageContainer.svelte";
  import type { Rule, RulesFile } from "../types";

  let {
    rulesFile = { defaultBrowser: "", rules: [] },
    installedBrowsers = [],
    profilesByBrowser = {},
  }: {
    rulesFile: RulesFile;
    installedBrowsers: string[];
    profilesByBrowser: Record<string, string[]>;
  } = $props();

  const CUSTOM = "__custom__";
  const SAVE_DEBOUNCE = 500;

  // Local editable copies
  let defaultBrowser = $state(rulesFile.defaultBrowser ?? "");
  let defaultProfile = $state(rulesFile.defaultProfile ?? "");
  let rules = $state<Rule[]>(rulesFile.rules.map((r: Rule) => ({ ...r })));

  // "Custom..." state for default browser
  let defaultBrowserIsCustom = $state(false);
  let defaultProfileIsCustom = $state(false);

  // Per-row "Custom..." state
  let rowIsCustom = $state<boolean[]>([]);
  let rowProfileIsCustom = $state<boolean[]>([]);

  function profileOptions(browser: string): string[] {
    return profilesByBrowser[browser] ?? [];
  }

  let saveTimer: ReturnType<typeof setTimeout>;

  function scheduleSave() {
    clearTimeout(saveTimer);
    saveTimer = setTimeout(() => {
      const payload: RulesFile = {
        defaultBrowser,
        defaultProfile,
        rules,
      };
      window.finicky.sendMessage({ type: "saveRules", payload });
    }, SAVE_DEBOUNCE);
  }

  function onDefaultBrowserSelect(e: Event) {
    const val = (e.target as HTMLSelectElement).value;
    if (val === CUSTOM) {
      defaultBrowserIsCustom = true;
      defaultBrowser = "";
      defaultProfile = "";
      defaultProfileIsCustom = false;
    } else {
      defaultBrowserIsCustom = false;
      defaultBrowser = val;
      defaultProfile = "";
      defaultProfileIsCustom = false;
      if (val && profilesByBrowser[val] === undefined) {
        window.finicky.sendMessage({ type: "getBrowserProfiles", browser: val });
      }
      scheduleSave();
    }
  }

  function onDefaultBrowserCustomInput(e: Event) {
    defaultBrowser = (e.target as HTMLInputElement).value;
    scheduleSave();
  }

  function onDefaultProfileSelect(e: Event) {
    const val = (e.target as HTMLSelectElement).value;
    if (val === CUSTOM) {
      defaultProfileIsCustom = true;
      defaultProfile = "";
    } else {
      defaultProfileIsCustom = false;
      defaultProfile = val;
      scheduleSave();
    }
  }

  function onDefaultProfileCustomInput(e: Event) {
    defaultProfile = (e.target as HTMLInputElement).value;
    scheduleSave();
  }

  function onRowBrowserSelect(i: number, e: Event) {
    const val = (e.target as HTMLSelectElement).value;
    if (val === CUSTOM) {
      rowIsCustom[i] = true;
      rules[i] = { ...rules[i], browser: "", profile: "" };
      rowProfileIsCustom[i] = false;
    } else {
      rowIsCustom[i] = false;
      rules[i] = { ...rules[i], browser: val, profile: "" };
      rowProfileIsCustom[i] = false;
      if (val && profilesByBrowser[val] === undefined) {
        window.finicky.sendMessage({ type: "getBrowserProfiles", browser: val });
      }
      scheduleSave();
    }
  }

  function onRowProfileSelect(i: number, e: Event) {
    const val = (e.target as HTMLSelectElement).value;
    if (val === CUSTOM) {
      rowProfileIsCustom[i] = true;
      rules[i] = { ...rules[i], profile: "" };
    } else {
      rowProfileIsCustom[i] = false;
      rules[i] = { ...rules[i], profile: val };
      scheduleSave();
    }
  }

  function onRowProfileCustomInput(i: number, e: Event) {
    rules[i] = { ...rules[i], profile: (e.target as HTMLInputElement).value };
    scheduleSave();
  }

  function onRowBrowserCustomInput(i: number, e: Event) {
    rules[i] = { ...rules[i], browser: (e.target as HTMLInputElement).value };
    scheduleSave();
  }

  function onRowMatchInput(i: number, e: Event) {
    rules[i] = { ...rules[i], match: (e.target as HTMLInputElement).value };
    scheduleSave();
  }

  function addRule() {
    rules = [...rules, { match: "", browser: "", profile: "" }];
    rowIsCustom = [...rowIsCustom, false];
    rowProfileIsCustom = [...rowProfileIsCustom, false];
    scheduleSave();
  }

  function removeRule(i: number) {
    rules = rules.filter((_, idx) => idx !== i);
    rowIsCustom = rowIsCustom.filter((_, idx) => idx !== i);
    rowProfileIsCustom = rowProfileIsCustom.filter((_, idx) => idx !== i);
    scheduleSave();
  }

  // Drag-to-reorder
  let dragIndex = $state<number | null>(null);

  function onDragStart(i: number) {
    dragIndex = i;
  }

  function onDragOver(e: DragEvent, i: number) {
    e.preventDefault();
    if (dragIndex === null || dragIndex === i) return;
    const newRules = [...rules];
    const newCustom = [...rowIsCustom];
    const newProfileCustom = [...rowProfileIsCustom];
    const [movedRule] = newRules.splice(dragIndex, 1);
    const [movedCustom] = newCustom.splice(dragIndex, 1);
    const [movedProfileCustom] = newProfileCustom.splice(dragIndex, 1);
    newRules.splice(i, 0, movedRule);
    newCustom.splice(i, 0, movedCustom);
    newProfileCustom.splice(i, 0, movedProfileCustom);
    rules = newRules;
    rowIsCustom = newCustom;
    rowProfileIsCustom = newProfileCustom;
    dragIndex = i;
  }

  function onDragEnd() {
    dragIndex = null;
    scheduleSave();
  }

  // Request data from native on mount
  $effect(() => {
    window.finicky.sendMessage({ type: "getRules" });
    window.finicky.sendMessage({ type: "getInstalledBrowsers" });
  });

  // Sync incoming props into local state when they change; also fetch profiles for known browsers
  $effect(() => {
    const newDefaultBrowser = rulesFile.defaultBrowser ?? "";
    const newDefaultProfile = rulesFile.defaultProfile ?? "";
    const newRules = rulesFile.rules.map((r: Rule) => ({ ...r }));
    defaultBrowser = newDefaultBrowser;
    defaultProfile = newDefaultProfile;
    rules = newRules;
    defaultBrowserIsCustom =
      newDefaultBrowser !== "" && !installedBrowsers.includes(newDefaultBrowser);
    defaultProfileIsCustom =
      newDefaultProfile !== "" && !profileOptions(newDefaultBrowser).includes(newDefaultProfile);
    rowIsCustom = newRules.map(
      (r) => r.browser !== "" && !installedBrowsers.includes(r.browser)
    );
    rowProfileIsCustom = newRules.map(
      (r) => (r.profile ?? "") !== "" && !profileOptions(r.browser).includes(r.profile ?? "")
    );
    // Fetch profiles for all browsers present in the rules file
    const browsersToFetch = new Set<string>();
    if (newDefaultBrowser) browsersToFetch.add(newDefaultBrowser);
    for (const r of newRules) {
      if (r.browser) browsersToFetch.add(r.browser);
    }
    for (const b of browsersToFetch) {
      window.finicky.sendMessage({ type: "getBrowserProfiles", browser: b });
    }
  });
</script>

<PageContainer
  title="Rules"
  description="Rules are checked before your JS config. The first matching rule wins."
>
  <!-- Default browser -->
  <div class="section">
    <div class="section-header">
      <span class="section-label">Default browser</span>
      <span class="section-hint">Used when no rule matches</span>
    </div>
    <div class="browser-select-row">
      {#if defaultBrowserIsCustom}
        <input
          class="text-input"
          type="text"
          placeholder="e.g. Firefox"
          value={defaultBrowser}
          oninput={onDefaultBrowserCustomInput}
        />
        <button
          class="clear-custom-btn"
          onclick={() => {
            defaultBrowserIsCustom = false;
            defaultBrowser = "";
            scheduleSave();
          }}
        >
          ✕
        </button>
      {:else}
        <select
          class="browser-dropdown"
          value={defaultBrowser}
          onchange={onDefaultBrowserSelect}
        >
          <option value="">System default</option>
          {#each installedBrowsers as b}
            <option value={b}>{b}</option>
          {/each}
          <option value={CUSTOM}>Custom...</option>
        </select>
      {/if}
      {#if !defaultBrowserIsCustom && defaultBrowser && profileOptions(defaultBrowser).length > 0}
        {#if defaultProfileIsCustom}
          <input
            class="text-input"
            type="text"
            placeholder="Profile name"
            value={defaultProfile}
            oninput={onDefaultProfileCustomInput}
          />
          <button
            class="clear-custom-btn"
            onclick={() => {
              defaultProfileIsCustom = false;
              defaultProfile = "";
              scheduleSave();
            }}
          >
            ✕
          </button>
        {:else}
          <select
            class="browser-dropdown"
            value={defaultProfile}
            onchange={onDefaultProfileSelect}
          >
            <option value="">No profile</option>
            {#each profileOptions(defaultBrowser) as p}
              <option value={p}>{p}</option>
            {/each}
            <option value={CUSTOM}>Custom...</option>
          </select>
        {/if}
      {/if}
    </div>
  </div>

  <!-- Rules list -->
  <div class="section">
    <div class="section-header">
      <span class="section-label">Rules</span>
    </div>

    {#if rules.length === 0}
      <div class="empty-rules">
        No rules yet. Add one below.
      </div>
    {:else}
      <div class="rules-list">
        {#each rules as rule, i}
          <div
            class="rule-row"
            class:dragging={dragIndex === i}
            draggable="true"
            ondragstart={() => onDragStart(i)}
            ondragover={(e) => onDragOver(e, i)}
            ondragend={onDragEnd}
            role="listitem"
          >
            <div class="drag-handle" aria-hidden="true">⠿</div>

            <input
              class="text-input pattern-input"
              type="text"
              placeholder="*.example.com/*"
              value={rule.match}
              oninput={(e) => onRowMatchInput(i, e)}
            />

            <div class="arrow">→</div>

            {#if rowIsCustom[i]}
              <input
                class="text-input browser-input"
                type="text"
                placeholder="e.g. Firefox"
                value={rule.browser}
                oninput={(e) => onRowBrowserCustomInput(i, e)}
              />
              <button
                class="clear-custom-btn"
                onclick={() => {
                  rowIsCustom[i] = false;
                  rules[i] = { ...rules[i], browser: "" };
                  scheduleSave();
                }}
              >
                ✕
              </button>
            {:else}
              <select
                class="browser-dropdown"
                value={rule.browser}
                onchange={(e) => onRowBrowserSelect(i, e)}
              >
                <option value="">Select browser</option>
                {#each installedBrowsers as b}
                  <option value={b}>{b}</option>
                {/each}
                <option value={CUSTOM}>Custom...</option>
              </select>
            {/if}

            {#if !rowIsCustom[i] && rule.browser && profileOptions(rule.browser).length > 0}
              {#if rowProfileIsCustom[i]}
                <input
                  class="text-input browser-input"
                  type="text"
                  placeholder="Profile name"
                  value={rule.profile ?? ""}
                  oninput={(e) => onRowProfileCustomInput(i, e)}
                />
                <button
                  class="clear-custom-btn"
                  onclick={() => {
                    rowProfileIsCustom[i] = false;
                    rules[i] = { ...rules[i], profile: "" };
                    scheduleSave();
                  }}
                >
                  ✕
                </button>
              {:else}
                <select
                  class="browser-dropdown"
                  value={rule.profile ?? ""}
                  onchange={(e) => onRowProfileSelect(i, e)}
                >
                  <option value="">No profile</option>
                  {#each profileOptions(rule.browser) as p}
                    <option value={p}>{p}</option>
                  {/each}
                  <option value={CUSTOM}>Custom...</option>
                </select>
              {/if}
            {/if}

            <button
              class="delete-btn"
              onclick={() => removeRule(i)}
              aria-label="Delete rule"
            >
              ✕
            </button>
          </div>
        {/each}
      </div>
    {/if}

    <button class="add-rule-btn" onclick={addRule}>+ Add rule</button>
  </div>
</PageContainer>

<style>
  .section {
    background: rgba(0, 0, 0, 0.1);
    border-radius: 12px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .section-header {
    display: flex;
    align-items: baseline;
    gap: 10px;
  }

  .section-label {
    color: var(--text-primary);
    font-size: 0.9em;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .section-hint {
    color: var(--text-secondary);
    font-size: 0.82em;
    opacity: 0.7;
  }

  .browser-select-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .browser-dropdown {
    flex: 1;
    padding: 8px 12px;
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    font-size: 0.9em;
    cursor: pointer;
  }

  .browser-dropdown:focus {
    outline: none;
    border-color: var(--accent-color);
  }

  .text-input {
    flex: 1;
    padding: 8px 12px;
    background: rgba(0, 0, 0, 0.2);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    color: var(--text-primary);
    font-size: 0.9em;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }

  .text-input:focus {
    outline: none;
    border-color: var(--accent-color);
  }

  .text-input::placeholder {
    color: var(--text-secondary);
    opacity: 0.4;
  }

  .browser-input {
    font-family: inherit;
  }

  .empty-rules {
    color: var(--text-secondary);
    font-size: 0.9em;
    opacity: 0.6;
    padding: 16px 0 4px;
    text-align: center;
  }

  .rules-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .rule-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    background: rgba(0, 0, 0, 0.15);
    border-radius: 8px;
    border: 1px solid transparent;
    transition: border-color 0.15s;
  }

  .rule-row.dragging {
    border-color: var(--accent-color);
    opacity: 0.7;
  }

  .drag-handle {
    color: var(--text-secondary);
    opacity: 0.4;
    cursor: grab;
    font-size: 1.1em;
    user-select: none;
    flex-shrink: 0;
  }

  .pattern-input {
    flex: 2;
  }

  .arrow {
    color: var(--text-secondary);
    opacity: 0.5;
    flex-shrink: 0;
    font-size: 0.9em;
  }

  .delete-btn,
  .clear-custom-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    opacity: 0.4;
    cursor: pointer;
    font-size: 0.85em;
    padding: 4px 6px;
    border-radius: 4px;
    flex-shrink: 0;
    transition: opacity 0.15s;
  }

  .delete-btn:hover,
  .clear-custom-btn:hover {
    opacity: 1;
    color: #f44336;
  }

  .add-rule-btn {
    align-self: flex-start;
    background: rgba(180, 84, 255, 0.12);
    border: 1px solid rgba(180, 84, 255, 0.3);
    border-radius: 8px;
    color: var(--accent-color);
    font-size: 0.88em;
    padding: 8px 16px;
    cursor: pointer;
    transition: background 0.15s;
  }

  .add-rule-btn:hover {
    background: rgba(180, 84, 255, 0.22);
  }
</style>
