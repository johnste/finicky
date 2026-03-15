export interface Rule {
  match: string;
  browser: string;
  profile?: string;
}

export interface RulesFile {
  defaultBrowser: string;
  defaultProfile?: string;
  rules: Rule[];
}

export interface LogEntry {
  level: string;
  msg: string;
  time: string;
  error?: string;
  [key: string]: any; // Allow for additional dynamic fields
}

declare global {
  interface Window {
    finicky: {
      sendMessage: (msg: any) => void;
      receiveMessage: (msg: any) => void;
      /** Stub queue populated by the WKUserScript before the Svelte app is ready */
      _queue?: any[];
    };
    webkit?: {
      messageHandlers?: {
        finicky?: {
          postMessage: (msg: string) => void;
        };
      };
    };
  }
}

export interface UpdateInfo {
  version: string;
  hasUpdate: boolean;
  updateCheckEnabled: boolean;
  downloadUrl: string;
  releaseUrl: string;
}

export interface ConfigOptions {
  keepRunning: boolean;
  hideIcon: boolean;
  logRequests: boolean;
  checkForUpdates: boolean;
}

export interface ConfigInfo {
  configPath: string;
  handlers?: number;
  rewrites?: number;
  defaultBrowser?: string;
  options?: {
    keepRunning?: boolean;
    hideIcon?: boolean;
    logRequests?: boolean;
    checkForUpdates?: boolean;
  };
}
