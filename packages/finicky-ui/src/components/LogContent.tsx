import type { LogEntry } from "../types";
import { formatLogEntry } from "../utils/text";
import styles from "./LogContent.module.css";

interface Props {
  messageBuffer: LogEntry[];
  showDebug: boolean;
}

const levelClassMap: Record<string, string> = {
  error: styles.logLevelError,
  warn: styles.logLevelWarn,
  debug: styles.logLevelDebug,
};

export function LogContent({ messageBuffer, showDebug }: Props) {
  return (
    <ol className={styles.logContent}>
      {messageBuffer
        .filter((entry) => showDebug || entry.level.toLowerCase() !== "debug")
        .map((entry, i) => (
          <li key={i} className={styles.logEntry}>
            <span className={styles.logTime} title={entry.time}>
              {new Date(entry.time).toLocaleTimeString()}
            </span>
            <div className={`${styles.logMessage} ${levelClassMap[entry.level.toLowerCase()] ?? ""}`}>
              {formatLogEntry(entry).map((part, j) =>
                part.type === "url" ? (
                  <a key={j} href={part.content} target="_blank" rel="noopener noreferrer">
                    {part.content}
                  </a>
                ) : (
                  <span key={j}>{part.content}</span>
                )
              )}
            </div>
          </li>
        ))}
    </ol>
  );
}
