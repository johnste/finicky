import {
  FinickyConfig,
  UrlObject,
  Options,
  Matcher,
  BrowserResult,
  Browser,
  UrlFunction,
  ProcessOptions,
  PartialUrl,
  ConfigAPI,
} from "./types";
import { urlSchema, appDescriptorSchema } from "./schemas";
import {
  createRegularExpression,
  guessAppType,
  composeUrl,
  validateSchema,
  deprecate,
} from "./utils";

// Assume the module.exports is available
declare const module: {
  exports?: FinickyConfig;
};

// The finicky api is available here
declare const finicky: ConfigAPI;

export function processUrl(
  config: FinickyConfig,
  url: string,
  processOptions: ProcessOptions
) {
  let options: Options = {
    urlString: url,
    url: finicky.parseUrl(url),
    keys: finicky.getKeys(),
    sourceBundleIdentifier: processOptions?.opener?.bundleId,
    sourceProcessPath: processOptions?.opener?.path,
    ...processOptions,
  };

  if (config?.options?.logRequests) {
    if (processOptions?.opener) {
      const app = processOptions.opener;
      finicky.log(
        `Opening ${url} from ${app.name || "N/A"}\n\tbundleId: ${
          app.bundleId || "N/A"
        }\n\tpath: ${app.path || "N/A"}`
      );
    } else {
      finicky.log(`Opening ${url} from an unknown application`);
    }
  }

  if (!config) {
    // If there's no config available use Safari as the browser
    return resolveBrowser("Safari", options);
  }

  // Rewrite the url
  options = processUrlRewrites(config, options);

  return processHandlers(config, options);
}

function processUrlRewrites(config: FinickyConfig, options: Options) {
  if (Array.isArray(config.rewrite)) {
    for (let rewrite of config.rewrite) {
      if (isMatch(rewrite.match, options)) {
        options = rewriteUrl(rewrite.url, options);
      }
    }
  }

  return options;
}

function processHandlers(config: FinickyConfig, options: Options) {
  if (Array.isArray(config.handlers)) {
    for (let handler of config.handlers) {
      if (isMatch(handler.match, options)) {
        if (handler.url) {
          options = rewriteUrl(handler.url, options);
        }
        return resolveBrowser(handler.browser, options);
      }
    }
  }

  return resolveBrowser(config.defaultBrowser, options);
}

function rewriteUrl(url: PartialUrl | UrlFunction, options: Options) {
  let urlResult = resolveUrl(url, options);
  validateSchema({ url: urlResult }, urlSchema);
  if (typeof urlResult === "string") {
    return {
      ...options,
      url: finicky.parseUrl(urlResult),
      urlString: urlResult,
    };
  }

  return {
    ...options,
    url: urlResult,
    urlString: composeUrl(urlResult),
  };
}

function isMatch(matcher: Matcher | Matcher[], options: Options) {
  if (!matcher) {
    return false;
  }

  const matchers = Array.isArray(matcher) ? matcher : [matcher];

  return matchers.some((matcher) => {
    if (matcher instanceof RegExp) {
      return matcher.test(options.urlString);
    } else if (typeof matcher === "string") {
      const regex = createRegularExpression(matcher);
      return regex.test(options.urlString);
    } else if (typeof matcher === "function") {
      // Add a deprecation warning when accessing certain deprecated properties
      const deprecatedOptions = deprecate(
        options,
        new Map([
          [
            "keys",
            "Use finicky.getKeys() instead, see https://github.com/johnste/finicky/wiki/Configuration#parameters",
          ],
          [
            "sourceBundleIdentifier",
            "Use opener.bundleId instead, see https://github.com/johnste/finicky/wiki/Configuration#parameters",
          ],
          [
            "sourceProcessPath",
            "Use opener.path instead, see https://github.com/johnste/finicky/wiki/Configuration#parameters",
          ],
        ])
      );

      return !!matcher(deprecatedOptions);
    }

    return false;
  });
}

// Recursively resolve handler to value
function resolveUrl(result: PartialUrl | UrlFunction, options: Options) {
  if (typeof result === "string") {
    return result;
  } else if (typeof result === "object") {
    return { ...options.url, ...result } as UrlObject;
  }

  const resolved = result(options);
  if (typeof resolved === "string") {
    return resolved;
  }

  return { ...options.url, ...resolved } as UrlObject;
}

function resolveBrowser(browser: BrowserResult, options: Options) {
  if (typeof browser === "function") {
    browser = browser(options);
  }

  if (!Array.isArray(browser)) {
    browser = [browser];
  }

  const browsers = browser.map(createBrowser);

  return { browsers, url: options.urlString };
}

function createBrowser(browser: Browser) {
  // If all we got was a string, try to figure out if it's a bundle identifier or an application name
  if (typeof browser === "string" || browser === null) {
    browser = {
      name: browser,
    };
  }

  if (typeof browser === "object" && !browser.appType) {
    const name = browser.name === null ? "" : browser.name;

    browser = {
      ...browser,
      name,
      appType: guessAppType(browser.name),
    };
  }

  validateSchema(browser, appDescriptorSchema);

  return browser;
}
