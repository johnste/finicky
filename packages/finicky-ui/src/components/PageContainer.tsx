import type { ReactNode } from "react";
import styles from "./PageContainer.module.css";

interface Props {
  children: ReactNode;
  center?: boolean;
  title?: string;
  description?: ReactNode;
}

export function PageContainer({ children, center, title, description }: Props) {
  return (
    <div className={`${styles.pageContainer}${center ? " " + styles.center : ""}`}>
      {title && (
        <div className={styles.headerSection}>
          <h2>{title}</h2>
          {description && <p className={styles.pageDescription}>{description}</p>}
        </div>
      )}
      {children}
    </div>
  );
}
