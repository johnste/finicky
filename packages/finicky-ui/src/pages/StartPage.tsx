import { useState, useEffect, useRef, useSyncExternalStore } from "react";
import { useDebouncedCallback } from "../lib/useDebouncedCallback";
import { Link } from "react-router-dom";
import { PageContainer } from "../components/PageContainer";
import { BrowserProfileSelector } from "../components/BrowserProfileSelector";
import { OptionRow } from "../components/OptionRow";
import { Tooltip } from "../components/Tooltip";
import { ExternalIcon } from "../components/icons/External";
import { LockIcon } from "../components/icons/Lock";
import { toast } from "../lib/toast";
import { appStore } from "../lib/appStore";
import { api } from "../lib/api";
import type { RulesFile, ConfigInfo, UpdateInfo } from "../types";
import styles from "./StartPage.module.css";

const SAVE_DEBOUNCE = 500;
const SAFARI = "Safari";

interface Options {
  keepRunning: boolean;
  hideIcon: boolean;
  logRequests: boolean;
  checkForUpdates: boolean;
}

function resolveOptions(rulesFile: RulesFile, config: ConfigInfo): Options {
  return {
    keepRunning: rulesFile.options?.keepRunning ?? config.options?.keepRunning ?? true,
    hideIcon: rulesFile.options?.hideIcon ?? config.options?.hideIcon ?? false,
    logRequests: rulesFile.options?.logRequests ?? config.options?.logRequests ?? false,
    checkForUpdates: rulesFile.options?.checkForUpdates ?? config.options?.checkForUpdates ?? true,
  };
}

function isBrowserCustom(browser: string, installed: string[]): boolean {
  return browser !== "" && !installed.includes(browser);
}

function isProfileCustom(profile: string, browser: string, byBrowser: Record<string, string[]>): boolean {
  return profile !== "" && !(byBrowser[browser] ?? []).includes(profile);
}

function initialBp(rulesFile: RulesFile, config: ConfigInfo, hasJsConfig: boolean) {
  return {
    browser: hasJsConfig ? (config.defaultBrowser ?? "") : (rulesFile.defaultBrowser || SAFARI),
    profile: rulesFile.defaultProfile ?? "",
    browserCustom: false,
    profileCustom: false,
  };
}

function resolveBrowserIsCustom(customMode: boolean, browser: string, installed: string[]): boolean {
  return customMode || isBrowserCustom(browser, installed);
}

function resolveProfileIsCustom(customMode: boolean, profile: string, browser: string, byBrowser: Record<string, string[]>): boolean {
  return customMode || isProfileCustom(profile, browser, byBrowser);
}

function onLockedClick() {
  toast.show(
    "Configuration loaded from a JavaScript configuration file",
    "info",
    "These settings are managed by your config file and can't be changed here."
  );
}

function UpdateCard({ updateInfo }: { updateInfo: UpdateInfo }) {
  if (updateInfo.hasUpdate) {
    return (
      <div className={`${styles.statusCard} ${styles.info}`}>
        <div className={styles.updateHeader}>
          <h3>New Version Available</h3>
          <span className={styles.updateVersion}>{updateInfo.version}</span>
        </div>
        <div className={styles.updateActions}>
          <a href={updateInfo.downloadUrl} target="_blank" rel="noopener noreferrer" className={styles.downloadBtn}>
            Download {updateInfo.version}
          </a>
          <a href={updateInfo.releaseUrl} target="_blank" rel="noopener noreferrer" className={styles.releaseLink}>
            Release notes
          </a>
        </div>
      </div>
    );
  }
  if (!updateInfo.updateCheckEnabled) {
    return (
      <div className={`${styles.statusCard} ${styles.info}`}>
        <h3>Update check is disabled</h3>
        <a href="https://github.com/johnste/finicky/releases" target="_blank" rel="noopener noreferrer">
          Check releases
        </a>
      </div>
    );
  }
  return null;
}

