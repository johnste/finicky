import type { ConfigInfo, LogEntry, RulesFile, UpdateInfo } from "../types";

interface AppState {
  version: string;
  hasConfig: boolean;
  config: ConfigInfo;
  updateInfo: UpdateInfo | null;
  rulesFile: RulesFile;
  installedBrowsers: string[];
  profilesByBrowser: Record<string, string[]>;
  messageBuffer: LogEntry[];
}

const initialState: AppState = {
  version: "v0.0.0",
  hasConfig: false,
  config: { configPath: "" },
  updateInfo: null,
  rulesFile: { defaultBrowser: "", rules: [] },
  installedBrowsers: [],
  profilesByBrowser: {},
  messageBuffer: [],
};

function createAppStore() {
  let state = initialState;
  let listeners: (() => void)[] = [];

  function notify() {
    for (const l of listeners) l();
  }

  return {
    getSnapshot: () => state,
    getNumErrors: () =>
      state.messageBuffer.filter((m) => m.level.toLowerCase() === "error").length,
    subscribe(listener: () => void) {
      listeners.push(listener);
      return () => {
        listeners = listeners.filter((l) => l !== listener);
      };
    },
    update(patch: Partial<AppState>) {
      state = { ...state, ...patch };
      notify();
    },
    addBrowserProfiles(browser: string, profiles: string[]) {
      state = { ...state, profilesByBrowser: { ...state.profilesByBrowser, [browser]: profiles } };
      notify();
    },
    appendMessage(entry: LogEntry) {
      state = { ...state, messageBuffer: [...state.messageBuffer, entry] };
      notify();
    },
    clearMessages() {
      state = { ...state, messageBuffer: [] };
      notify();
    },
  };
}

export const appStore = createAppStore();
