
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
    case Chromium = "org.chromium.chromium"
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

    if let profileOption: [String] = getProfileOption(browserOpts: browserOpts) {
        commandArgs.append(contentsOf: profileOption)
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

private func getProfileOption(browserOpts: BrowserOpts) -> [String]? {
    var profileOption: [String]? {
        switch browserOpts.bundleId.lowercased() {
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
        case
            Browser.Chromium.rawValue:
            return getChromeProfileDir(browserOpts: browserOpts)
        default: return nil
        }
    }
    return profileOption
}

private func getChromeProfileDir(browserOpts: BrowserOpts) -> [String]? {
    let fileManager = FileManager.default

    // 1. Check if browserOpts.profile is an existing directory
    if let profile = browserOpts.profile, !profile.isEmpty, fileManager.fileExists(atPath: profile) && (try? fileManager.attributesOfItem(atPath: profile)[.type] as? FileAttributeType) == .typeDirectory {
        return ["--profile-directory=\(profile)"]
    }

    // 2. If browserOpts.appPath is empty, return browserOpts.profile (if it's not empty, it will be handled by the first if)
    guard let appPath = browserOpts.appPath, !appPath.isEmpty else {
      if let profile = browserOpts.profile, !profile.isEmpty {
          return ["--profile-directory=\(profile)"]
      } else {
        print("Warning: No app path or valid profile directory provided for Chromium.")
        return nil
      }
    }

    // 3. Loop through directories inside browserOpts.appPath
    guard let contents = try? fileManager.contentsOfDirectory(atPath: appPath) else {
        print("Error: Could not read Chromium profile base directory: \(appPath)")
        return nil
    }

    for dir in contents where (try? fileManager.attributesOfItem(atPath: URL(fileURLWithPath: appPath).appendingPathComponent(dir).path)[.type] as? FileAttributeType) == .typeDirectory {
        let preferencesPath = URL(fileURLWithPath: appPath).appendingPathComponent(dir).appendingPathComponent("Preferences").path

        if fileManager.fileExists(atPath: preferencesPath) {
            do {
                let preferencesData = try Data(contentsOf: URL(fileURLWithPath: preferencesPath))
                if let preferences = try JSONSerialization.jsonObject(with: preferencesData) as? [String: Any],
                   let profile = preferences["profile"] as? [String: Any],
                   let name = profile["name"] as? String,
                   let profileName = browserOpts.profileName, // Access profileName from browserOpts
                   name == profileName {
                    return ["--profile-directory=\(dir)"] // Return the directory name
                }
            } catch {
                print("Error processing profile directory \(dir): \(error)")
                // Handle JSON parsing or file reading errors
            }
        }
    }

    print("Warning: No matching profile directory found in \(appPath) for Chromium.")
    return nil // Profile name not found
}