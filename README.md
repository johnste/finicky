<div align="center"><picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/ab66e6cc-25d1-4f5f-9c98-c742ecb2261f">
  <img alt="Finicky Logo" height="110" src="https://github.com/user-attachments/assets/067d7619-a2be-49dd-8a4f-5e9a46fa632a">
</picture>
<br/><br/>
 <strong>Always open the right browser</strong><br>
    <br/>


</div>

Finicky is a macOS application that allows you to set up rules that decide which browser is opened for every link or url. With Finicky as your default browser, you can tell it to open Facebook or Reddit in one browser, and Trello or LinkedIn in another.

- Decide what urls to open in what browser or app
- Edit urls before opening them
- Complete control over configuration using JavaScript



[![GitHub prerelease](https://badgen.net/github/release/johnste/finicky?color=purple)](https://GitHub.com/johnste/finicky/releases/) ![MIT License](https://badgen.net/github/license/johnste/finicky) ![Finicky v4 release](https://badgen.net/github/milestones/johnste/finicky/6?color=pink)

## Table of Contents

- [Getting started](#getting-started)
- [Example configuration](#example-configuration)
- [Documentation](#documentation)
- [Configuration tips](#configuration-tips)
- [Alternatives](#alternatives)
- [Building Finicky from source](#building-finicky-from-source)
- [License](#license)

## Getting started

Download from [releases](https://github.com/johnste/finicky/releases)

Finicky 4+ is in beta, and should work for most cases. The documentation needs to be updated, but most Finicky v3 configurations should work. Supports MacOS 12+.

Finicky 3.4.0 is the latest stable release, but it is a few years old and unsupported.



## Example configuration

```js
// ~/.finicky.js
export default { 
  defaultBrowser: "Google Chrome",
  rewrite: [
    {
      // Redirect all urls to use https
      match: (url) => url.protocol === "http:",
      url: (url) => {
        url.protocol = "https:";
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
      match: /workplace/,
      browser: "Firefox",
    },
    {
      // Open google.com and *.google.com urls in Google Chrome
      match: [
        "google.com*", // match google.com urls
        "*.google.com*", // match google.com subdomains
      ],
      browser: "Google Chrome",
    },
  ],
};
```

See the [documentation](#documentation) for all the features Finicky supports.

## Documentation

Finicky has extensive support for matching, rewriting and starting browsers or other application that handle urls. See the wiki for the [full configuration documentation](https://github.com/johnste/finicky/wiki/Configuration-(v3)) explaining all available, APIs and options as well as detail information on how to match on urls.

⚠️ Please note that Finicky 4 will affect the interface slightly, details to come ⚠️

## Configuration tips

See the wiki page for other [configuration tips](https://github.com/johnste/finicky/wiki/Configuration-ideas) by users of Finicky.

⚠️ Please note that Finicky 4 will affect the interface slightly, details to come ⚠️

## Alternatives

If you are looking for something that lets you pick the browser to activate in a graphical interface, check out [Browserosaurus](https://browserosaurus.com/) by Will Stone, an open source browser prompter for macOS. It works really well together with Finicky!

## Building Finicky from source

If you'd like to build Finicky from source:

1. Install Go 1.23.4
2. Install Node 22
3. Run `./scripts/install.sh` from base folder to install dependencies
4. Run `./scripts/build.sh` from base folder to build Finicky

### Questions

Have any other questions or need help? Please feel free to reach out to me on [Bluesky](https://bsky.app/profile/mejkarsense.se) or post an issue here

## License

[MIT](https://raw.githubusercontent.com/johnste/finicky/master/LICENSE)

Icon designed by [@uetchy](https://github.com/uetchy)
