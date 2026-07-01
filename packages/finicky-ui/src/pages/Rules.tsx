import { useState, useEffect, useRef, useSyncExternalStore } from "react";
import clsx from "clsx";
import { PageContainer } from "../components/PageContainer";
import { BrowserProfileSelector } from "../components/BrowserProfileSelector";
import { Tooltip } from "../components/Tooltip";
import { WarningIcon } from "../components/icons/Warning";
import { XIcon } from "../components/icons/X";
import { appStore } from "../lib/appStore";
import { api } from "../lib/api";
import { useRulesSave } from "../lib/useRulesSave";
import type { Rule, BrowserProfile, BrowserOptions, BrowserProfileCustom } from "../types";
import styles from "./Rules.module.css";

const SAVE_DEBOUNCE = 3000;

type RowState = BrowserProfileCustom;

function patternNeedsWildcard(pattern: string): boolean {
  return !/\*/.test(pattern.trim()) && pattern.trim().length > 0;
}

function normalizeRules(rules: Rule[]): Rule[] {
  return rules.map((r) => ({
    ...r,
    match: Array.isArray(r.match) ? r.match : r.match ? [r.match as unknown as string] : [""],
  }));
}

function profileIsCustom(rule: Rule, profiles: Record<string, string[]>): boolean {
  const profile = rule.profile ?? "";
  if (!profile) return false;
  const options = profiles[rule.browser];
  return options !== undefined && !options.includes(profile);
}

function computeRowStates(rules: Rule[], browsers: BrowserOptions): RowState[] {
  return rules.map((r) => ({
    browser: r.browser !== "" && !browsers.installed.includes(r.browser),
    profile: profileIsCustom(r, browsers.profiles),
  }));
}

async function fetchMissingProfiles(rules: Rule[], browsers: BrowserOptions) {
  const missing = [...new Set(rules.map((r) => r.browser).filter((b) => b && browsers.profiles[b] === undefined))];
  await Promise.all(
    missing.map(async (b) => {
      try {
        appStore.addBrowserProfiles(b, await api.getBrowserProfiles(b));
      } catch {}
    })
  );
}

function useDragSort(onReorder: (from: number, to: number) => void, onDone: () => void) {
  const [dragIndex, setDragIndex] = useState<number | null>(null);
  return {
    dragIndex,
    onDragStart: (i: number) => setDragIndex(i),
    onDragOver: (e: React.DragEvent, i: number) => {
      e.preventDefault();
      if (dragIndex === null || dragIndex === i) return;
      onReorder(dragIndex, i);
      setDragIndex(i);
    },
    onDragEnd: () => { setDragIndex(null); onDone(); },
  };
}

function usePendingFocus() {
  const ref = useRef<{ rule: number; pattern: number } | null>(null);
  return {
    set: (rule: number, pattern: number) => { ref.current = { rule, pattern }; },
    inputRef: (el: HTMLInputElement | null, ruleIdx: number, patternIdx: number) => {
      if (el && ref.current?.rule === ruleIdx && ref.current?.pattern === patternIdx) {
        el.focus();
        ref.current = null;
      }
    },
  };
}

interface RuleRowCallbacks {
  onDragStart: () => void;
  onDragOver: (e: React.DragEvent) => void;
  onDragEnd: () => void;
  onChange: (value: BrowserProfile, custom: BrowserProfileCustom, committed: boolean) => void;
  onMatchInput: (j: number, e: React.ChangeEvent<HTMLInputElement>) => void;
  onBlurSave: () => void;
  onAddPattern: () => void;
  onRemovePattern: (j: number) => void;
  onRemoveRule: () => void;
  onRequestProfiles: (b: string) => void;
  inputRef: (el: HTMLInputElement | null, j: number) => void;
}

interface RuleRowProps {
  rule: Rule;
  isDragging: boolean;
  custom: BrowserProfileCustom;
  browsers: BrowserOptions;
  callbacks: RuleRowCallbacks;
}

