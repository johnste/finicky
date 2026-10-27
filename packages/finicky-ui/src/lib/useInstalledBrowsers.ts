import { useEffect } from "react";
import { api } from "./api";
import { appStore } from "./appStore";

// The installed-browser list only otherwise comes from the one-time
// /api/initial-data fetch at app launch, and the WebView page is never
// reloaded for the life of the process, so without this a browser
// installed while Finicky is running would never show up.
export function useInstalledBrowsers() {
  useEffect(() => {
    api.getBrowsers().then((installed) => appStore.update({ installedBrowsers: installed })).catch(() => {});
  }, []);
}
