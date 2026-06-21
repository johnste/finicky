import { useSyncExternalStore } from "react";
import { toast } from "../lib/toast";
import type { Toast } from "../lib/toast";
import styles from "./ToastContainer.module.css";

const typeClass: Record<Toast["type"], string> = {
  success: styles.toastSuccess,
  error: styles.toastError,
  warning: styles.toastWarning,
  info: styles.toastInfo,
};

export function ToastContainer() {
  const toasts = useSyncExternalStore(toast.subscribe, toast.getSnapshot);

  return (
    <div className={styles.toastContainer}>
      {toasts.map((t) => (
        <div key={t.id} className={`${styles.toastItem} ${typeClass[t.type]}`} role="alert">
          <div className={styles.toastContent}>
            <svg className={styles.toastTimer} viewBox="0 0 24 24">
              <circle className={styles.timerBackground} cx="12" cy="12" r="10" />
              <circle
                key={t.key}
                className={styles.timerProgress}
                cx="12"
                cy="12"
                r="10"
                style={{ animationDuration: `${t.duration}ms` }}
              />
            </svg>
            <span className={styles.toastMessage}>{t.message}</span>
          </div>
          {t.extra && <pre className={styles.toastExtra}>{t.extra}</pre>}
        </div>
      ))}
    </div>
  );
}
