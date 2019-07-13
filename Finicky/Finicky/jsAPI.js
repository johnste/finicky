/**
 * Extends finicky js api with some utility functions.
 */

finicky.matchDomains = function(matchers, ...args) {
  if (args.length > 0) {
    throw new Error(
      "finicky.matchDomains(domains) only accepts one argument. See https://johnste.github.io/finicky-docs/interfaces/_finickyapi_.finicky.html#matchdomains for more information"
    );
  }

  if (!Array.isArray(matchers)) {
    matchers = [matchers];
  }

  matchers.forEach(matcher => {
    if (matcher instanceof RegExp || typeof matcher === "string") {
      return;
    }
    throw new Error(
      `finicky.matchDomains(domains): Unrecognized domain "${matcher}"`
    );
  });

  return function({ url }) {
    const domain = url.host;
    return matchers.some(matcher => {
      if (matcher instanceof RegExp) {
        return matcher.test(domain);
      } else if (typeof matcher === "string") {
        return matcher === domain;
      }

      return false;
    });
  };
};

// Warn when using deprecated API methods
finicky.onUrl = function() {
  finicky.log(
    "finicky.onUrl is no longer supported in this version of Finicky, please go to https://github.com/johnste/finicky for updated documentation"
  );
  finicky.notify(
    "finicky.onUrl is no longer supported",
    "Check the Finicky website for updated documentation"
  );
};

finicky.setDefaultBrowser = function() {
  finicky.log(
    "finicky.setDefaultBrowser is no longer supported in this version of Finicky, please go to https://github.com/johnste/finicky for updated documentation"
  );
  finicky.notify(
    "finicky.setDefaultBrowser is no longer supported",
    "Check the Finicky website for updated documentation"
  );
};
