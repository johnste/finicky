import "core-js/features/url";
import "core-js/features/url-search-params";
import {
  ConfigSchema,
  Config,
  OpenUrlOptions,
  ProcessInfo,
  BrowserSpecificationSchema,
  BrowserSpecification,
  UrlTransformSpecification,
  UrlMatcherPattern,
  BrowserConfig,
  BrowserConfigStrict,
  AppType,
} from "./configSchema";
import * as utilities from "./utilities";
import { matchWildcard } from "./wildcard";
import { fromError } from "zod-validation-error";
import { isLegacyURLObject, legacyURLObjectToString } from "./legacyURLObject";
import { FinickyURL } from "./FinickyURL";
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
    // Don't log parsing errors in test environment as they are expected
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
        "No default export found for configuration namespace, using legacy configuration. Please update your configuration to use `export default { ... }` instead of `module.exports = { ... }` syntax. https://github.com/johnste/finicky/wiki/Use-Modern-ECMAScript-Module-Syntax"
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

export function getConfigState(config: Config) {
  return {
    handlers: config.handlers?.length || 0,
    rewrites: config.rewrite?.length || 0,
    defaultBrowser:
      resolveBrowser(config.defaultBrowser, new URL("https://example.com"), {
        opener: null,
      })?.name || "None",
  };
}

export function openUrl(
  urlString: string,
  opener: ProcessInfo | null,
  config: object
) {
  if (!validateConfig(config)) {
    throw new Error("Invalid config");
  }

  let url = new FinickyURL(urlString, opener);

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

export function createBrowserConfig(
  browser: string | BrowserConfig | null
): Omit<BrowserConfigStrict, "url"> {
  const defaults = {
    appType: "appName" as const,
    openInBackground: undefined,
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

export function resolveBrowserInfo(
  browser: string
): Pick<BrowserConfigStrict, "name" | "appType" | "profile"> {
  const [name, profile] = browser.split(":");
  const appType = autodetectAppStringType(name);

  return { name, appType, profile: profile || "" };
}

export function resolveBrowser(
  browser: BrowserSpecification,
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
    BrowserSpecificationSchema.parse(config);

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

export function autodetectAppStringType(app: string | null): AppType {
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

export function rewriteUrl(
  rewrite: UrlTransformSpecification,
  url: URL | FinickyURL,
  options: OpenUrlOptions
): FinickyURL {
  if (rewrite instanceof FinickyURL) {
    return rewrite;
  }

  if (typeof rewrite === "string") {
    return new FinickyURL(rewrite, options.opener || null);
  }

  if (typeof rewrite === "function") {
    const result = rewrite(url, options);
    return rewriteUrl(result, url, options);
  }

  // Convert URL to FinickyURL if it's not already one
  if (rewrite instanceof URL) {
    return new FinickyURL(rewrite.href, options.opener || null);
  }

  if (isLegacyURLObject(rewrite)) {
    return new FinickyURL(
      legacyURLObjectToString(rewrite),
      options.opener || null
    );
  }

  return url as FinickyURL;
}

export function isMatch(
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