function RuleRow({ rule, isDragging, custom, browsers, callbacks: cb }: RuleRowProps) {
  return (
    <div
      className={clsx(
        styles.ruleRow,
        isDragging && styles.dragging,
        !rule.browser && !custom.browser && styles.noBrowser
      )}
      draggable
      onDragStart={cb.onDragStart}
      onDragOver={cb.onDragOver}
      onDragEnd={cb.onDragEnd}
      role="listitem"
    >
      <div className={styles.ruleTop}>
        <div className={styles.dragHandle} aria-hidden="true">⠿</div>
        <BrowserProfileSelector
          value={{ browser: rule.browser, profile: rule.profile ?? "" }}
          custom={custom}
          browsers={browsers}
          required
          onChange={cb.onChange}
          onRequestProfiles={cb.onRequestProfiles}
        />
        {!rule.browser && !custom.browser && (
          <span className={styles.browserRequiredHint}>
            <WarningIcon />
            Browser required
          </span>
        )}
        <button className={styles.deleteBtn} onClick={cb.onRemoveRule} aria-label="Delete rule">
          <XIcon />
        </button>
      </div>

      <div className={styles.ruleBottom}>
        <div className={styles.patterns}>
          {rule.match.map((pattern, j) => (
            <div key={j} className={styles.patternRow}>
              <div className={clsx(styles.patternInputWrapper, patternNeedsWildcard(pattern) && styles.hasWarning)}>
                <input
                  ref={(el) => cb.inputRef(el, j)}
                  className={clsx(styles.textInput, styles.patternInput)}
                  type="text"
                  placeholder="*.example.com/*"
                  value={pattern}
                  onChange={(e) => cb.onMatchInput(j, e)}
                  onBlur={cb.onBlurSave}
                />
                {patternNeedsWildcard(pattern) && (
                  <Tooltip text="Exact URLs rarely match, maybe you want to use a wildcard?">
                    <span className={styles.patternWarningIcon} aria-label="Exact URLs rarely match">
                      <WarningIcon />
                    </span>
                  </Tooltip>
                )}
              </div>
              {rule.match.length > 1 && (
                <button className={styles.removePatternBtn} onClick={() => cb.onRemovePattern(j)} aria-label="Remove pattern">
                  <XIcon />
                </button>
              )}
            </div>
          ))}
          <button className={styles.addPatternBtn} onClick={cb.onAddPattern}>
            + URL
          </button>
        </div>
      </div>
    </div>
  );
}

