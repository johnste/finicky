import { useSyncExternalStore } from "react";
import { ExternalIcon } from "./icons/External";
import { toast } from "../lib/toast";
import { appStore } from "../lib/appStore";
import styles from "./Footer.module.css";

function basename(path: string): string {
  return path.split("/").pop() || path;
}

function showPathToast(label: string, description: string) {
  toast.show(label, "info", description, 5000);
}

export function Footer() {
  const { version, hasConfig, config, rulesFile, updateInfo } = useSyncExternalStore(
    appStore.subscribe,
    appStore.getSnapshot
  );

  return (
    <div className={styles.footer}>
      {hasConfig || rulesFile.path ? (
        <>
          <span className={styles.configLabel}>Config loaded:</span>
          {hasConfig && (
            <button
              className={styles.configBadge}
              onClick={() => showPathToast(basename(config.configPath), "Configuration loaded from\n${config.configPath}")}
            >
              ✓ {basename(config.configPath)}
            </button>
          )}
          {rulesFile.path && rulesFile.path !== config.configPath && (
            <button
              className={styles.configBadge}
              onClick={() => showPathToast(basename(rulesFile.path!), "Configuration loaded from\n${rulesFile.path}")}
            >
              ✓ {basename(rulesFile.path)}
            </button>
          )}
        </>
      ) : (
        <>
          <span className={`${styles.configStatus} ${styles.warning}`}>No config</span>
          <a
            href="https://github.com/johnste/finicky/wiki/Getting-started"
            target="_blank"
            rel="noopener noreferrer"
            className={styles.configLink}
          >
            Get started
            <ExternalIcon />
          </a>
        </>
      )}
      <span className={styles.spacer} />
      {updateInfo && updateInfo.updateCheckEnabled && !updateInfo.hasUpdate && (
        <span className={styles.upToDate}>✓ Up to date</span>
      )}
      <span className={styles.version}>{version}</span>
    </div>
  );
}
