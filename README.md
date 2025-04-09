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

- [Getting started](#getting-started)
- [Starter configuration](#starter-configuration)
- [Configuration](#documentation)
- [Building Finicky from source](#building-finicky-from-source)

## Getting started

Download from [releases](https://github.com/johnste/finicky/releases)

Finicky 4+ is in beta, and should work for most cases. The documentation needs to be updated, but most Finicky v3 configurations should work. Supports MacOS 12+.

## Starter configuration

Here's a short example configuration that can help you get started

```js
// ~/.finicky.js
export default {
  defaultBrowser: "Google Chrome",
  {
    checkForUpdate: true
  },
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
      // Open apple.com and example.com urls in Safari
      match: ["apple.com*", "example.com*"],
      browser: "Safari",
    },
    {
      // Open any url that includes the string "workplace" in Firefox
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

## Building Finicky from source

If you'd like to build Finicky from source:

1. Install Go 1.23.4
2. Install Node 22
3. Run `./scripts/install.sh` from base folder to install dependencies
4. Run `./scripts/build.sh` from base folder to build Finicky

### Works well with

If you are looking for something that lets you pick the browser to activate in a graphical interface, check out [Browserosaurus](https://browserosaurus.com/) by Will Stone, an open source browser prompter for macOS. It works really well together with Finicky!

### Questions

Have any other questions or need help? Please feel free to reach out to me on [Bluesky](https://bsky.app/profile/mejkarsense.se) or post an issue here

## License

[MIT](https://raw.githubusercontent.com/johnste/finicky/master/LICENSE)

Icon designed by [@uetchy](https://github.com/uetchy)
