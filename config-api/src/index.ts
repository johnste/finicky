import "core-js/features/url";
import "core-js/features/url-search-params";
import {
  ConfigSchema,
  Config,
  OpenUrlOptions,
  ProcessInfo,
  BrowserConfigStrictSchema,
  BrowserPattern,
  UrlPattern,
  UrlMatcherPattern,
  BrowserConfig,
  BrowserConfigStrict,
  AppType,
} from "./configSchema";
import * as utilities from "./utilities";
import { matchWildcard } from "./wildcard";

export { utilities };

export function validateConfig(config: object): config is Config {
  const result = ConfigSchema.safeParse(config);

  if (!result.success && process.env.NODE_ENV !== "test") {
    console.error(result.error);
  }

  return result.success;
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

  let url = new URL(urlString);

  const options: OpenUrlOptions = {
    opener: opener,
  };

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
        return resolveBrowser(handler.open, url, options);
      }
    }
  }

  return resolveBrowser(config.defaultBrowser, url, options);
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
  url: URL,
  options: OpenUrlOptions
): BrowserConfigStrict {
  const config =
    typeof browser === "function" ? browser(url, options) : browser;

  const browserConfig = createBrowserConfig(config);

  const finalConfig = { ...browserConfig, url: url.href };
  const result = BrowserConfigStrictSchema.safeParse(finalConfig);

  if (!result.success) {
    throw new Error(`Invalid browser config: ${result.error}`);
  }

  return finalConfig;
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

  if (process.env.NODE_ENV !== "test") {
    console.log(`App value "${app}" autodetected as type "${appType}"`);
  }

  return appType;
}

function rewriteUrl(
  rewrite: UrlPattern,
  url: URL,
  options: OpenUrlOptions
): URL {
  if (typeof rewrite === "string") {
    return new URL(rewrite);
  }

  if (typeof rewrite === "function") {
    const result = rewrite(url, options);
    return new URL(result);
  }

  return url;
}

function isMatch(
  match: UrlMatcherPattern,
  url: URL,
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
