import { useSyncExternalStore } from "react";
import { PageContainer } from "../components/PageContainer";
import { LogContent } from "../components/LogContent";
import { appStore } from "../lib/appStore";
import { useLogControls } from "../lib/useLogControls";
import styles from "./LogViewer.module.css";

export function LogViewer() {
  const { messageBuffer } = useSyncExternalStore(appStore.subscribe, appStore.getSnapshot);
  const { showDebug, handleShowDebugChange, copyLogs } = useLogControls(messageBuffer);

  const controls = (
    <div className={styles.logControls}>
      <label className={styles.debugToggle}>
        <input type="checkbox" checked={showDebug} onChange={handleShowDebugChange} />
        Show debug
      </label>
      <button onClick={appStore.clearMessages}>Clear</button>
      <button onClick={copyLogs}>Copy</button>
    </div>
  );

  return (
    <PageContainer title="Logs" description={controls}>
      <LogContent messageBuffer={messageBuffer} showDebug={showDebug} />
    </PageContainer>
  );
}
