<div align="center"><picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/ab66e6cc-25d1-4f5f-9c98-c742ecb2261f">
  <img alt="Finicky Logo" height="110" src="https://github.com/user-attachments/assets/067d7619-a2be-49dd-8a4f-5e9a46fa632a">
</picture>
<br/><br/>
 <strong>Always open the right browser</strong><br>
    <br/>

</div>

Finicky is a macOS application that allows you to set up rules that decide which browser is opened for every url. With Finicky as your default browser, you can tell it to open Bluesky or Reddit in one browser, and LinkedIn or Google Meet in another.

- Route any URL to your preferred browser with powerful matching rules
- Automatically edit URLs before opening them (e.g., force HTTPS, remove tracking parameters)
- Write rules in JavaScript or TypeScript for complete control
- Create complex routing logic with regular expressions and custom functions
- Handle multiple browsers and apps with a single configuration
- Keep your workflow organized by separating work and personal browsing

[![GitHub prerelease](https://badgen.net/github/release/johnste/finicky?color=purple)](https://GitHub.com/johnste/finicky/releases/) ![MIT License](https://badgen.net/github/license/johnste/finicky) ![Finicky v4 release](https://badgen.net/github/milestones/johnste/finicky/6?color=pink)

## Table of Contents

- [Install](#install)
- [Basic configuration](#basic-configuration)
- [Configuration](#documentation)

## Install

- Download from [releases](https://github.com/johnste/finicky/releases)
- Or install via homebrew: `brew install --cask finicky`

## Basic configuration

Here's a short example configuration that can help you get started

```js
// ~/.finicky.js
export default {
  defaultBrowser: "Google Chrome",
  rewrite: [
    {
      // Redirect all x.com urls to use xcancel.com
      match: "x.com/*",
      url: (url) => {
        url.host = "xcancel.com";
        return url;
      },
    },
  ],
  handlers: [
    {
      // Open all bsky.app urls in Firefox
      match: "bsky.app/*",
      browser: "Firefox",
    },
    {
      // Open google.com and *.google.com urls in Google Chrome
      match: [
        "google.com/*", // match google.com urls
        "*.google.com*", // also match google.com subdomains
      ],
      browser: "Google Chrome",
    },
  ],
};
```

See the [configuration](#configuration) for all the features Finicky supports.

## Configuration

Finicky has extensive support for matching, rewriting and starting browsers or other application that handle urls. See the wiki for the [full configuration documentation](<https://github.com/johnste/finicky/wiki/Configuration-(v4)>) explaining available, APIs and options as well as detail information on how to match on urls.

- The wiki has some good [configuration ideas](https://github.com/johnste/finicky/wiki/Configuration-ideas).
- Visit [discussions](https://github.com/johnste/finicky/discussions) to discuss supporting specific apps.

# Other

### Browser extensions

Finicky has browser extensions for Chrome and Firefox. They add an "open with Finicky" on links, and alt-clicking links opens them in Finicky directly.

- [Firefox Add-ons](https://addons.mozilla.org/en-US/firefox/addon/finicky/)
- [Chrome Web Store](https://chromewebstore.google.com/detail/finicky/kcnjhpdfmjcbohngnmobipdllkhnpdbk)

### Building Finicky from source

See [Building Finicky from source](https://github.com/johnste/finicky/wiki/Building-Finicky-from-source)

### Works well with

If you are looking for something that lets you pick the browser to activate in a graphical interface, check out [Browserosaurus](https://browserosaurus.com/) by Will Stone, an open source browser prompter for macOS. It works really well together with Finicky!

### Questions

Have any other questions or need help? Please feel free to reach out to me on [Bluesky](https://bsky.app/profile/mejkarsense.se) or post an issue here

## License

[MIT](https://raw.githubusercontent.com/johnste/finicky/master/LICENSE)

Icon designed by [@uetchy](https://github.com/uetchy)
