import { useState, useEffect } from "react";
import { PageContainer } from "../components/PageContainer";
import { LinkIcon } from "../components/icons/Link";
import { InfoIcon } from "../components/icons/Info";
import { SpinnerIcon } from "../components/icons/Spinner";
import { api } from "../lib/api";
import type { TestUrlResult } from "../types";
import styles from "./TestUrl.module.css";

function isValidUrl(url: string): boolean {
  if (!url.trim()) return false;
  try {
    new URL(url.includes("://") ? url : `https://${url}`);
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

function ResultItem({ label, value, itemClass, valueClass }: {
  label: string;
  value: React.ReactNode;
  itemClass?: string;
  valueClass?: string;
}) {
  return (
    <div className={`${styles.resultItem}${itemClass ? " " + itemClass : ""}`}>
      <span className={styles.resultLabel}>{label}</span>
      <span className={`${styles.resultValue}${valueClass ? " " + valueClass : ""}`}>{value}</span>
    </div>
  );
}

// fallow-ignore-next-line complexity
function UrlStateView({ testUrl }: { testUrl: string }) {
  if (testUrl.trim() && !isValidUrl(testUrl)) {
    return (
      <div className={styles.hintMessage}>
        <InfoIcon />
        Enter a valid URL to see how Finicky will handle it
      </div>
    );
  }
  if (!testUrl.trim()) {
    return (
      <div className={styles.emptyState}>
        <LinkIcon />
        <p>Enter a URL above to test your configuration</p>
      </div>
    );
  }
  return null;
}

export function TestUrl() {
  const [testUrl, setTestUrl] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<TestUrlResult | null>(null);

  useEffect(() => {
    if (!isValidUrl(testUrl)) {
      setResult(null);
      setLoading(false);
      return;
    }

    const loadingTimer = setTimeout(() => setLoading(true), LOADING_DELAY);
    const debounceTimer = setTimeout(async () => {
      try {
        setResult((await api.testUrl(normalizeUrl(testUrl))) as TestUrlResult);
      } catch {
        setResult(null);
      } finally {
        setLoading(false);
      }
    }, DEBOUNCE_DELAY);

    return () => {
      clearTimeout(loadingTimer);
      clearTimeout(debounceTimer);
    };
  }, [testUrl]); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <PageContainer
      title="Test"
      description="Test how Finicky will handle a URL based on your current configuration"
    >
      <div className={styles.inputSection}>
        <label htmlFor="url-input" className={styles.inputLabel}>
          Enter URL
          {loading && (
            <div className={styles.loadingSpinner}>
              <SpinnerIcon />
            </div>
          )}
        </label>
        <input
          id="url-input"
          type="text"
          className={styles.urlInput}
          placeholder="https://example.com"
          autoCapitalize="off"
          autoCorrect="off"
          required
          value={testUrl}
          onChange={(e) => setTestUrl(e.target.value)}
        />
      </div>

      {result ? (
        <div className={styles.resultSection}>
          <div className={styles.resultHeader}>
            <h3>Result</h3>
          </div>
          <div className={styles.resultGrid}>
            <ResultItem label="Browser" value={result.browser} valueClass={styles.browser} />
            <ResultItem label="Profile" value={result.profile || "N/A"} />
            {typeof result.openInBackground === "boolean" && (
              <ResultItem label="Open in background" value={result.openInBackground ? "Yes" : "No"} />
            )}
            <ResultItem label="Final URL" value={result.url} itemClass={styles.fullWidth} valueClass={styles.url} />
          </div>
        </div>
      ) : (
        <UrlStateView testUrl={testUrl} />
      )}
    </PageContainer>
  );
}
