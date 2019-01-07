(function() {
  function validateValue(value, handlerNum) {
    if (
      typeof value !== "string" &&
      typeof value !== "object" &&
      typeof value !== "function"
    ) {
      throw new Error(
        `Handler #${index} value: expected one of [string, object], found "${typeof value}"`
      );
    }

    if (typeof value === "object") {
      const hasBundleId = typeof value.bundleId === "string";
      const hasName = typeof value.name === "string";

      if (!hasBundleId && !hasName) {
        throw new Error(
          `Handler #${index} value: expected object to have string property "bundleId" OR "name" `
        );
      }

      if (hasBundleId && hasName) {
        throw new Error(
          `Handler #${index} value: expected object to have only one of the string properties "bundleId" OR "name"`
        );
      }
    }
  }

  function validateMatch(matchers, handlerNum) {
    const isArray = Array.isArray(matchers);
    matchers = isArray ? matchers : [matchers];

    matchers.forEach((matcher, index) => {
      const preamble =
        (isArray
          ? `Handler #${handlerNum} matcher #${index}`
          : `Handler #${handlerNum} matcher`) +
        ": expected one of [function, string, Regular Expression]";

      if (!["function", "string", "object"].includes(typeof matcher)) {
        throw new Error(`${preamble}, found "${typeof matcher}"`);
      }

      if (typeof matcher === "object" && !(matcher instanceof RegExp)) {
        throw new Error(
          `${preamble}, found object that was not a Regular Expression`
        );
      }
    });
  }

  function validateHandler(handler, index) {
    if (typeof handler !== "object" || !handler.match || !handler.value) {
      throw new Error(
        `Handler #${index}: Must be an object with a properties: [match, value]`
      );
    }

    validateMatch(handler.match, index);
    validateValue(handler.value, index);
  }

  return function validateConfig() {
    if (!module || !module.exports) {
      throw new Error("module.exports is not defined");
    }

    if (!Array.isArray(module.exports.handlers)) {
      throw new Error("module.exports.handlers is not an array");
    }

    module.exports.handlers.forEach(validateHandler);

    return true;
  };
})();
