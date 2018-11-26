(function() {
  function assert(boolean, message) {
    if (!boolean) {
      throw new Error(message);
    }
  }

  function validMatcher(matcher) {
    if (!matcher) {
      return false;
    }

    if (!["function", "string", "object"].includes(typeof matcher)) {
      return false;
    }

    if (typeof matcher === "object" && !(matcher instanceof RegExp)) {
      return false;
    }

    return true;
  }

  function validHandler(handler) {
    if (typeof handler !== "object") {
      return false;
    }

    if (!handler.value || !handler.match) {
      return false;
    }

    return true;
  }

  function resolveMatch(urlString, urlObject, matchers) {
    if (!matchers) {
      return false;
    }

    if (!Array.isArray(matchers)) {
      matchers = [matchers];
    }

    return matchers.some(matcher => {
      assert(validMatcher(matcher), "matcher is valid");

      if (matcher instanceof RegExp) {
        return matcher.test(urlString);
      }

      if (typeof matcher === "string") {
        return matcher === urlString;
      }

      if (typeof matcher === "function") {
        return matcher(urlString, urlObject) || false;
      }

      return false;
    });
  }

  return function processUrl(urlString, urlObject) {
    try {
      if (
        !module ||
        !module.exports ||
        !Array.isArray(module.exports.handlers)
      ) {
        return false;
      }

      const matches = module.exports.handlers.filter(handler => {
        try {
          assert(validHandler(handler), "handler is valid");

          const isMatch = resolveMatch(urlString, urlObject, handler.match);

          return isMatch && handler.value;
        } catch (ex) {
          return false;
        }
      });

      if (matches.length === 0) {
        return undefined;
      }

      return matches[0].value;
    } catch (ex) {
      return undefined;
    }
  }
})();
