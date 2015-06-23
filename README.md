# Finicky

*Always open the right browser*

Finicky allows you to set up rules that decide which browser is opened for every link that would open the default browser. Open Facebook, Instagram or Reddit in one browser, and Trello, Google Drive or LinkedIn in another.

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

#### Configuration

Finicky identifies browsers by their bundle identifier. These are bundle identifiers for common browsers:

| Browser              | Bundle Identifier        |
|----------------------|--------------------------|
| Google Chrome        | com.google.Chrome        |
| Google Chrome Canary | com.google.Chrome.canary |
| Opera                | com.operasoftware.Opera  |
| Mozilla Firefox      | org.mozilla.firefox      |
| Safari               | com.apple.Safari         |

##### Example configuration

```javascript

prefix = "^https?:\/\/"

var config = {
	'com.google.Chrome.canary': [
		"bitbucket\.org",
		"([a-z]+)?.google\.com",
	],
	'com.google.Chrome': [
		"youtube\.com",
		"facebook\.com",
		"twitter\.com",
	]
};

for(browser in config) {
	var patterns = config[browser]
	config[browser] = patterns.map(function(pattern){
		return prefix + pattern;
	});
}

api.config(config)

api.defaultBrowser('com.google.Chrome')

api.onUrl(function(url) {
	// Replace all bing links to google instead
	url.replace(/^https?:\/\/www\.bing\.com/search', 'https://duckduckgo.com')
});
```

#### Usage

Click any link that would start your default browser and Finicky starts the one which pattern matches.

#### Documentation

[Javascript API Documentation](https://github.com/johnste/finicky/wiki/Javascript-API-Documentation)
