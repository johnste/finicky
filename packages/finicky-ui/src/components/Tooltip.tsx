import type { ReactNode } from "react";
import styles from "./Tooltip.module.css";

interface Props {
  children: ReactNode;
  text: string;
  block?: boolean;
}

export function Tooltip({ children, text, block }: Props) {
  return (
    <span className={`${styles.tooltipAnchor}${block ? " " + styles.block : ""}`}>
      {children}
      <span className={styles.tooltipText}>{text}</span>
    </span>
  );
}
