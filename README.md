# Finicky

*Always open the right browser*

Finicky allows you to set up rules that decide which browser is opened for every link that would open the default browser. Open Facebook or Reddit in one browser, and Trello or LinkedIn in another. Or even open Spotify links in the Spotify client.

#### Install

Install XCode and XCode command line tools, then from a terminal:

    git clone https://github.com/johnste/finicky.git
    cd finicky/Finicky
    xcodebuild

When complete you'll find a freshly built **Finicky** app in
`build/release`.

To install it just drag-drop it to your `/Applications` folder.

When you first run Finicky, you'll need to allow it to be set as the default browser.

Create a file called `.finicky.js` in your home directory.

##### Example configuration

```javascript

finicky.defaultBrowser('com.google.Chrome')

// Open work stuff in Canary
finicky.onUrl(function(url) {
	if (url.match(/^https?:\/\/\bitbucket\.org|trello\.com|([a-z]+)?.google\.com)/)) {
		return {
			bundleIdentifier: 'com.google.Chrome.canary'
		}
	}
});

// Open social network links in Google Chrome
finicky.onUrl(function(url) {
	if (url.match(/^https?:\/\/(youtube|facebook|twitter|linkedin)\.com/)) {
		return {
			bundleIdentifier: 'com.google.Chrome'
		}
	}
});

// Open Spotify links in client
finicky.onUrl(function(url) {
	if (url.match(/^https?:\/\/open\.spotify\.com/)) {
		return {
			bundleIdentifier: 'com.spotify.client'
		}
	}
});

// Rewrite Twitter status links to open in Twitter client
finicky.onUrl(function(url) {
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
finicky.onUrl(function(url) {
    var url = url.replace(/^https?:\/\/www\.bing\.com\/search/, 'https://duckduckgo.com')
    return {
    	url: url
    }
});
```

#### Documentation

[Javascript API Documentation](https://github.com/johnste/finicky/wiki/Javascript-API-Documentation)
