(function() {
  const { validate, getErrors } = fastidious;

  const schema = {
    defaultBrowser: validate.string.isRequired,
    options: validate.shape({
      hideIcon: validate.boolean,
      urlShorteners: validate.arrayOf(validate.string)
    }),
    rewrite: validate.arrayOf(
      validate.shape({
        match: validate.oneOf([
          validate.string,
          validate.function,
          validate.regex,
          validate.arrayOf(
            validate.oneOf([validate.string, validate.function, validate.regex])
          )
        ]).isRequired,
        url: validate.oneOf([validate.string, validate.function]).isRequired
      }).isRequired
    ),
    handlers: validate.arrayOf(
      validate.shape({
        match: validate.oneOf([
          validate.string,
          validate.function,
          validate.regex,
          validate.arrayOf(
            validate.oneOf([validate.string, validate.function, validate.regex])
          )
        ]).isRequired,
        browser: validate.oneOf([
          validate.string,
          validate.shape({
            name: validate.string.isRequired,
            type: validate.oneOf(["name", "bundleIdentifier"]).isRequired
          })
        ]).isRequired
      })
    )
  };

  return function validateConfig() {
    if (!module) {
      throw new Error("module is not defined");
    }

    const invalid = getErrors(module.exports, schema, "module.exports.");

    if (invalid.length === 0) {
      return true;
    }

    throw new Error(invalid.join("\n"));
  };
})();
