
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

public func getBrowserCommand(_ bundleId: String, profileName:String = "Default", url: URL, openInBackground: Bool) -> [String] {
    var command = ["open", url.absoluteString, "-b", bundleId]
    if(bundleId == "com.google.Chrome"){
        command = ["/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",   String(format: "--profile-directory=%@", profileName),  url.absoluteString]
    }

    if openInBackground {
        command.append("-g")
    }

    return command
}
