import { useSyncExternalStore } from "react";
import { Link, useLocation } from "react-router-dom";
import { appStore } from "../lib/appStore";
import styles from "./TabBar.module.css";

export interface TabDef {
  path: string;
  label: string;
  Icon: React.ComponentType;
}

export interface BottomTabDef extends TabDef {
  showErrors?: boolean;
}

interface Props {
  tabs: TabDef[];
  bottomTabs: BottomTabDef[];
}

export function TabBar({ tabs, bottomTabs }: Props) {
  const numErrors = useSyncExternalStore(appStore.subscribe, appStore.getNumErrors);
  const location = useLocation();

  function isActive(path: string) {
    return location.pathname === path;
  }

  return (
    <div className={styles.tabContainer}>
      <nav className={styles.tabBar}>
        {tabs.map(({ path, label, Icon }) => (
          <Link key={path} to={path}>
            <div className={`${styles.tabContent}${isActive(path) ? " " + styles.active : ""}`}>
              <div className={styles.iconContainer}>
                <Icon />
              </div>
              <span>{label}</span>
            </div>
          </Link>
        ))}
        <div className={styles.spacer} />
        <div className={styles.bottomTabs}>
          {bottomTabs.map(({ path, label, Icon, showErrors }) => (
            <Link key={path} to={path}>
              <div className={`${styles.tabContent} ${styles.tabSmall}${isActive(path) ? " " + styles.active : ""}`}>
                <div className={styles.iconContainer}>
                  <Icon />
                  {showErrors && numErrors > 0 && (
                    <div className={styles.errorBadge}>{numErrors}</div>
                  )}
                </div>
                <span>{label}</span>
              </div>
            </Link>
          ))}
        </div>
      </nav>
    </div>
  );
}
