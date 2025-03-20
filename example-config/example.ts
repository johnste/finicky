import type {
  BrowserHandler,
  FinickyConfig,
} from "/Applications/Finicky.app/Contents/Resources/finicky.d.ts";

const handler: BrowserHandler = {
  match: (url: URL, { opener }) => {
    return url.host.includes("workplace");
  },
  browser: "Firefox",
};

const config: FinickyConfig = {
  defaultBrowser: "Google Chrome",
  handlers: [
    handler,
    {
      match: (url: URL, { opener }) => {
        console.log("opener", opener);
        return opener?.name.includes("Slack") || false;
      },
      browser: "Firefox",
    },
  ],
};

export default config;
