// @ts-check

/**
 * @typedef {import('/Applications/Finicky.app/Contents/finicky.d.ts')} Globals
 * @typedef {import('/Applications/Finicky.app/Contents/finicky.d.ts').FinickyConfig} FinickyConfig
 */

const { matchHostnames } = finicky;

/**
 * @type {FinickyConfig}
 */
export default {
  defaultBrowser: "Firefox",
  options: {
    urlShorteners: [],
    logRequests: true,
  },
  rewrite: [
    {
      match: "*query=value*",
      url: (url) => url,
    },
    {
      match: () => {
        console.log(finicky.getModifierKeys());
        console.log(finicky.getSystemInfo());
        return false;
      },
      url: (url) => url,
    },
    // {
    //   match: "https://slack-redir.net/link?url=*",
    //   url: (url ) => url,
    // },
    // {
    //   match: (url) => {
    //     console.log('url was', typeof url,url instanceof URL, url);
    //     return true;
    //   },
    //   url: ( url ) => url,
    // },
    // {
    //   match: /test/,
    //   url: ( url ) => url,
    // },
  ],

  handlers: [
    {
      // Open workplace related sites in work browser
      match: "?query=value",
      open: "Safari", // "Arc" // "Brave Browser", //, //"Firefox Developer Edition"
    },
  ],
};
