import "core-js/features/url";
import "core-js/features/url-search-params";
import {
  ConfigSchema,
  SimpleConfig,
  Config,
  OpenUrlOptions,
  ProcessInfo,
  BrowserPattern,
  UrlPattern,
  UrlMatcherPattern,
  BrowserConfig,
  BrowserConfigStrict,
} from "./configSchema";
import * as utilities from "./utilities";

export { utilities };

export function validateConfig(config: object): config is Config {
  const result = ConfigSchema.safeParse(config);

  if (!result.success) {
    console.error(result.error);
  }

  return result.success;
}

export function mergeConfig(
  config: Config | undefined,
  simpleConfig: SimpleConfig | undefined
): Config | undefined {
  let convertedSimpleConfig: Config | undefined = undefined;
  if (simpleConfig) {
    const { defaultBrowser, ...handlers } = simpleConfig;
    convertedSimpleConfig = {
      defaultBrowser,
      handlers: Object.entries(handlers).map(([key, value]) => ({
        match: key,
        browser: value,
      })),
    };
  }

  if (config) {
    if (config.defaultBrowser && convertedSimpleConfig?.defaultBrowser) {
      console.warn(
        "defaultBrowser is set in both config and simpleConfig, using config value"
      );
    }
    return {
      ...config,
      handlers: [
        ...(config.handlers || []),
        ...(convertedSimpleConfig?.handlers || []),
      ],
    };
  } else if (convertedSimpleConfig) {
    return convertedSimpleConfig;
  }
}

export function getOption(option: keyof Config["options"], config: Config): unknown {
  return config.options && option in config.options ? config.options[option] : undefined;
}

export function openUrl(
  urlString: string,
  pid: number,
  opener: ProcessInfo | null,
  config: object
) {
  if (!validateConfig(config)) {
    throw new Error("Invalid config");
  }

  let url = new URL(urlString);

  const options: OpenUrlOptions = {
    pid,
    ...opener,
  };

  if (config.rewrite) {
    for (const rewrite of config.rewrite) {
      if (isMatch(rewrite.match, url, options)) {
        url = rewriteUrl(rewrite.url, url, options);
      }
    }
  }

  if (config.handlers) {
    for (const handler of config.handlers) {
      if (isMatch(handler.match, url, options)) {
        return resolveBrowser(handler.browser, url, options);
      }
    }
  }

  return resolveBrowser(config.defaultBrowser, url, options);
}

function resolveBrowser(
  browser: BrowserPattern,
  url: URL,
  options: OpenUrlOptions
): BrowserConfigStrict {
  const defaults = {
    appType: "name" as const,
    openInBackground: false,
    profile: "",
    args: [],
  };

  const createBrowserConfig = (
    browserName: string | BrowserConfig
  ): BrowserConfigStrict => {
    if (typeof browserName === "string") {
      return {
        ...defaults,
        name: browserName,
        url: url.href,
      };
    }
    return { ...defaults, ...browserName, url: url.href };
  };

  if (typeof browser === "object") {
    return createBrowserConfig(browser);
  }

  if (typeof browser === "function") {
    browser = browser(url, options);
    return createBrowserConfig(browser);
  }

  return createBrowserConfig(browser);
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
    console.log(typeof result, result instanceof URL);
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

    const urlWithoutProtocol = url.href.replace(/^\w+:(\/\/)?/, "");

    if (match.includes("*")) {
      return matchWildcard(match, urlWithoutProtocol);
    }

    return urlWithoutProtocol.includes(match);
  }

  if (match instanceof RegExp) {
    return match.test(url.href);
  }

  if (typeof match === "function") {
    return match(url, options);
  }

  return false;
}

function matchWildcard(pattern: string, str: string): boolean {
  try {
    // First handle escaped asterisks by temporarily replacing them
    const ESCAPED_ASTERISK_PLACEHOLDER = "\u0000";
    const patternWithEscapedAsterisks = pattern.replace(
      /\\\*/g,
      ESCAPED_ASTERISK_PLACEHOLDER
    );

    // Then escape all special regex chars except asterisk
    const escaped = patternWithEscapedAsterisks.replace(
      /[.+?^${}()|[\]\\]/g,
      "\\$&"
    );

    // Replace unescaped asterisks with .* for wildcard matching
    const regexPattern = escaped.replace(/\*/g, ".*");

    // Finally, restore the escaped asterisks
    const finalPattern = regexPattern.replace(
      new RegExp(ESCAPED_ASTERISK_PLACEHOLDER, "g"),
      "\\*"
    );

    return new RegExp(finalPattern).test(str);
  } catch (error) {
    console.warn("Invalid wildcard pattern:", pattern, error);
    return false;
  }
}
