# Finicky ![Finicky logo](https://raw.githubusercontent.com/johnste/finicky/master/Finicky/Finicky/statusitem%402x.png)

*Always open the right browser*

Finicky is an OS X application that allows you to set up rules that decide which browser is opened for every link that would open the default browser. Open Facebook or Reddit in one browser, and Trello or LinkedIn in another. Or Spotify links in the Spotify client. Or whatever url in whatever app.

Features include:
- Url rewriting
- Opening links in background
- Resolving short urls before running them through handlers

#### Install

1. Download [the latest release](https://github.com/johnste/finicky/releases), unzip and drop Finicky.app in your application folder. Or, if you have [homebrew-cask](https://github.com/caskroom/homebrew-cask) set up, install with `brew cask install finicky`.
2. Create a file called `.finicky.js` in your home directory and set up a default browser and some url handlers.
3. Start finicky. Please allow it to be set as the default browser.
4. And you're done. All http and https links clicked that would have opened a link in your default browser are now first handled by Finicky.

#### Building from source
Install XCode and XCode command line tools, then from a terminal:

    git clone https://github.com/johnste/finicky.git
    cd finicky/Finicky
    xcodebuild

When complete you'll find a freshly built **Finicky** app in
`build/release`.

#### Documentation
[Javascript API Documentation](https://github.com/johnste/finicky/wiki/Javascript-API-Documentation)

#### Example configuration
```javascript

finicky.setDefaultBrowser('com.google.Chrome')

// Open social network links in Google Chrome
finicky.onUrl(function(url, opts) {
	if (url.match(/^https?:\/\/(youtube|facebook|twitter|linkedin)\.com/)) {
		return {
			bundleIdentifier: 'com.google.Chrome'
		}
	}
});

// Open Spotify links in client
finicky.onUrl(function(url, opts) {
	if (url.match(/^https?:\/\/open\.spotify\.com/)) {
		return {
			bundleIdentifier: 'com.spotify.client'
		}
	}
});

// Open Twitter links in client
finicky.onUrl(function(url, opts) {
	var matches = url.match(/^https?:\/\/twitter\.com\/.+\/status\/([0-9]+)/)
	if (matches && matches[1]) {
		var statusId = matches[1];
		return {
			url: 'twitter://status?id=' + statusId,
			bundleIdentifier: 'com.twitter.twitter-mac'
		}
	}
});

// Rewrite all Bing links to DuckDuckGo instead
finicky.onUrl(function(url, opts) {
    var url = url.replace(/^https?:\/\/www\.bing\.com\/search/, 'https://duckduckgo.com')
    return {
    	url: url
    }
});

// Always open links from Mail in Safari
finicky.onUrl(function(url, opts) {
	var sourceApplication = opts && opts.sourceBundleIdentifier
	if (sourceApplication === 'com.apple.mail') {
		return {
			bundleIdentifier: 'com.apple.safari'
		}
	}
});

```
