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
  checkForUpdate: boolean;
}

export interface ConfigInfo {
  configPath: string;
  handlers?: number;
  rewrites?: number;
  defaultBrowser?: string;
  keepRunning?: boolean;
  hideIcon?: boolean;
  logRequests?: boolean;
  checkForUpdate?: boolean;
}
