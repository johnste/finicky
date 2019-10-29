
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

    if (app.bundleId == Browser.Chrome.rawValue) {
        let chromeScript = """
            tell application "Google Chrome"
                set incognitoWindows to (every window whose mode is "incognito")
                set myLink to "http://example.com/?shark=yeswindow"
                if (count of incognitoWindows) > 0 then
                    set incognitoWindow to first window whose mode is "incognito"
                    tell incognitoWindow to make new tab at after (get active tab) with properties {URL:myLink}
                else
                    set incognitoWindow to make new window with properties {mode:"incognito"}
                    tell incognitoWindow to set URL of active tab to myLink
                end if
            end tell
        """
        executeScript(chromeScript)
        return
    }

    var command = ["open", url.absoluteString, "-b", app.bundleId]

    if app.openInBackground ?? defaultOpenInBackground {
        command.append("-g")
    }

    shell(command)
}

func executeScript(_ script: String) {
    var error: NSDictionary?
    if let scriptObject = NSAppleScript(source: script) {
        if let output: NSAppleEventDescriptor = scriptObject.executeAndReturnError(
                                                                           &error) {
            print(output.stringValue)
        } else if (error != nil) {
            showNotification(title: "error: \(error)")
        }
    }
}


