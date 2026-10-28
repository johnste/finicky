import type { ReactNode } from "react";
import clsx from "clsx";
import styles from "./Tooltip.module.css";

interface Props {
  children: ReactNode;
  text: string;
  block?: boolean;
}

export function Tooltip({ children, text, block }: Props) {
  return (
    <span className={clsx(styles.tooltipAnchor, block && styles.block)}>
      {children}
      <span className={styles.tooltipText}>{text}</span>
    </span>
  );
}
