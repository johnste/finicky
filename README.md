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

Finicky is an Mac OS application that allows you to set up rules that decide which browser is opened for every link. Open Facebook or Reddit in one browser, and Trello or LinkedIn in another.

- Write rules to open urls in any browser
- Rewrite and replace parts of urls before opening them

## Table of Contents

<!-- To regenerate toc run `npx doctoc README.md --github` -->
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Installation](#installation)
- [Example configuration](#example-configuration)
  - [Basic configuration](#basic-configuration)
  - [Rewrite urls](#rewrite-urls)
  - [Advanced usage, settings](#advanced-usage-settings)
- [API Reference](#api-reference)
- [Issues](#issues)
  - [Bugs](#bugs)
  - [Feature Requests](#feature-requests)
- [Questions](#questions)
- [License](#license)
- [Building from source](#building-from-source)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation

1. Install Finicky:

- Download [the latest release](https://github.com/johnste/finicky/releases), unzip and put `Finicky.app` in your application folder.
- Alternatively, you can install with [homebrew-cask](https://github.com/caskroom/homebrew-cask): `brew cask install finicky`.

2. Create a file called `.finicky.js` with configuration
   ([examples](#example-configuration)) in your home directory.
3. Start Finicky. Please allow it to be set as the default browser.
4. And you're done. All links clicked that would have opened your browser are now first handled by Finicky.

## Example configuration

### Basic configuration

```js
module.exports = {
  defaultBrowser: "Google Chrome",
  handlers: [{
    // Open apple.com and example.org urls in Safari
    match: finicky.matchDomains(["apple.com", "example.org"]),
    browser: "Safari"
  }, {
    // Open any url including the string "workplace" in Firefox
    match: /workplace/,
    browser: "Firefox"
  }];
}
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
        sourceBundleIdentifier === "com.tinyspeck.chatlyio",
      browser: "Safari"
    },
    {
      match: ["http://zombo.com"],
      browser: {
        name: "Google Chrome Canary",
        // Force opening the link in the background
        openInBackground: true
      }
    }
  ]
};
```

## API Reference

TBD

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

## Building from source

Install XCode and XCode command line tools and then run commands:

```shell
    git clone https://github.com/johnste/finicky.git
    cd finicky/Finicky
    xcodebuild
```

When complete you'll find a freshly built **Finicky** app in
`build/release`.
