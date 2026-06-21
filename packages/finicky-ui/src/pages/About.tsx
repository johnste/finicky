import { useSyncExternalStore } from "react";
import { PageContainer } from "../components/PageContainer";
import { ExternalIcon } from "../components/icons/External";
import { appStore } from "../lib/appStore";
import styles from "./About.module.css";

export function About() {
  const { version } = useSyncExternalStore(appStore.subscribe, appStore.getSnapshot);
  return (
    <PageContainer>
      <div className={styles.logoSection}>
        <div className={styles.appInfo}>
          <span className={styles.appName}>Finicky</span>
          <span className={styles.version}>Version {version}</span>
        </div>
      </div>

      <div className={styles.section}>
        <p>
          Finicky is a macOS application that lets you set up rules to decide which browser to open
          for every link.
        </p>
        <a
          href="https://github.com/johnste/finicky"
          target="_blank"
          rel="noopener noreferrer"
          className={styles.repoLink}
        >
          View on GitHub <ExternalIcon />
        </a>
      </div>

      <div className={styles.section}>
        <h2>Credits</h2>
        <p>
          Created by{" "}
          <a href="https://github.com/johnste" target="_blank" rel="noopener noreferrer">
            John Sterling
          </a>
        </p>
        <p>
          Icon designed by{" "}
          <a href="https://github.com/uetchy" target="_blank" rel="noopener noreferrer">
            @uetchy
          </a>
        </p>
        <p>
          <a
            href="https://github.com/johnste/finicky/graphs/contributors"
            target="_blank"
            rel="noopener noreferrer"
          >
            View all contributors
          </a>
        </p>
      </div>

      <div className={styles.section}>
        <p>
          If you find Finicky useful, consider{" "}
          <a href="https://github.com/sponsors/johnste" target="_blank" rel="noopener noreferrer">
            sponsoring the project
          </a>
          .
        </p>
      </div>
    </PageContainer>
  );
}
