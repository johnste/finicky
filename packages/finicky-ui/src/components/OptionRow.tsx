import { LockIcon } from "./icons/Lock";
import { Tooltip } from "./Tooltip";
import styles from "./OptionRow.module.css";

const LOCKED_TOOLTIP = "JavaScript configuration file loaded — these settings can't be changed here";

interface Props {
  label: string;
  hint: string;
  checked: boolean;
  locked?: boolean;
  onLockedClick?: () => void;
  onChange?: (checked: boolean) => void;
}

function Inner({ label, hint, checked, locked, onChange }: Props) {
  return (
    <div className={styles.optionInfo}>
      <div className={styles.optionText}>
        <span className={styles.optionLabel}>{label}</span>
        <span className={styles.optionHint}>{hint}</span>
      </div>
      <label className={`${styles.toggle}${locked ? " " + styles.locked : ""}`}>
        <input
          type="checkbox"
          checked={checked}
          disabled={locked}
          onChange={(e) => onChange?.(e.target.checked)}
        />
        <span className={styles.toggleSlider} />
        {locked && (
          <span className={styles.nubLock}>
            <LockIcon />
          </span>
        )}
      </label>
    </div>
  );
}

export function OptionRow(props: Props) {
  const { locked, onLockedClick } = props;

  if (locked) {
    return (
      <Tooltip text={LOCKED_TOOLTIP} block>
        <button type="button" className={`${styles.optionRow} ${styles.locked}`} onClick={onLockedClick}>
          <Inner {...props} />
        </button>
      </Tooltip>
    );
  }

  return (
    <div className={styles.optionRow}>
      <Inner {...props} />
    </div>
  );
}
