(function() {
  function isBundleIdentifier(value) {
    // Regular expression to match Uniform Type Identifiers
    // Adapted from https://stackoverflow.com/a/34241710/1698327
    const bundleIdRegex = /^[A-Za-z]{2,6}((?!-)\.[A-Za-z0-9-]{1,63})+$/;
    if (bundleIdRegex.test(value)) {
      return true;
    }
    return false;
  }

  function resolveMatchers(matchers, urlString, urlObject) {
    if (!matchers) {
      return false;
    }

    matchers = Array.isArray(matchers) ? matchers : [matchers]

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

  function processResult(result, urlString, urlObject) {
    if (typeof result === "function") {
      result = result(urlString, urlObject)
    }

    if (typeof result !== "string" && typeof result !== "object") {
      throw new Error(
        `Handler result invalid, expected one of [string, object], found ${typeof result}`
      );
    }

    // If all we got was a string, try to figure out if it's a bundle identifier or an application name
    if (typeof result === "string") {
      const type = isBundleIdentifier(result) ? "bundleId" : "appName";

      return {
        value: result,
        type
      };
    }

    if (typeof result === "object") {
      if (typeof result.type !== "string" || typeof result.value !== "string") {
        throw new Error("Missing valid type and/or value from result")
      }

      return result;
    }

    throw new Error("Unrecognized result value " + JSON.stringify(result));
  }

  return function processUrl(urlString, urlObject) {
    for (handler of module.exports.handlers) {
      const match = resolveMatchers(handler.match, urlString, urlObject);
      if (match) {
        return processResult(handler.value, urlString, urlObject);
      }
    }

    if (module.exports.defaultBrowser) {
      return processResult(module.exports.defaultBrowser, urlString, urlObject);
    }
  };
})();
