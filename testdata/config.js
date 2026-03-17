/** @type {Parameters<typeof import("finicky").default>[0]} */
module.exports = {
  defaultBrowser: "Safari",
  handlers: [
    // This rule lives in the JS config and takes priority over rules.json
    {
      match: "github.com/*",
      browser: "Google Chrome",
    },
    {
      match: /linear\.app/,
      browser: "Firefox",
    },
  ],
  options: {
    keepRunning: true,
    logRequests: true,
    checkForUpdates: true,
  },
};
