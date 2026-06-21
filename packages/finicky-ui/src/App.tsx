import { useEffect } from "react";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import { TabBar } from "./components/TabBar";
import { Footer } from "./components/Footer";
import { ToastContainer } from "./components/ToastContainer";
import { StartPage } from "./pages/StartPage";
import { LogViewer } from "./pages/LogViewer";
import { TestUrl } from "./pages/TestUrl";
import { About } from "./pages/About";
import { Rules } from "./pages/Rules";
import { PreferencesIcon } from "./components/icons/Preferences";
import { RulesIcon } from "./components/icons/Rules";
import { TestIcon } from "./components/icons/Test";
import { LogsIcon } from "./components/icons/Logs";
import { AboutIcon } from "./components/icons/About";
import { api } from "./lib/api";
import { appStore } from "./lib/appStore";
import { toast } from "./lib/toast";
import { useSSE } from "./lib/useSSE";
import type { TabDef, BottomTabDef } from "./components/TabBar";
import styles from "./App.module.css";

const TABS: TabDef[] = [
  { path: "/", label: "Preferences", Icon: PreferencesIcon },
  { path: "/rules", label: "Rules", Icon: RulesIcon },
  { path: "/test", label: "Test", Icon: TestIcon },
];

const BOTTOM_TABS: BottomTabDef[] = [
  { path: "/troubleshoot", label: "Logs", Icon: LogsIcon, showErrors: true },
  { path: "/about", label: "About", Icon: AboutIcon },
];

export default function App() {
  useSSE();

  useEffect(() => {
    let cancelled = false;
    const MAX_ATTEMPTS = 3;

    async function load(attempt = 0) {
      try {
        const data = await api.initialData();
        if (cancelled) return;
        appStore.update({
          version: (data.version as string) ?? "v0.0.0",
          installedBrowsers: (data.installedBrowsers as string[]) ?? [],
          rulesFile: (data.rules as any) ?? { defaultBrowser: "", rules: [] },
          ...(data.config ? { hasConfig: true, config: data.config as any } : {}),
          ...(data.updateInfo ? { updateInfo: data.updateInfo as any } : {}),
        });
      } catch (err) {
        if (cancelled) return;
        if (attempt < MAX_ATTEMPTS) {
          setTimeout(() => load(attempt + 1), 500 * (attempt + 1));
        } else {
          toast.error(
            "Failed to load configuration",
            err instanceof Error ? err.message : String(err)
          );
        }
      }
    }
    load();

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <MemoryRouter>
      <main className={styles.main}>
        <div className={styles.layout}>
          <TabBar tabs={TABS} bottomTabs={BOTTOM_TABS} />
          <div className={styles.container}>
            <div className={styles.content}>
              <Routes>
                <Route path="/" element={<StartPage />} />
                <Route path="/troubleshoot" element={<LogViewer />} />
                <Route path="/test" element={<TestUrl />} />
                <Route path="/about" element={<About />} />
                <Route path="/rules" element={<Rules />} />
              </Routes>
            </div>
          </div>
        </div>
        <Footer />
      </main>
      <ToastContainer />
    </MemoryRouter>
  );
}
