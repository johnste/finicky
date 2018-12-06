(function() {
  function resolveMatch(urlString, urlObject, match) {
    if (!match) {
      return false;
    }

    if (!Array.isArray(match)) {
      match = [match];
    }

    return match.some(matcher => {
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
  function isBundleIdentifier(value) {
    // Regular expression to match Uniform Type Identifiers
    // Adapted from https://stackoverflow.com/a/34241710/1698327
    const bundleIdRegex = /^[A-Za-z]{2,6}((?!-)\.[A-Za-z0-9-]{1,63})+$/;
    if (bundleIdRegex.test(value)) {
      return true;
    }
    return false;
  }

  function processResult(value, urlString, urlObject) {

    if (typeof value === "function") {
      value = processResult(urlString, urlObject)
    }

    if (typeof value !== "string" && typeof value !== "object") {
      throw new Error(
        `Handler result invalid, expected one of [string, object], found ${typeof value}`
      );
    }

    // If all we got was a string, try to figure out if it's a bundle identifier or an application name
    if (typeof value === "string") {
      const type = isBundleIdentifier(value) ? "bundleId" : "name";
      return {
        [type]: value
      };
    }

    if (typeof value === "object") {
      const hasBundleId =
        Reflect.has(value, "bundleId") && typeof value.bundleId === "string";
      const hasName =
        Reflect.has(value, "name") && typeof value.name === "string";

      if (!hasBundleId && !hasName) {
        throw new Error("Missing bundleId and name from handler result")
      }

      if (hasBundleId && hasName) {
        throw new Error("Handler result contains both bundleId and name")
      }
    }
  }

  return function processUrl(urlString, urlObject) {
    finicky.log(JSON.stringify(urlObject));

    for (handler in module.exports.handler) {
      const match = resolveMatch(urlString, urlObject, handler.match);
      if (match) {
        return processResult(matches[0].value, urlString, urlObject);
      }
    }
  };
})();