export function StartPage() {
  const { hasConfig, config, updateInfo, rulesFile, installedBrowsers, profilesByBrowser, messageBuffer } =
    useSyncExternalStore(appStore.subscribe, appStore.getSnapshot);
  const hasJsConfig = config.hasJsConfig ?? false;
  const numErrors = messageBuffer.filter((m) => m.level.toLowerCase() === "error").length;

  const [options, setOptions] = useState<Options>(() => resolveOptions(rulesFile, config));
  const [bp, setBp] = useState(() => initialBp(rulesFile, config, hasJsConfig));

  const pendingRef = useRef({ options, bp, rulesFile });
  pendingRef.current = { options, bp, rulesFile };

  useEffect(() => {
    setOptions(resolveOptions(rulesFile, config));
    setBp(initialBp(rulesFile, config, hasJsConfig));
  }, [rulesFile, config, hasJsConfig]);

  const saveDebounced = useDebouncedCallback(async () => {
    const { options, bp, rulesFile } = pendingRef.current;
    try {
      const updated = await api.saveRules({ ...rulesFile, defaultBrowser: bp.browser, defaultProfile: bp.profile, options });
      appStore.update({ rulesFile: updated as any });
    } catch {}
  }, SAVE_DEBOUNCE);

  const defaultBrowserIsCustom = resolveBrowserIsCustom(bp.browserCustom, bp.browser, installedBrowsers);
  const defaultProfileIsCustom = resolveProfileIsCustom(bp.profileCustom, bp.profile, bp.browser, profilesByBrowser);

  function save() { if (!hasJsConfig) saveDebounced.flush(); }
  function scheduleSave() { if (!hasJsConfig) saveDebounced.schedule(); }

  function setOption<K extends keyof Options>(key: K, value: Options[K]) {
    setOptions((prev) => ({ ...prev, [key]: value }));
    pendingRef.current = { ...pendingRef.current, options: { ...pendingRef.current.options, [key]: value } };
    scheduleSave();
  }

  const { keepRunning, hideIcon, logRequests, checkForUpdates } = options;

  return (
    <PageContainer
      title={hasConfig ? "Configuration" : "No Valid Configuration Found"}
      description={hasConfig ? "Current settings from your configuration file" : undefined}
    >
      {!hasConfig && (
        <div className={styles.noConfigMessage}>
          <a href="https://github.com/johnste/finicky/wiki/Getting-started" target="_blank" rel="noopener noreferrer" className={styles.externalLink}>
            Learn how to get started
            <ExternalIcon />
          </a>
        </div>
      )}

      <div className={`${styles.section}${hasJsConfig ? " " + styles.readonly : ""}`}>
        <div className={styles.sectionHeader}>
          <span className={styles.sectionLabel}>Default browser</span>
          {hasJsConfig ? (
            <Tooltip text="JavaScript configuration file loaded — these settings can't be changed here">
              <span className={styles.lockInline}><LockIcon /></span>
            </Tooltip>
          ) : (
            <span className={styles.sectionHint}>Used when no rule matches</span>
          )}
        </div>
        <BrowserProfileSelector
          value={{ browser: bp.browser, profile: bp.profile }}
          custom={{ browser: defaultBrowserIsCustom, profile: defaultProfileIsCustom }}
          browsers={{ installed: installedBrowsers, profiles: profilesByBrowser }}
          disabled={hasJsConfig}
          onChange={(value, custom, committed) => {
            const next = { ...bp, ...value, browserCustom: custom.browser, profileCustom: custom.profile };
            setBp(next);
            pendingRef.current = { ...pendingRef.current, bp: next };
            if (committed) save(); else scheduleSave();
          }}
          onRequestProfiles={async (b) => { try { appStore.addBrowserProfiles(b, await api.getBrowserProfiles(b)); } catch {} }}
        />
      </div>

      <div className={styles.configOptions}>
        <div className={styles.optionsGrid}>
          <OptionRow label="Keep running" hint="App stays open in the background"
            checked={keepRunning} locked={hasJsConfig} onLockedClick={onLockedClick}
            onChange={(v) => setOption("keepRunning", v)} />
          <OptionRow label="Hide icon" hint="Hide menu bar icon"
            checked={hideIcon} locked={hasJsConfig} onLockedClick={onLockedClick}
            onChange={(v) => setOption("hideIcon", v)} />
          <OptionRow label="Log requests" hint="Log all URL handling to file"
            checked={logRequests} locked={hasJsConfig} onLockedClick={onLockedClick}
            onChange={(v) => setOption("logRequests", v)} />
          <OptionRow label="Check for updates" hint="Automatically check for new versions"
            checked={checkForUpdates} locked={hasJsConfig} onLockedClick={onLockedClick}
            onChange={(v) => setOption("checkForUpdates", v)} />
        </div>
      </div>

      {numErrors > 0 && (
        <div className={`${styles.statusCard} ${styles.error}`}>
          <h3>Errors</h3>
          <p>{numErrors} errors encountered.</p>
          <Link to="/troubleshoot">Troubleshooting</Link>
        </div>
      )}

      {updateInfo && <UpdateCard updateInfo={updateInfo} />}
    </PageContainer>
  );
}
