(function() {
  return function processUrl(urlString, urlObject) {
    for (handler of module.exports.handlers) {
      const match = getMatch(handler.match, urlString, urlObject);
      if (match) {
        return processResult(handler.value, urlString, urlObject);
      }
    }

    if (module.exports.defaultBrowser) {
      return processResult(module.exports.defaultBrowser, urlString, urlObject);
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

  function processResult(result, urlString, urlObject) {
    if (typeof result === "function") {
      result = result(urlString, urlObject);
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
      if (typeof result.value !== "string") {
        throw new Error(`Missing or invalid property value in result: ${typeof result.value}; ${result}`);
      }

      let type;
      if (typeof result.type) {
        type = result.type;
      } else {
        type = isBundleIdentifier(result) ? "bundleId" : "appName";
      }



      const invalidKeys = Object.keys(result).filter(key => !["value", "url", "type"].includes(key));
      if (invalidKeys.length > 0) {
        throw new Error(`Found unrecognized keys in result: ${invalidKeys.join(',')}`);
      }

      return result;
    }

    throw new Error("Unrecognized result value " + JSON.stringify(result));
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
})();
