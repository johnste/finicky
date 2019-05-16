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

Finicky is an Mac OS application that allows you to set up rules that decide which browser is opened for every link. Open Facebook or Reddit in one browser, and Trello or LinkedIn in another. Or Spotify links in the Spotify client. Or whatever url in whatever app.

<!-- To regenerate toc run `npx doctoc README.md --github` -->
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Install](#install)
- [Example configuration](#example-configuration)
  - [Simple configuration to get started:](#simple-configuration-to-get-started)
  - [Rewrite urls before being handled](#rewrite-urls-before-being-handled)
  - [Optional settings](#optional-settings)
- [Documentation](#documentation)
- [Questions](#questions)
- [License](#license)
- [Building from source](#building-from-source)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Installation

1. Download [the latest release](https://github.com/johnste/finicky/releases), unzip and drop Finicky.app in your application folder. Alternatively, if you have [homebrew-cask](https://github.com/caskroom/homebrew-cask) available, install with `brew cask install finicky`.
2. Create a file called `.finicky.js` with some [configuration](#example-configuration) in your home directory.
3. Start Finicky. Please allow it to be set as the default browser.
4. And you're done. All http and https links clicked that would have opened your default browser are now first handled by Finicky.

# Example configuration

## Simple configuration to get started

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

## Rewrite urls before being handled

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

## Other settings

```js
module.exports = {
  defaultBrowser: "Google Chrome",
  options: {
    // Hide the finicky icon from the top bar
    hideIcon: true
  },
  handlers: [
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

# Documentation

[Javascript API Documentation](https://github.com/johnste/finicky/wiki#javascript-api)

# Questions

Have any other questions or need help? Please feel free to reach out to me on [Twitter](https://twitter.com/johnste_)

# License

[MIT](https://raw.githubusercontent.com/johnste/finicky/master/LICENSE)

# Building from source

Install XCode and XCode command line tools, then from a terminal:

```shell
    git clone https://github.com/johnste/finicky.git
    cd finicky/Finicky
    xcodebuild
```

When complete you'll find a freshly built **Finicky** app in
`build/release`.