export function Rules() {
  const { rulesFile, installedBrowsers, profilesByBrowser, config } = useSyncExternalStore(
    appStore.subscribe,
    appStore.getSnapshot
  );
  const hasJsConfig = config.hasJsConfig ?? false;
  const browsers = { installed: installedBrowsers, profiles: profilesByBrowser };

  const [rules, setRules] = useState<Rule[]>(() => normalizeRules(rulesFile.rules));
  const [rowStates, setRowStates] = useState<RowState[]>([]);

  // save()/scheduleSave() can flush immediately (synchronously, in the same
  // tick as the setRules() call below it), before React has re-rendered and
  // produced a fresh `rules` value. rulesRef is updated in lockstep with
  // every setRules() call so useRulesSave always reads the latest rules
  // regardless of render timing, instead of a stale pre-edit closure.
  const rulesRef = useRef(rules);
  function updateRules(next: Rule[] | ((prev: Rule[]) => Rule[])) {
    const resolved = typeof next === "function" ? (next as (prev: Rule[]) => Rule[])(rulesRef.current) : next;
    rulesRef.current = resolved;
    setRules(resolved);
  }

  const { save, scheduleSave, isPending } = useRulesSave(
    () => ({ ...rulesFile, rules: rulesRef.current }),
    SAVE_DEBOUNCE
  );
  const drag = useDragSort(
    (from, to) => {
      const move = <T,>(arr: T[]) => { const a = [...arr]; a.splice(to, 0, ...a.splice(from, 1)); return a; };
      updateRules(move);
      setRowStates(move);
    },
    scheduleSave
  );
  const focus = usePendingFocus();

  // The installed-browser list only otherwise comes from the one-time
  // /api/initial-data fetch at app launch, and the WebView page is never
  // reloaded for the life of the process, so without this a browser
  // installed while Finicky is running would never show up here.
  useEffect(() => {
    api.getBrowsers().then((installed) => appStore.update({ installedBrowsers: installed })).catch(() => {});
  }, []);

  useEffect(() => {
    if (isPending.current) return;
    const newRules = normalizeRules(rulesFile.rules);
    updateRules(newRules);
    setRowStates(computeRowStates(newRules, browsers));
    fetchMissingProfiles(newRules, browsers);
  }, [rulesFile.rules, installedBrowsers]); // profilesByBrowser intentionally omitted

  useEffect(() => {
    setRowStates((prev) =>
      prev.map((s, i) => ({ ...s, profile: profileIsCustom(rules[i], profilesByBrowser) }))
    );
  }, [profilesByBrowser]);

  function onRowMatchInput(i: number, j: number, e: React.ChangeEvent<HTMLInputElement>) {
    const newMatch = [...rules[i].match];
    newMatch[j] = e.target.value;
    updateRules(rules.map((r, idx) => (idx === i ? { ...r, match: newMatch } : r)));
    scheduleSave();
  }

  function addPattern(i: number) {
    focus.set(i, rules[i].match.length);
    updateRules(rules.map((r, idx) => (idx === i ? { ...r, match: [...r.match, ""] } : r)));
  }

  function removePattern(i: number, j: number) {
    const newMatch = rules[i].match.filter((_, idx) => idx !== j);
    updateRules(rules.map((r, idx) => (idx === i ? { ...r, match: newMatch.length > 0 ? newMatch : [""] } : r)));
    save();
  }

  function addRule() {
    focus.set(rules.length, 0);
    updateRules([...rules, { match: [""], browser: "", profile: "" }]);
    setRowStates((prev) => [...prev, { browser: false, profile: false }]);
  }

  function removeRule(i: number) {
    updateRules(rules.filter((_, idx) => idx !== i));
    setRowStates((prev) => prev.filter((_, idx) => idx !== i));
    save();
  }

  function handleChange(i: number, { browser, profile }: BrowserProfile, custom: BrowserProfileCustom) {
    updateRules(rules.map((r, idx) => (idx === i ? { ...r, browser, profile } : r)));
    setRowStates((prev) => prev.map((s, idx) => (idx === i ? custom : s)));
  }

  const description = hasJsConfig ? (
    <>
      The first matching rule wins.{" "}
      <span style={{ color: "var(--accent-color)" }}>JavaScript configuration file loaded</span> — its handlers run
      first and take priority over these rules.
    </>
  ) : (
    <>
      The first matching rule wins. Use <code>*</code> as a wildcard, e.g.{" "}
      <code>*example.com/*</code>.
    </>
  );

  return (
    <PageContainer title="Rules" description={description}>
      {rules.length === 0 ? (
        <div className={styles.emptyRules}>No rules yet. Add one below.</div>
      ) : (
        <div className={styles.rulesList}>
          {rules.map((rule, i) => (
            <RuleRow
              key={i}
              rule={rule}
              isDragging={drag.dragIndex === i}
              custom={rowStates[i] ?? { browser: false, profile: false }}
              browsers={browsers}
              callbacks={{
                onDragStart: () => drag.onDragStart(i),
                onDragOver: (e) => drag.onDragOver(e, i),
                onDragEnd: drag.onDragEnd,
                onChange: (bp, custom, committed) => {
                  handleChange(i, bp, custom);
                  if (committed) save(); else scheduleSave();
                },
                onMatchInput: (j, e) => onRowMatchInput(i, j, e),
                onBlurSave: save,
                onAddPattern: () => addPattern(i),
                onRemovePattern: (j) => removePattern(i, j),
                onRemoveRule: () => removeRule(i),
                onRequestProfiles: async (b) => { try { appStore.addBrowserProfiles(b, await api.getBrowserProfiles(b)); } catch {} },
                inputRef: (el, j) => focus.inputRef(el, i, j),
              }}
            />
          ))}
        </div>
      )}
      <button className={styles.addRuleBtn} onClick={addRule}>
        + Add rule
      </button>
    </PageContainer>
  );
}
