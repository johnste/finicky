module.exports = {
  defaultBrowser: "Google Chrome",
  rewrite: [
    {
      match: /example\.com/,
      url: ({ url }) => ({
        ...url,
        host: "example.org"
      })
    }
  ],
  handlers: [
    {
      // Open workplace related sites in work browser
      match: [/workplace/],
      browser: [{ name: "Google Chrome Canary", openInBackground: true }, "Safari"]
    },
    {
      // Open workplace related sites in work browser
      match: [/test/],
      browser: [null, "Safari"]
    },
  ]
};
