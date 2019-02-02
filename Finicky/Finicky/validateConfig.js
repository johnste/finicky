(function() {
  return function validateConfig() {
    if (!module || !module.exports) {
      throw new Error("module.exports is not defined");
    }

    if (!Array.isArray(module.exports.handlers)) {
      throw new Error("module.exports.handlers is not an array");
    }

    module.exports.handlers.forEach(validateHandler);
  };

  function validateHandler(handler, index) {
    if (typeof handler !== "object" || !handler.match || !handler.value) {
      throw new Error(
        `Handler #${index}: Expected object with properties: { match, value }`
      );
    }

    validateMatch(handler.match, index);
    validateValue(handler.value, index);
  }

  function validateMatch(matcher, handlerNum) {
    const matchers = Array.isArray(matcher) ? matcher : [matcher];

    matchers.forEach((matcher, index) => {
      const preamble = `Handler #${handlerNum} matcher #${index}: expected one of [function, string, Regular Expression]`;

      if (!isOfType(matcher, ["function", "string", "object"])) {
        throw new Error(`${preamble}, found "${typeof matcher}": ${matcher}`);
      }

      if (typeof matcher === "object" && !(matcher instanceof RegExp)) {
        throw new Error(
          `${preamble}, found object that was not a Regular Expression`
        );
      }
    });
  }

  function validateValue(value, handlerNum) {
    if (!isOfType(value, ["string", "object", "function"])) {
      throw new Error(
        `Handler #${handlerNum} value: expected one of [string, object, function], found "${typeof value}": ${value}`
      );
    }

    if (typeof value === "object") {
      const hasBundleId = typeof value.bundleId === "string";
      const hasName = typeof value.name === "string";

      if ((!hasBundleId && !hasName) || (hasBundleId && hasName)) {
        throw new Error(
          `Handler #${handlerNum} value: expected object to have string property "bundleId" OR "name" `
        );
      }
    }
  }

  function isOfType(value, type = []) {
    const types = Array.isArray(type) ? type : [type];
    return types.includes(typeof value);
  }
})();
