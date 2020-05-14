
import Foundation

enum Browser: String {
    case Chrome = "com.google.chrome"
    case ChromeCanary = "com.google.chrome.canary"
    case Brave = "com.brave.browser"
    case BraveDev = "com.brave.browser.dev"
    case Safari = "com.apple.safari"
    case Firefox = "org.mozilla.firefox"
    case FirefoxDeveloperEdition = "org.mozilla.firefoxdeveloperedition"
    case Opera = "com.operasoftware.opera"
}

public func getBrowserCommand(_ browserOpts: BrowserOpts, url: URL) -> [String] {
    var command = ["open"]

    // appPath takes priority over bundleId as it is always unique.
    if let appPath = browserOpts.appPath {
        command.append(contentsOf: ["-a", appPath])
    } else if let bundleId = browserOpts.bundleId {
        command.append(contentsOf: ["-b", bundleId])
    } else {}

    if browserOpts.openInBackground {
        command.append("-g")
    }

    if let profile = browserOpts.profile, let bundleId: String = browserOpts.bundleId {
        command.append("-n")
        let profileOption: [String] = getProfileOption(bundleId: bundleId, profile: profile)
        command.append("--args")
        command.append(contentsOf: profileOption)
    }

    command.append(url.absoluteString)

    return command
}

private func getProfileOption(bundleId: String, profile: String) -> [String] {
    var profileOption: [String] {
        switch bundleId.lowercased() {
        case Browser.Brave.rawValue: return ["--profile-directory=\(profile)"]
        case Browser.BraveDev.rawValue: return ["--profile-directory=\(profile)"]
        case Browser.Chrome.rawValue: return ["--profile-directory=\(profile)"]
        case Browser.Firefox.rawValue: return ["-P", profile]
        case Browser.FirefoxDeveloperEdition.rawValue: return ["-P", profile]
        default: return [""]
        }
    }
    return profileOption
}
