
import Foundation


enum Browser: String {
    case Chrome = "com.google.Chrome"
    case ChromeCanary = "com.google.Chrome.canary"
    case Brave = "com.brave.Browser"
    case BraveDev = "com.brave.Browser.dev"
    case Safari = "com.apple.Safari"
    case Firefox = "org.mozilla.firefox"
    case FirefoxDeveloperEdition = "org.mozilla.firefoxdeveloperedition"
    case Opera = "com.operasoftware.Opera"
}

public func startBrowser(app: BrowserOpts, url: URL, defaultOpenInBackground: Bool) {
    var command = ["open", url.absoluteString, "-b", app.bundleId]

    if (app.bundleId == Browser.Chrome.rawValue) {
        StartChromeIncognito(app: app, url: url, defaultOpenInBackground: defaultOpenInBackground)
        return
    }


    if app.openInBackground ?? defaultOpenInBackground {
        command.append("-g")
    }

    print(command.joined(separator: " "))
    shell(command)
}




