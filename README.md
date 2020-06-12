<div align="center">
  <h1><img
    height="100"
    width="336"
    alt="finicky logo"
    src="https://raw.githubusercontent.com/johnste/finicky/gh-pages/finicky-logo.svg"
  />
  </h1>

  <strong>Always open the right browser</strong>

</div>

Finicky is a macOS application that allows you to set up rules that decide which browser is opened for every link or url. With Finicky as your default browser, you can tell it to open Facebook or Reddit in one browser, and Trello or LinkedIn in another.

- Decide what urls to open in what browser or app
- Edit urls before opening them
- Complete control over configuration using JavaScript

[![GitHub start](https://badgen.net/github/stars/johnste/finicky?color=pink&icon=github)](https://GitHub.com/johnste/finicky/)
[![GitHub release](https://badgen.net/github/release/johnste/finicky/stable?color=purple)](https://GitHub.com/johnste/finicky/releases/)




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
  - [Rewrite urls](#rewrite-urls)
- [Documentation](#documentation)
- [Configuration ideas](#configuration-ideas)
- [Alternatives](#alternatives)
- [Support development](#support-development)
- [Issues](#issues)
  - [Bugs](#bugs)
  - [Feature Requests](#feature-requests)
  - [Questions](#questions)
- [License](#license)


<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation

1. Installation alternatives:

- Download [the latest release](https://github.com/johnste/finicky/releases), unzip and put `Finicky.app` in your application folder.
- Install with [homebrew-cask](https://github.com/caskroom/homebrew-cask): `brew cask install finicky`.

2. Create a file called `.finicky.js` with configuration
   ([examples](#example-configuration)) in your home directory OR generate a basic configuration with [Finicky Kickstart](https://finicky-kickstart.now.sh/)
   
3. Start Finicky. Please allow it to be set as the default browser.
4. And you're done. All links clicked that would have opened your browser are now first handled by Finicky.

## Example configuration

```js
// ~/.finicky.js

module.exports = {
  defaultBrowser: "Google Chrome",
  rewrite: [
    {
      // Redirect all urls to use https
      match: ({ url }) => url.protocol === "http",
      url: { protocol: "https" }
    }
  ],
  handlers: [
    {
      // Open apple.com and example.org urls in Safari
      match: ["apple.com*", "example.org*"],
      browser: "Safari"
    },
    {
      // Open any url that includes the string "workplace" in Firefox
      match: /workplace/,
      browser: "Firefox"
    },
    {
      // Open google.com and *.google.com urls in Google Chrome
      match: [
        "google.com*", // match google.com domain as string (to make regular expression less complicated)
        "*.google.com*" // match all google.com subdomains
      ],
      browser: "Google Chrome"
    }
  ]
};
```

See the [documentation](#documentation) for all the features Finicky supports.

## Documentation

Finicky has extensive support for matching, rewriting and starting browsers or other application that handle urls. See the wiki for the [full configuration documentation](https://github.com/johnste/finicky/wiki/Configuration) explaining all available, APIs and options as well as detail information on how to match on urls.

## Configuration ideas

See the wiki page for other [configuration ideas](https://github.com/johnste/finicky/wiki/Configuration-ideas) by users of Finicky.

## Alternatives

If you are looking for something that lets you pick the browser to activate in a graphical interface, check out [Browserosaurus](https://browserosaurus.com/) by Will Stone, an open source browser prompter for macOS. It works really well together with Finicky!

## Support development

If you want to help support further development of finicky, feel free to buy me a coffee on ko-fi.

<a href='https://ko-fi.com/E1E5W973' target='_blank'><img height='36' style='border:0px;height:36px;' src='https://az743702.vo.msecnd.net/cdn/kofi1.png?v=2' border='0' alt='Buy Me a Coffee at ko-fi.com' /></a>

## Issues

### Bugs

Please file an issue for bugs, missing documentation, or unexpected behavior.

[**See Bugs**](https://github.com/johnste/finicky/issues?q=is%3aopen+is%3aissue+label%3abug)

### Feature Requests

Please file an issue to suggest new features. Vote on feature requests by adding
a 👍.

[**See Feature Requests**](https://github.com/johnste/finicky/labels/feature%20request)

### Questions

Have any other questions or need help? Please feel free to reach out to me on [Twitter](https://twitter.com/johnste_).

## License

[MIT](https://raw.githubusercontent.com/johnste/finicky/master/LICENSE)
