
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

public func getBrowserCommand(_ browserOpts: BrowserOpts, url: URL) -> [String] {
    var command = ["open", url.absoluteString]
    
    // appPath takes priority over bundleId as it is always unique.
    if let appPath = browserOpts.appPath {
        command.append(contentsOf: ["-a", appPath])
    } else if let bundleId = browserOpts.bundleId {
        command.append(contentsOf: ["-b", bundleId])
    }

    if browserOpts.openInBackground {
        command.append("-g")
    }
    
    print(command)

    return command
}
