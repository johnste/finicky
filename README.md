# Finicky

*Always open the right browser*

#### Install

Install XCode and XCode command line tools, then from a terminal:

    git clone https://github.com/johnste/finicky.git
    cd finicky/Finicky
    xcodebuild

When complete you'll find a freshly built **Finicky** app in
`build/release`.

To install it just drag-drop it to your `/Applications` folder.

When you first run Finicky, you'll need to allow it to be set as the default browser.

Create a filed called `.finicky.js` in your home directory.

Example configuration:

```javascript

/*
	com.google.Chrome.canary
	com.google.Chrome
	com.apple.Safari
	org.mozilla.firefox
	com.operasoftware.Opera
*/

prefix = "^https?:\/\/"

var config = {
	'com.google.Chrome.canary': [
		"bitbucket\.org",
		"trello\.com",
		"build.gooengine\.com",
		"([a-z]+)?.google\.com",
	],
	'com.google.Chrome': [
		"youtube\.com",
		"facebook\.com",
		"twitter\.com",
	]
};

var format = function(pattern){
	return prefix + pattern;
}

for(browser in config) {
	var patterns = config[browser]
	config[browser] = patterns.map(format)
}

api.config(config)

api.defaultBrowser('com.google.Chrome')

```

#### Usage

Click any link that would start your default browser and Finicky starts the one which pattern matches.

#### Documentation

TODO