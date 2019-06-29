
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

public func getBrowserCommand(_ bundleId: String, url: URL, openInBackground: Bool) -> [String] {
    var command = ["open", url.absoluteString, "-b", bundleId]

    if openInBackground {
        command.append("-g")
    }

    return command
}
