import "core-js/features/url";
import "core-js/features/url-search-params";
import {
  ConfigSchema,
  Config,
  OpenUrlOptions,
  ProcessInfo,
  BrowserPatternSchema,
  BrowserPattern,
  UrlPattern,
  UrlMatcherPattern,
  BrowserConfig,
  BrowserConfigStrict,
  AppType,
} from "./configSchema";
import * as utilities from "./utilities";
import { matchWildcard } from "./wildcard";
import { fromError } from "zod-validation-error";

export { utilities };

export function validateConfig(config: object): config is Config {
  if (!config) {
    console.error(
      "Could not find configuration object, please check your config file and make sure it has a default export."
    );
    return false;
  }

  try {
    ConfigSchema.parse(config);
    return true;
  } catch (ex) {
    if (process.env.NODE_ENV !== "test") {
      console.error(fromError(ex).toString());
    }
    return false;
  }
}

export function getConfiguration(namespace: string): Config {
  const namespaceObj = (self as any)[namespace];
  if (namespaceObj) {
    if (namespaceObj.default) {
      return namespaceObj.default;
    } else {
      console.warn(
        "No default export found for configuration namespace, assuming legacy config"
      );
      return namespaceObj;
    }
  }

  throw new Error(
    "Could not find configuration object, please check your config file and make sure it has a default export."
  );
}

export function getOption(
  option: keyof Config["options"],
  config: Config
): unknown {
  return config.options && option in config.options
    ? config.options[option]
    : undefined;
}

export function openUrl(
  urlString: string,
  opener: ProcessInfo | null,
  config: object
) {
  if (!validateConfig(config)) {
    throw new Error("Invalid config");
  }

  let url = createUrlWithProxy(urlString, opener);

  const options: OpenUrlOptions = {
    opener: opener,
  };

  let error: string | undefined;

  try {
    if (config.rewrite) {
      for (const rewrite of config.rewrite) {
        if (isMatch(rewrite.match, url, options)) {
          url = rewriteUrl(rewrite.url, url, options);
        }
      }
    }

    if (config.handlers) {
      for (const [index, handler] of config.handlers.entries()) {
        if (isMatch(handler.match, url, options)) {
          return {
            browser: resolveBrowser(handler.browser, url, options),
          };
        }
      }
    }
  } catch (ex: unknown) {
    error = ex instanceof Error ? ex.message : String(ex);
  }

  const browser = resolveBrowser(config.defaultBrowser, url, options);

  return {
    browser,
    error,
  };
}

export function getConfigInfo(config: Config): string {
  const numHandlers = config.handlers?.length || 0;
  const numRewrites = config.rewrite?.length || 0;
  return `Config loaded with ${numHandlers} handler${
    numHandlers === 1 ? "" : "s"
  } and ${numRewrites} rewrite${numRewrites === 1 ? "" : "s"}`;
}

function createBrowserConfig(
  browser: string | BrowserConfig | null
): Omit<BrowserConfigStrict, "url"> {
  const defaults = {
    appType: "appName" as const,
    openInBackground: false,
    profile: "",
    args: [],
  };

  if (browser === null) {
    return {
      ...defaults,
      name: "",
      appType: "none" as const,
    };
  }

  if (typeof browser === "string") {
    const browserInfo = resolveBrowserInfo(browser);

    return {
      ...defaults,
      ...browserInfo,
    };
  }

  return { ...defaults, ...browser };
}

function resolveBrowserInfo(
  browser: string
): Pick<BrowserConfigStrict, "name" | "appType" | "profile"> {
  const [name, profile] = browser.split(":");
  const appType = autodetectAppStringType(name);

  return { name, appType, profile: profile || "" };
}

function resolveBrowser(
  browser: BrowserPattern,
  url: URL | FinickyURL,
  options: OpenUrlOptions
): BrowserConfigStrict {
  const config =
    typeof browser === "function" ? browser(url, options) : browser;

  if (config === undefined) {
    throw new Error(
      JSON.stringify(
        {
          message: "Browser config cannot be undefined",
          error: "Browser config must be a string, object, or null",
        },
        null,
        2
      )
    );
  }

  try {
    BrowserPatternSchema.parse(config);

    const browserConfig = createBrowserConfig(config);
    const finalConfig = { ...browserConfig, url: url.href };

    return finalConfig;
  } catch (ex: unknown) {
    throw new Error(
      JSON.stringify(
        {
          message: "Invalid browser option",
          browser: config,
          error: fromError(ex).toString(),
        },
        null,
        2
      )
    );
  }
}

