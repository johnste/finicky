import { parseUrl, matchHostnames } from "./createAPI";

/**
 * Finicky Configuration Types
 */

type LogFunction = (...messages: string[]) => void;
type NotifyFunction = (title: string, subtitle: string) => void;
type BatteryFunction = () => {
  chargePercentage: number;
  isCharging: boolean;
  isPluggedIn: boolean;
};
type SystemInfoFunction = () => {
  name: string;
  localizedName: string;
  address: string;
};

// Finicky Config API
export type ConfigAPI = {
  log: LogFunction;
  notify: NotifyFunction;
  getBattery: BatteryFunction;
  getSystemInfo: SystemInfoFunction;
  getUrlParts: typeof parseUrl;
  parseUrl: typeof parseUrl;
  getKeys(): KeyOptions;
  matchHostnames: typeof matchHostnames;
  matchDomains: typeof matchHostnames;
};

/**
 * This represents the full `.finicky.js` `module.exports` object.
 *
 * Example:
 *
 * ```js
 *  module.exports = {
 *    defaultBrowser: "Google Chrome",
 *    options: {
 *      hideIcon: false
 *    },
 *    handlers: [{
 *      match: finicky.matchHostnames("example.com'),
 *      browser: "Firefox"
 *    }]
 *  }
 * ```
 */
export interface FinickyConfig {
  /** The default browser or app to open for urls where no other handler
   * matches.
   */
  defaultBrowser: Browser | BrowserFunction | Array<Browser | BrowserFunction>;
  options?: {
    /* Whether or not to hide the finicky icon in the menu bar */
    hideIcon?: boolean;
    /**
     * An array of domain names to replace the built in list of url
     * shortener domains. Note that using this option replaces the list
     * completely.
     *
     * Alternatively a function that returns an array of domains
     */
    urlShorteners?: string[] | ((hostnames: string[]) => string[]);
    /**
     * Log all requests to console
     */
    logRequests?: boolean;
  };
  /* An array of Rewriters that can change the url being opened */
  rewrite?: Rewriter[];
  /* An array of Handlers to select which browser to open for urls */
  handlers?: Handler[];
}

export type BrowserResult =
  | Browser
  | BrowserFunction
  | Array<Browser | BrowserFunction>;

/**
 * A handler contains a matcher and a browser. If the matcher matches when opening a url, the browser in the handler will be opened.
 */
export interface Handler {
  match: Matcher | Matcher[];
  url?: PartialUrl | UrlFunction;
  browser: BrowserResult;
}

/**
 * A rewriter contains a matcher and a url. If the matcher matches when opening a url, the final url will be changed to whatever the url property is.
 */
export interface Rewriter {
  match: Matcher | Matcher[];
  url: PartialUrl | UrlFunction;
}

/**
 * Matches urls (or other properties) to decide if to change the url or open a browser.
 *
 * If the matcher is a string, if the url equals the string exactly, it will match.
 * If the matcher is a regular expression, if it matches any part of the url, it will match.
 * If the matcher is a [[MatcherFunction]], it will match if the function returns `true`
 *
 */
export type Matcher = string | RegExp | MatcherFunction;
export type MatcherFunction = (options: Options) => boolean;

/**
 * Represents a browser or app that finicky could start
 */
export type Browser = string | BrowserObject;

/**
 * Represents the type of app to start. "None" means to not start any app.
 */
export type AppType = "appName" | "bundleId" | "appPath" | "none";

/**
 * Represents a browser or app to open
 */
export interface BrowserObject {
  name: string;
  appType?: AppType;
  openInBackground?: boolean;
  profile?: string;
  args?: string[];
}

/**
 * A function that returns a browser to open
 */
type BrowserFunction = (options: Options) => Browser;

/**
 * Represents a url that will be handled by finicky
 */
export type Url = string | UrlObject;

/**
 * Represents a url that will be handled by finicky
 */
export type PartialUrl = string | Partial<UrlObject>;

/**
 * An object that represents a url
 */
export interface UrlObject {
  protocol: string;
  username?: string;
  password?: string;
  host: string;
  port?: number;
  pathname?: string;
  search?: string;
  hash?: string;
}

/**
 * An object that represents a key options object
 */
export interface KeyOptions {
  shift: boolean;
  option: boolean;
  command: boolean;
  control: boolean;
  capsLock: boolean;
  function: boolean;
}

/**
 * A function that returns a url
 */
export type UrlFunction = (options: Options) => PartialUrl;

export type Application = {
  pid: number;
  path?: string;
  bundleId?: string;
  name?: string;
};

/**
 * Options sent as the argument to [[ProcessUrl]]
 */
export interface ProcessOptions {
  /** The opening application */
  opener: Application;
}

/**
 * Options sent as the argument to [[MatcherFunction]], [[BrowserFunction]] and [[UrlFunction]]
 */
export interface Options extends ProcessOptions {
  /** The url being opened */
  urlString: string;
  /** The url being opened as an object */
  url: UrlObject;
  /** If opened from an app, this string contains the bundle identifier from that app
   * @deprecated Use opener.bundleId instead
   */
  sourceBundleIdentifier?: string;
  /** If opened from an app, this string contains the path to that app
   * @deprecated Use opener.path instead
   */
  sourceProcessPath?: string;
  /** The state of keyboard state. E.g. shift === true if pressed.
   * @deprecated Use finicky.getKeys() instead
   */
  keys: KeyOptions;
}
