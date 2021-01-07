
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

enum Browser: String {
    case Chrome = "com.google.chrome"
    case ChromeCanary = "com.google.chrome.canary"
    case Brave = "com.brave.browser"
    case BraveDev = "com.brave.browser.dev"
    case Safari = "com.apple.safari"
    case Firefox = "org.mozilla.firefox"
    case FirefoxDeveloperEdition = "org.mozilla.firefoxdeveloperedition"
    case Opera = "com.operasoftware.opera"
    case Edge = "com.microsoft.edgemac"
}

public func getBrowserCommand(_ browserOpts: BrowserOpts, url: URL) -> [String] {
    var command = ["open"]
    var commandArgs: [String] = []

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
    }

    if commandArgs.count > 0 {
        command.append("-n")
        command.append("--args")
        command.append(contentsOf: commandArgs)
    }

    // command.append(contentsOf: ["--app-id=jhgifngagadhdmngcahhimadjfofillb", "--profile-directory=Profile 13", "--app-launch-url-for-shortcuts-menu-item=https://meet.google.com/fqu-ckzh-vap?hs=122&ijlm=1609867843360"])

    if browserOpts.passUrlAsArg {
        command.append(url.absoluteString)
    }

    return command
}

private func getProfileOption(bundleId: String, profile: String) -> [String]? {
    var profileOption: [String]? {
        switch bundleId.lowercased() {
        case Browser.Brave.rawValue: return ["--profile-directory=\(profile)"]
        case Browser.BraveDev.rawValue: return ["--profile-directory=\(profile)"]
        case Browser.Chrome.rawValue, Browser.Edge.rawValue: return ["--profile-directory=\(profile)"]

//        Disabling Firefox support due to unreliable performance
//        Link: https://github.com/johnste/finicky/pull/113#issuecomment-672180597
//
//        case Browser.Firefox.rawValue: return ["-P", profile]
//        case Browser.FirefoxDeveloperEdition.rawValue: return ["-P", profile]

        default: return nil
        }
    }
    return profileOption
}
