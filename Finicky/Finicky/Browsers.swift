
import AppKit
import Foundation

public func getActiveApp(browsers: [BrowserOpts]) -> BrowserOpts? {
    if browsers.count == 0 {
        return nil
    }

    if browsers.count == 1 {
        return browsers.first
    }

    for browser in browsers {
        if let bundleId = browser.bundleId {
            let apps = NSRunningApplication.runningApplications(withBundleIdentifier: bundleId)
            if !apps.isEmpty {
                let app: NSRunningApplication = apps[0]
                let bundleIdentifier = app.bundleIdentifier
                if bundleIdentifier != nil {
                    return browser
                }
            }
        }
    }

    // If we are here, no apps are running, so we return the first bundleIds in the array instead.
    return browsers.first
}

public func openUrlWithBrowser(_ url: URL, browserOpts: BrowserOpts) {
    print("Opening \(browserOpts) at: " + url.absoluteString)
    let command = getBrowserCommand(browserOpts, url: url)
    shell(command)
}

// keep all browser bundle IDs lowercase
enum Browser: String {
    case Blisk = "org.blisk.blisk"
    case Brave = "com.brave.browser"
    case BraveBeta = "com.brave.browser.beta"
    case BraveDev = "com.brave.browser.dev"
    case Chrome = "com.google.chrome"
    case ChromeCanary = "com.google.chrome.canary"
    case Edge = "com.microsoft.edgemac"
    case EdgeBeta = "com.microsoft.edgemac.beta"
    case Firefox = "org.mozilla.firefox"
    case FirefoxDeveloperEdition = "org.mozilla.firefoxdeveloperedition"
    case Opera = "com.operasoftware.opera"
    case Vivaldi = "com.vivaldi.vivaldi"
    case Safari = "com.apple.safari"
    case Wavebox = "com.bookry.wavebox"
}

public func getBrowserCommand(_ browserOpts: BrowserOpts, url: URL) -> [String] {
    var command = ["open"]
    var commandArgs: [String] = []
    var appendUrl = true

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
        if let profileOption: [String] = getProfileOption(bundleId: bundleId, profile: profile) {
            commandArgs.append(contentsOf: profileOption)
        }
    }

    if browserOpts.args.count > 0 {
        commandArgs.append(contentsOf: browserOpts.args)

        // do not auto-append the URL when args has been explicitly defined
        appendUrl = false
    }

    if commandArgs.count > 0 {
        command.append("-n")
        command.append("--args")
        command.append(contentsOf: commandArgs)
    }

    if appendUrl {
        command.append(url.absoluteString)
    }

    return command
}

private func getProfileOption(bundleId: String, profile: String) -> [String]? {
    var profileOption: [String]? {
        switch bundleId.lowercased() {
        case
            Browser.Brave.rawValue,
            Browser.BraveBeta.rawValue,
            Browser.BraveDev.rawValue,
            Browser.Chrome.rawValue,
            Browser.Edge.rawValue,
            Browser.EdgeBeta.rawValue,
            Browser.Vivaldi.rawValue,
            Browser.Wavebox.rawValue:
            return ["--profile-directory=\(profile)"]

            // Blisk and Opera doesn't support multiple profiles even though they are Chromium based
            // case Browser.Blisk.rawValue: return ["--profile-directory=\(profile)"]
            // case Browser.Opera.rawValue: return ["--profile-directory=\(profile)"]

            // Disabling Firefox support due to unreliable performance
            // Link: https://github.com/johnste/finicky/pull/113#issuecomment-672180597
            //
            // case Browser.Firefox.rawValue: return ["-P", profile]
            // case Browser.FirefoxDeveloperEdition.rawValue: return ["-P", profile]

        default: return nil
        }
    }
    return profileOption
}
