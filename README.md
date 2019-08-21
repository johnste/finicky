<div align="center">
  <h1>Finicky<img
    height="36"
    width="36"
    alt="finicky logo - hand pointing downwards"
    src="https://raw.githubusercontent.com/johnste/finicky/master/Finicky/Finicky/statusitem%402x.png"
  />
  </h1>

  <p>Always open the right browser</p>

</div>

Finicky is an macOS application that allows you to set up rules that decide which browser is opened for every link or url. With Finicky as your default browser, you can tell it to open Facebook or Reddit in one browser, and Trello or LinkedIn in another.

- Write rules to open urls in any browser
- Rewrite and replace parts of urls before opening them

[![GitHub downloads](https://img.shields.io/github/downloads/johnste/finicky/total.svg?style=flat-square)](https://GitHub.com/johnste/finicky/releases/)
[![GitHub release](https://img.shields.io/github/release-pre/johnste/finicky.svg?style=flat-square)](https://GitHub.com/johnste/finicky/releases/)

<div align="center">
<img src="https://raw.githubusercontent.com/johnste/finicky/gh-pages/Screenshot.png" alt="Finicky screenshot" width="75%"/>
</div>

## Table of Contents

<!-- To regenerate toc run `npx doctoc README.md --github` -->
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Installation](#installation)
- [Example configuration](#example-configuration)
  - [Basic configuration](#basic-configuration)
  - [Wildcard matching](#wildcard-matching)
  - [Opening preferred browser](#opening-preferred-browser)
  - [Rewrite urls](#rewrite-urls)
  - [Advanced usage, settings](#advanced-usage-settings)
  - [Keyboard modifiers](#keyboard-modifiers)
  - [Configuration ideas](#configuration-ideas)
- [API Reference](#api-reference)
- [Issues](#issues)
  - [Bugs](#bugs)
  - [Feature Requests](#feature-requests)
- [Questions](#questions)
- [License](#license)
- [Support development](#support-development)
- [Building from source](#building-from-source)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation

1. Install Finicky:

- Download [the latest release](https://github.com/johnste/finicky/releases), unzip and put `Finicky.app` in your application folder.
  <!-- - Alternatively, you can install with [homebrew-cask](https://github.com/caskroom/homebrew-cask): `brew cask install finicky`. -->

2. Create a file called `.finicky.js` with configuration
   ([examples](#example-configuration)) in your home directory.
3. Start Finicky. Please allow it to be set as the default browser.
4. And you're done. All links clicked that would have opened your browser are now first handled by Finicky.

## Example configuration

### Basic configuration

```js
module.exports = {
  defaultBrowser: "Google Chrome",
  handlers: [
    {
      // Open apple.com and example.org urls in Safari
      match: finicky.matchDomains(["apple.com", "example.org"]),
      browser: "Safari"
    },
    {
      // Open any url including the string "workplace" in Firefox
      match: /workplace/,
      browser: "Firefox"
    }
  ]
};
```

### Wildcard matching

```js
module.exports = {
  defaultBrowser: "Safari",
  handlers: [
    {
      // Open google.com and *.google.com urls in Google Chrome
      match: finicky.matchDomains([
        "google.com", // match google.com domain as string (to make regular expression less complicated)
        /.*\.google.com$/ // match all google.com subdomains
      ]),
      browser: "Google Chrome"
    }
  ]
};
```

### Opening preferred browser

```js
module.exports = {
  defaultBrowser: "Safari",
  handlers: [
    {
      match: finicky.matchDomains(["example.com"]),
      // Opens the first running browsers in the list. If none are running, the first one will be started.
      browser: ["Google Chrome", "Safari", "Firefox"]
    }
  ]
};
```

### Rewrite urls

```js
module.exports = {
  defaultBrowser: "Google Chrome",
  rewrite: [
    {
      // Redirect all urls to use https
      match: ({ url }) => url.protocol === "http",
      url: ({ url }) => ({
        ...url,
        protocol: "https"
      })
    },
    {
      // Avoid being rickrolled
      match: [
        "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
        "https://www.youtube.com/watch?v=oHg5SJYRHA0"
      ],
      url: "about:blank"
    }
  ]
};
```

### Advanced usage, settings

```js
module.exports = {
  defaultBrowser: "Google Chrome",
  options: {
    // Hide the finicky icon from the top bar
    hideIcon: true
  },
  handlers: [
    {
      // Open any link clicked in Slack in Safari
      match: ({ sourceBundleIdentifier }) =>
        sourceBundleIdentifier === "com.tinyspeck.slackmacgap",
      browser: "Safari"
    },
    {
      match: ["http://zombo.com"],
      browser: {
        name: "Google Chrome Canary",
        // Force opening the link in the background
        openInBackground: true
      }
    },
    {
      match: ["http://example.com"],
      // Don't open any browser for this url, effectively blocking it
      browser: null
    }
  ]
};
```

### Keyboard modifiers

```js
module.exports = {
  defaultBrowser: "Google Chrome",
  handlers: [
    {
      // Open links in Safari when the option key is pressed
      // Valid keys are: shift, option, command, control, capsLock, and function.
      // Please not that control usually opens a tooltip menu instead of visiting a link
      match: ({ keys }) => keys.option,
      browser: "Safari"
    }
  ]
};
```

### Configuration ideas

See the wiki page for other [configuration ideas](https://github.com/johnste/finicky/wiki/Configuration-ideas)

## API Reference

- [Configuration Reference](https://johnste.github.io/finicky-docs/interfaces/_finickyconfig_.finickyconfig.html)
- [API Reference](https://johnste.github.io/finicky-docs/interfaces/_finickyapi_.finicky.html)

## Issues

### Bugs

Please file an issue for bugs, missing documentation, or unexpected behavior.

[**See Bugs**](https://github.com/johnste/finicky/issues?q=is%3aopen+is%3aissue+label%3abug)

### Feature Requests

Please file an issue to suggest new features. Vote on feature requests by adding
a 👍.

[**See Feature Requests**](https://github.com/johnste/finicky/labels/feature%20request)

## Questions

Have any other questions or need help? Please feel free to reach out to me on [Twitter](https://twitter.com/johnste_).

## License

[MIT](https://raw.githubusercontent.com/johnste/finicky/master/LICENSE)

## Support development

If you want to help support further development of finicky, feel free to buy me a coffee on ko-fi.

<a href='https://ko-fi.com/E1E5W973' target='_blank'><img height='36' style='border:0px;height:36px;' src='https://az743702.vo.msecnd.net/cdn/kofi1.png?v=2' border='0' alt='Buy Me a Coffee at ko-fi.com' /></a>

## Building from source

Install XCode and XCode command line tools and then run commands:

```shell
    git clone https://github.com/johnste/finicky.git
    cd finicky/Finicky
    xcodebuild
```

When complete you'll find a freshly built **Finicky** app in
`build/release`.