function autodetectAppStringType(app: string | null): AppType {
  let appType: AppType = "appName";

  if (app === null) {
    appType = "none";
  } else if (/^[a-zA-Z0-9 ]+$/.test(app)) {
    appType = "appName";
  }

  // The bundle ID string must contain only alphanumeric characters (A-Z, a-z, 0-9), hyphen (-), and period (.).
  // https://help.apple.com/xcode/mac/current/#/deve70ea917b
  else if (/^[a-zA-Z0-9.-]+$/.test(app)) {
    appType = "bundleId";
  }

  // The app path should be an absolute path to an app, e.g. /Applications/Google Chrome.app or relative to
  // the user's home directory, e.g. ~/Applications/Google Chrome.app
  else if (/^(~?(?:\/[^/\n]+)+\/[^/\n]+\.app)$/.test(app)) {
    appType = "path";
  }

  return appType;
}

/**
 * Legacy FinickyURL type used for backward compatibility
 */
export type FinickyURLObject = {
  username?: string;
  host: string;
  protocol?: string;
  pathname?: string;
  search?: string;
  password?: string;
  port?: number;
  hash?: string;
};

/**
 * FinickyURL class that extends URL to maintain backward compatibility
 * with legacy properties while providing deprecation warnings.
 */
class FinickyURL extends URL {
  private _opener: ProcessInfo | null;

  constructor(url: string, opener: ProcessInfo | null = null) {
    super(url);
    this._opener = opener;
  }

  get urlString(): string {
    console.warn(
      'Accessing legacy property "urlString" that is no longer supported. Please use the URL object\'s href property directly instead.'
    );
    return this.href;
  }

  get url(): FinickyURLObject {
    console.warn(
      'Accessing legacy property "url" that is no longer supported. Please use the URL object directly instead.'
    );
    return URLtoFinickyURL(this);
  }

  get opener(): ProcessInfo | null {
    console.warn(
      'Accessing legacy property "opener" that is no longer supported.'
    );
    return this._opener;
  }

  get keys() {
    throw new Error(
      'Accessing legacy property "keys" that is no longer supported, please use finicky.getModifierKeys() instead.'
    );
  }
}

function rewriteUrl(
  rewrite: UrlPattern,
  url: URL | FinickyURL,
  options: OpenUrlOptions
): FinickyURL {
  if (typeof rewrite === "string") {
    return createUrlWithProxy(rewrite, options.opener || null);
  }

  if (typeof rewrite === "function") {
    const result = rewrite(url, options);
    return rewriteUrl(result, url, options);
  }
  // Convert URL to FinickyURL if it's not already one
  if (!(url instanceof FinickyURL)) {
    return new FinickyURL(url.href, options.opener || null);
  }

  return url as FinickyURL;
}

function isMatch(
  match: UrlMatcherPattern,
  url: URL | FinickyURL,
  options: OpenUrlOptions
): boolean {
  if (Array.isArray(match)) {
    return match.some((m) => isMatch(m, url, options));
  }

  if (typeof match === "string") {
    if (match === "") {
      return false; // Empty string should not match anything
    }

    return matchWildcard(match, url.href);
  }

  if (match instanceof RegExp) {
    return match.test(url.href);
  }

  if (typeof match === "function") {
    return match(url, options);
  }

  return false;
}

/**
 * Creates a URL object with backward compatibility properties
 */
function createUrlWithProxy(
  url: string | FinickyURLObject,
  opener: ProcessInfo | null
): FinickyURL {
  let urlString: string;

  // Convert non-string URL to a string representation
  if (typeof url !== "string") {
    urlString = finicky_URLObjectToString(url);
  } else {
    urlString = url;
  }

  return new FinickyURL(urlString, opener);
}

/**
 * Converts a FinickyURLObject to a URL string
 * @param urlObj The FinickyURLObject to convert
 * @returns A URL string representation
 */
export function finicky_URLObjectToString(urlObj: FinickyURLObject): string {
  return `${urlObj.protocol ? urlObj.protocol.replace(":", "") : "https"}://${
    urlObj.username
  }${urlObj.password ? ":" + urlObj.password : ""}${
    urlObj.username || urlObj.password ? "@" : ""
  }${urlObj.host}${urlObj.port ? ":" + urlObj.port : ""}${urlObj.pathname}${
    urlObj.search ? "?" + urlObj.search : ""
  }${urlObj.hash ? "#" + urlObj.hash : ""}`;
}

/**
 * Converts a standard URL object to a FinickyURL format
 * @param url The URL object to convert
 * @returns A FinickyURL object
 */
export function URLtoFinickyURL(url: URL): FinickyURLObject {
  return {
    username: url.username,
    host: url.hostname,
    protocol: url.protocol.replace(":", ""),
    pathname: url.pathname,
    search: url.search.replace("?", ""),
    password: url.password,
    port: url.port ? parseInt(url.port) : undefined,
    hash: url.hash.replace("#", ""),
  };
}
