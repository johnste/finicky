import urlParse from "url-parse";
import { ISchema, IValidator } from "./fastidious/types";
import { getErrors, validate } from "./fastidious/index";
import {
  Url,
  FinickyConfig,
  UrlObject,
  Options,
  Matcher,
  urlSchema,
  BrowserResult,
  Browser,
  UrlFunction,
  ProcessOptions,
  Rewriter,
  PartialUrl,
} from "./types";
import { createRegularExpression } from './utils';


declare const module:
  | {
    exports?: FinickyConfig;
  }
  | undefined;

declare const finicky: {
  matchDomains(hostnames: string | string[]): boolean;
  matchHostnames(hostnames: string | string[]): boolean;
  log(value: string): void;
  notify(title: string, subtitle?: string): void;
  getUrlParts(url: string): UrlObject;
};

const appDescriptorSchema = {
  name: validate.string,
  appType: validate.oneOf([
    validate.value("bundleId"),
    validate.value("appName"),
    validate.value("appPath"),
    validate.value("none"),
  ]).isRequired,
  openInBackground: validate.boolean,
  profile: validate.string,
  args: validate.arrayOf(validate.string),
};

export function processUrl(
  config: FinickyConfig,
  url: string,
  processOptions?: ProcessOptions
) {
  if (!processOptions) {
    processOptions = {
      keys: {
        capsLock: false,
        command: false,
        shift: false,
        option: false,
        control: false,
        function: false,
      },
    };
  }

  let options = {
    urlString: url,
    url: finicky.getUrlParts(url),
    ...processOptions,
  };

  if (!config) {
    return processBrowserResult("Safari", options);
  }

  options = processUrlRewrites(config, options);

  if (Array.isArray(config.handlers)) {
    for (let handler of config.handlers) {
      if (isMatch(handler.match, options)) {
        if (handler.url) {
          options = rewriteUrl(handler.url, options)
        }
        return processBrowserResult(handler.browser, options);
      }
    }
  }

  return processBrowserResult(config.defaultBrowser, options);
}

function validateSchema(
  value: unknown,
  schema: ISchema | IValidator,
  path = ""
) {
  const errors = getErrors(value, schema, path);
  if (errors.length > 0) {
    throw new Error(
      errors.join("\n") + "\nReceived value: " + JSON.stringify(value, null, 2)
    );
  }
}

function createUrl(url: UrlObject) {
  const { protocol, host, pathname = "" } = url;
  let port = url.port ? `:${url.port}` : "";
  let search = url.search ? `?${url.search}` : "";
  let hash = url.hash ? `#${url.hash}` : "";
  let auth = url.username ? `${url.username}` : "";
  auth += url.password ? `:${url.password}` : "";

  return `${protocol}://${auth}${host}${port}${pathname}${search}${hash}`;
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

function rewriteUrl(url: PartialUrl | UrlFunction, options: Options) {

  let urlResult = resolveUrl(url, options);
  validateSchema({ url: urlResult }, urlSchema);
  if (typeof urlResult === "string") {
    options = {
      ...options,
      url: finicky.getUrlParts(urlResult),
      urlString: urlResult,
    };
  }
  else {
    options = {
      ...options,
      url: urlResult,
      urlString: createUrl(urlResult),
    };
  }
  return options;
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
      if (!regex) {
        return false;
      }
      return regex.test(options.urlString);
    } else if (typeof matcher === "function") {
      return !!matcher(options);
    }

    return false;
  });
}

// Recursively resolve handler to value
function resolveBrowser(result: BrowserResult, options: Options) {
  if (typeof result !== "function") {
    return result;
  }

  return result(options);
}

// Recursively resolve handler to value
function resolveUrl(result: PartialUrl | UrlFunction, options: Options) {
  if (typeof result === "string") {
    return result;
  }
  else if (typeof result === "object") {
    return { ...options.url, ...result } as UrlObject;
  }

  const resolved = result(options);
  if (typeof resolved === "string") {
    return resolved;
  }

  return { ...options.url, ...resolved } as UrlObject;
}

function getAppType(value: string) {
  if (value === null) {
    return "none";
  }

  if (looksLikeBundleIdentifier(value)) {
    return "bundleId";
  }

  if (looksLikeAbsolutePath(value)) {
    return "appPath";
  }

  return "appName";
}

function processBrowserResult(result: BrowserResult, options: Options) {
  let browser = resolveBrowser(result, options);

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
      appType: getAppType(browser.name),
    };
  }

  validateSchema(browser, appDescriptorSchema);

  return browser;
}

function looksLikeBundleIdentifier(value: string) {
  // Regular expression to match Uniform Type Identifiers
  // Adapted from https://stackoverflow.com/a/34241710/1698327
  const bundleIdRegex = /^[A-Za-z]{2,6}((?!-)\.[A-Za-z0-9-]{1,63})+$/;
  if (bundleIdRegex.test(value)) {
    return true;
  }
  return false;
}

function looksLikeAbsolutePath(value: string) {
  return value.startsWith("/") || value.startsWith("~");
}
