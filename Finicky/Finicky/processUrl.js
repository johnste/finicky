(function() {
  const { validate, getErrors } = fastidious;

  const appDescriptorSchema = validate.oneOf([
    validate.shape({
      name: validate.string,
      appType: validate.oneOf([
        validate.value("bundleId"),
        validate.value("appName"),
        validate.value("none")
      ]).isRequired,
      openInBackground: validate.boolean
    })
  ]).isRequired;

  const urlSchema = {
    url: validate.oneOf([
      validate.string,
      validate.shape({
        protocol: validate.oneOf(["http", "https"]).isRequired,
        username: validate.string,
        password: validate.string,
        host: validate.string.isRequired,
        port: validate.oneOf([validate.number, validate.value(null)]),
        pathname: validate.string,
        search: validate.string,
        hash: validate.string
      })
    ]).isRequired
  };

  return function processUrl(options = {}) {
    options = rewriteUrl(options);

    if (Array.isArray(module.exports.handlers)) {
      for (let handler of module.exports.handlers) {
        if (isMatch(handler.match, options)) {
          return processBrowserResult(handler.browser, options);
        }
      }
    }

    return processBrowserResult(module.exports.defaultBrowser, options);
  };

  function validateSchema(value, schema, path = "") {
    const errors = getErrors(value, schema, path);
    if (errors.length > 0) {
      throw new Error(
        errors.join("\n") + "\nReceived value: " + JSON.stringify(value, null, 2)
      );
    }
  }

  function createUrl(url) {
    const { protocol, host, pathname = "" } = url;
    let port = url.port ? `:${url.port}` : "";
    let search = url.search ? `?${url.search}` : "";
    let hash = url.hash ? `#${url.hash}` : "";
    let auth = url.username ? `${url.username}` : "";
    auth += url.password ? `:${url.password}` : "";

    return `${protocol}://${auth}${host}${port}${pathname}${search}${hash}`;
  }

  function rewriteUrl(options) {
    if (Array.isArray(module.exports.rewrite)) {
      for (let rewrite of module.exports.rewrite) {
        if (isMatch(rewrite.match, options)) {
          let urlResult = resolveFn(rewrite.url, options);

          validateSchema({ url: urlResult }, urlSchema);

          if (typeof urlResult === "string") {
            options = {
              ...options,
              url: finicky.getUrlParts(urlResult),
              urlString: urlResult
            };
          } else {
            options = {
              ...options,
              url: urlResult,
              urlString: createUrl(urlResult)
            };
          }
        }
      }
    }

    return options;
  }

  function isMatch(matcher, options) {
    if (!matcher) {
      return false;
    }

    const matchers = Array.isArray(matcher) ? matcher : [matcher];

    return matchers.some(matcher => {
      if (matcher instanceof RegExp) {
        return matcher.test(options.urlString);
      } else if (typeof matcher === "string") {
        return matcher === options.urlString;
      } else if (typeof matcher === "function") {
        return !!matcher(options);
      }

      return false;
    });
  }

  // Recursively resolve handler to value
  function resolveFn(handler, ...args) {
    if (typeof handler === "function") {
      return resolveFn(handler(...args), ...args);
    }
    return handler;
  }

  function getAppType(value) {
    if (value === null) {
      return "none";
    }

    return isBundleIdentifier(value) ? "bundleId" : "appName";
  }

  function processBrowserResult(handler, options) {
    let browser = resolveFn(handler, options);

    // If all we got was a string, try to figure out if it's a bundle identifier or an application name
    if (typeof browser === "string" || browser === null) {
      browser = {
        name: browser
      };
    }

    if (typeof browser === "object" && !browser.appType) {
      const name = browser.name === null ? "" : browser.name;

      browser = {
        ...browser,
        name,
        appType: getAppType(browser.name)
      };
    }

    validateSchema(browser, appDescriptorSchema);

    return { ...browser, url: options.urlString };
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
