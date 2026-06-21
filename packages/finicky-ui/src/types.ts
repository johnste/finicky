export interface BrowserProfile {
  browser: string;
  profile: string;
}

export interface BrowserOptions {
  installed: string[];
  profiles: Record<string, string[]>;
}

export interface BrowserProfileCustom {
  browser: boolean;
  profile: boolean;
}

export interface Rule {
  match: string[];
  browser: string;
  profile?: string;
}

export interface ConfigOptions {
  keepRunning: boolean;
  hideIcon: boolean;
  logRequests: boolean;
  checkForUpdates: boolean;
}

export interface RulesFile {
  defaultBrowser: string;
  defaultProfile?: string;
  options?: Partial<ConfigOptions>;
  rules: Rule[];
  path?: string;
}

export interface LogEntry {
  level: string;
  msg: string;
  time: string;
  error?: string;
  [key: string]: any;
}

export interface TestUrlResult {
  browser: string;
  url: string;
  openInBackground: boolean;
  profile?: string;
}

declare global {
  interface Window {
    __FINICKY_API__?: string;
  }
}

export interface UpdateInfo {
  version: string;
  hasUpdate: boolean;
  updateCheckEnabled: boolean;
  downloadUrl: string;
  releaseUrl: string;
}

export interface ConfigInfo {
  configPath: string;
  hasJsConfig?: boolean;
  handlers?: number;
  rewrites?: number;
  defaultBrowser?: string;
  options?: Partial<ConfigOptions>;
}
