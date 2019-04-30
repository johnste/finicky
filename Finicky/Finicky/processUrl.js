// if (typeof options === "object") {
//   validateObject(
//     options,
//     {},
//     {
//       openInBackground: "boolean"
//     }
//   );
// }

(function() {
  return function processUrl(urlString, urlObject) {
    for (handler of module.exports.handlers) {
      const match = getMatch(handler.match, urlString, urlObject);
      if (match) {
        return processBrowserResult(handler.browser, urlString, urlObject);
      }
    }

    if (module.exports.defaultBrowser) {
      return processBrowserResult(
        module.exports.defaultBrowser,
        urlString,
        urlObject
      );
    }
  };

  function getMatch(matcher, urlString, urlObject) {
    if (!matcher) {
      return false;
    }

    const matchers = Array.isArray(matcher) ? matcher : [matcher];

    return matchers.some(matcher => {
      if (matcher instanceof RegExp) {
        return matcher.test(urlString);
      }

      if (typeof matcher === "string") {
        return matcher === urlString;
      }

      if (typeof matcher === "function") {
        return !!matcher(urlString, urlObject);
      }

      return false;
    });
  }

  function processBrowserResult(browser, urlString, urlObject) {
    if (typeof browser === "function") {
      browser = browser(urlString, urlObject);
    }

    // If all we got was a string, try to figure out if it's a bundle identifier or an application name
    if (typeof browser === "string") {
      const type = isBundleIdentifier(browser) ? "bundleId" : "appName";

      return {
        value: browser,
        type
      };
    }

    if (typeof browser === "object") {
      if (typeof browser.type === "undefined") {
        browser.type = isBundleIdentifier(browser.value)
          ? "bundleId"
          : "appName";
      }

      validateObject(
        browser,
        {
          value: "string",
          type: "string"
        },
        {
          url: "string",
          options: "object"
        }
      );

      return browser;
    }

    throw new Error(
      "Unrecognized result value, expected type [string or object], found " +
        JSON.stringify(browser)
    );
  }

  function isBundleIdentifier(value) {
    // Regular expression to match Uniform Type Identifiers
    // Adapted from https://stackoverflow.com/a/34241710/1698327
    const bundleIdRegex = /^[A-Za-z]{2,6}((?!-)\.[A-Za-z0-9-]{1,63})+$/;
    if (bundleIdRegex.test(value)) {
      return true;
    }
    return false;
  }

  function validateObject(value, required = {}, optional = {}) {
    if (typeof value !== "object") {
      throw new Error("Expected object value");
    }

    const keys = Object.keys(value);
    const requiredKeys = Object.keys(required);
    const optionalKeys = Object.keys(optional);
    const allProperties = {
      ...optional,
      ...required
    };

    requiredKeys.forEach(key => {
      if (!keys.includes(key)) {
        throw new Error(
          `Required key "${key}" missing in "${JSON.stringify(value)}"`
        );
      }
    });

    keys.forEach(key => {
      if (![...requiredKeys, ...optionalKeys].includes(key)) {
        throw new Error(`Unknown key "${key}" in "${JSON.stringify(value)}"`);
      }

      const expectedType = allProperties[key];
      const actualType = typeof value[key];

      if (expectedType !== actualType) {
        throw new Error(
          `Wrong type for key "${key}", found "${actualType}", expected "${expectedType}"`
        );
      }
    });
  }
})();
