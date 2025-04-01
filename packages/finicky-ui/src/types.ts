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
