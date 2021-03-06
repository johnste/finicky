import AppKit
import Cocoa
import Foundation

@NSApplicationMain
class AppDelegate: NSObject, NSApplicationDelegate, NSUserNotificationCenterDelegate, NSTextFieldDelegate, NSTextViewDelegate {
    @IBOutlet var statusItemMenu: NSMenu!
    @IBOutlet var testConfigWindow: NSWindow!
    @IBOutlet var testUrlTextField: NSTextField!
    @IBOutlet var textView: NSTextView!
    @IBOutlet var openConfigMenuItem: NSMenuItem!
    @IBOutlet var createDefaultConfigMenuItem: NSMenuItem!
    @IBOutlet var replaceConfigMenuItem: NSMenuItem!
    @IBOutlet var revealInFinderMenuItem: NSMenuItem!
    @objc var statusItem: NSStatusItem!

    var configLoader: FinickyConfig!
    var settings: Settings!
    var shortUrlResolver = FNShortUrlResolver()

    func applicationWillFinishLaunching(_: Notification) {
        testUrlTextField.delegate = self
        ClearConsole()
        CheckDefaultBrowser()

        NSUserNotificationCenter.default.delegate = self
        let img: NSImage! = NSImage(named: "statusitem")
        img.isTemplate = true

        let invalidImg: NSImage! = NSImage(named: "statusitemerror")
        invalidImg.isTemplate = true

        // Workaround for some bug: -1 instead of NSVariableStatusItemLength
        statusItem = NSStatusBar.system.statusItem(withLength: CGFloat(-1))
        statusItem.menu = statusItemMenu
        (statusItem.button?.cell! as! NSButtonCell).highlightsBy = NSCell.StyleMask.changeBackgroundCellMask
        statusItem.button?.image = invalidImg

        toggleDockIcon(showIcon: false)

        func configureAppOptions(
            hideIcon: Bool,
            shortUrlProviders: [String]?,
            checkForUpdate: Bool
        ) {
            shortUrlResolver = FNShortUrlResolver(shortUrlProviders: shortUrlProviders ?? defaultUrlShorteners)

            if statusItem != nil {
                statusItem.isVisible = !hideIcon
            }

            if checkForUpdate {
                checkForAvailableUpdate(alwaysNotify: false)
            }
        }

        func updateStatus(status: FinickyConfig.Status) {
            switch status {
            case .valid:
                statusItem.button?.image = img
                openConfigMenuItem.isEnabled = true
                createDefaultConfigMenuItem.isEnabled = true
                replaceConfigMenuItem.isEnabled = true
                revealInFinderMenuItem.isEnabled = true
            case .invalid:
                statusItem.button?.image = invalidImg
                openConfigMenuItem.isEnabled = true
                createDefaultConfigMenuItem.isEnabled = true
                replaceConfigMenuItem.isEnabled = true
                revealInFinderMenuItem.isEnabled = true
            case .unavailable:
                statusItem.button?.image = invalidImg
                openConfigMenuItem.isEnabled = false
                createDefaultConfigMenuItem.isEnabled = true
                replaceConfigMenuItem.isEnabled = true
                revealInFinderMenuItem.isEnabled = false
            }
        }

        settings = Settings(userDefaults: .standard)

        if settings.configLocation == nil,
           FileManager.default.fileExists(atPath: FinickyConfig.defaultConfigLocation.path) {
            logToConsole("Found config file at \(FinickyConfig.defaultConfigLocation.path). Saving location to settings.")
            settings.configLocation = FinickyConfig.defaultConfigLocation
        }

        func getConfigPath() -> String? {
            return settings.configLocation?.absoluteURL.path
        }

        configLoader = FinickyConfig(
            configureAppCb: configureAppOptions,
            logCb: logToConsole,
            updateStatusCb: updateStatus,
            configPath: getConfigPath
        )

        let appleEventManager = NSAppleEventManager.shared()
        appleEventManager.setEventHandler(self, andSelector: #selector(AppDelegate.handleGetURLEvent(_:withReplyEvent:)), forEventClass: AEEventClass(kInternetEventClass), andEventID: AEEventID(kAEGetURL))
    }

    func CheckDefaultBrowser() {
        let url = CFURLCreateWithString(kCFAllocatorDefault, "http://" as CFString, nil)
        if let app = CFURLGetString(LSCopyDefaultApplicationURLForURL(url!, .all, nil)?.takeUnretainedValue()) as String? {
            if !app.contains("Finicky.app") {
                logToConsole("Finicky works best when it is set as the default browser")
            }
        }

        let bundleId = "net.kassett.Finicky"
        LSSetDefaultHandlerForURLScheme("http" as CFString, bundleId as CFString)
        LSSetDefaultHandlerForURLScheme("https" as CFString, bundleId as CFString)
        LSSetDefaultHandlerForURLScheme("finicky" as CFString, bundleId as CFString)
        LSSetDefaultHandlerForURLScheme("finickys" as CFString, bundleId as CFString)
    }

    @IBAction func reloadConfig(_: NSMenuItem? = nil) {
        configLoader.listenToChanges(showInitialSuccess: true)
    }

    @IBAction func openConfig(_: NSMenuItem? = nil) {
        guard let location = settings.configLocation else { return }
        NSWorkspace.shared.open(location)
    }

    @IBAction func revealConfigInFinder(_: NSMenuItem? = nil) {
        guard let location = settings.configLocation else { return }
        NSWorkspace.shared.selectFile(location.path, inFileViewerRootedAtPath: location.deletingLastPathComponent().path)
    }

    @IBAction func checkUpdates(_: NSMenuItem? = nil) {
        checkForAvailableUpdate(alwaysNotify: true)
    }

    @IBAction func createDefaultConfig(_: NSMenuItem? = nil) {
        guard let defaultConfigURL = Bundle.main.url(forResource: "defaultConfig", withExtension: "js") else {
            return
        }

        let suggestedConfigLocation = (settings.configLocation ?? FinickyConfig.defaultConfigLocation)

        let savePanel = NSSavePanel()
        savePanel.allowedFileTypes = ["js"]
        savePanel.allowsOtherFileTypes = false
        savePanel.isExtensionHidden = false
        savePanel.canSelectHiddenExtension = false
        savePanel.canCreateDirectories = true
        savePanel.nameFieldStringValue = suggestedConfigLocation.lastPathComponent
        savePanel.directoryURL = suggestedConfigLocation.deletingLastPathComponent()

        let modalResponse = savePanel.runModal()

        if modalResponse == .OK, let url = savePanel.url {
            try? FileManager.default.copyItem(at: defaultConfigURL, to: url)
            settings.configLocation = url
            reloadConfig()
            openConfig()
        }
    }

    @IBAction func replaceConfig(_: NSMenuItem? = nil) {
        let suggestedConfigLocation = (settings.configLocation ?? FinickyConfig.defaultConfigLocation)
        let openPanel = NSOpenPanel()
        openPanel.allowedFileTypes = ["js"]
        openPanel.canChooseFiles = true
        openPanel.canChooseDirectories = false
        openPanel.allowsOtherFileTypes = false
        openPanel.allowsMultipleSelection = false
        openPanel.directoryURL = suggestedConfigLocation.deletingLastPathComponent()

        let modalResponse = openPanel.runModal()

        if modalResponse == .OK, let url = openPanel.url {
            settings.configLocation = url
            reloadConfig()
        }
    }

    func checkForAvailableUpdate(alwaysNotify: Bool = false) {
        checkForUpdate(alwaysNotify) { (version: Version?, _) -> Void in
            if version == nil {
                showNotification(title: "You are running the latest version of Finicky")
                DispatchQueue.main.async {
                    self.logToConsole("You are running the latest version of Finicky")
                }
                return
            }

            if let version = version {
                print("Update available: \(version.title) \(version.version)")
                showNotification(title: "Update available: \(version.title) \(version.version)")
                DispatchQueue.main.async {
                    self.logToConsole("Update available: \(version.title) \(version.version). Download here: https://github.com/johnste/finicky/releases")
                }
            }
        }
    }

    @IBAction func showAboutPanel(_ sender: NSMenuItem) {
        NSApp.activate(ignoringOtherApps: true)
        NSApp.orderFrontStandardAboutPanel(sender)
    }

    @IBAction func showTestConfigWindow(_ sender: Any?) {
        NSApp.activate(ignoringOtherApps: true)
        testConfigWindow.center()
        testConfigWindow.orderFront(sender)
    }

    func logToConsole(_ message: String, clearConsole _: Bool = false) {
        let date = Date()
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
        let dateString = formatter.string(from: date)

        textView.isAutomaticQuoteSubstitutionEnabled = false
        textView.isAutomaticTextReplacementEnabled = false
        textView.isAutomaticDashSubstitutionEnabled = false
        textView.isAutomaticSpellingCorrectionEnabled = false

        textView.string = textView.string + dateString + " - " + message + "\n"
        textView.scrollToEndOfDocument(self)
        textView.isEditable = true
        textView.checkTextInDocument(nil)
        textView.isEditable = false
    }

    @IBAction func ClearConsole(_: Any? = nil) {
        textView.string = ""
    }

    @IBAction func testUrl(_ sender: NSTextField) {
        guard var value = sender.stringValue.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) else {
            logToConsole("Could not parse URL")
            return
        }

        if value.starts(with: "finickys://") || value.starts(with: "finicky://") {
            logToConsole("Finicky will convert finickys:// and finicky:// urls to https:// and http:// respectively")
            value = value.replacingOccurrences(of: "finicky://", with: "http://", options: .literal, range: nil)
            value = value.replacingOccurrences(of: "finickys://", with: "https://", options: .literal, range: nil)
        }

        if !value.starts(with: "https://"), !value.starts(with: "http://") {
            logToConsole("Finicky only understands https:// and http:// urls")
            return
        }

        if let url = URL(string: value) {
            shortUrlResolver.resolveUrl(url, callback: { (URL) -> Void in
                // Dispatch the call to the main thread
                // https://developer.apple.com/documentation/code_diagnostics/main_thread_checker
                DispatchQueue.main.async {
                    self.performTest(url: URL)
                }
            })
        }
    }

    func performTest(url: URL) {
        let opener = Application(pid: getpid())
        guard let appDescriptor = configLoader.determineOpeningApp(url: url, opener: opener) else {
            return
        }

        var description = ""

        if appDescriptor.browsers.count == 1 {
            if let browser = appDescriptor.browsers.first {
                description += "Opens browser: \(browser.name)\(browser.openInBackground ? " (opens in background)" : "")"
            }
        } else if appDescriptor.browsers.count == 0 {
            description += "Won't open any browser"
        } else {
            description += "Opens first active browser of: "
            for (index, browser) in appDescriptor.browsers.enumerated() {
                description += "[\(index)]: \(browser.name) \(browser.openInBackground ? "(opens in background)" : "")"
            }
        }

        description += ", url: \(appDescriptor.url)"
        logToConsole(description)
    }

    @discardableResult
    @objc func toggleDockIcon(showIcon state: Bool) -> Bool {
        var result: Bool
        if state {
            result = NSApp.setActivationPolicy(NSApplication.ActivationPolicy.regular)
        } else {
            result = NSApp.setActivationPolicy(NSApplication.ActivationPolicy.accessory)
        }
        return result
    }

    @objc func handleGetURLEvent(_ event: NSAppleEventDescriptor?, withReplyEvent _: NSAppleEventDescriptor?) {
        toggleDockIcon(showIcon: false)
        let pid = event!.attributeDescriptor(forKeyword: AEKeyword(keySenderPIDAttr))!.int32Value

        let opener = Application(pid: pid)
        var url = URL(string: event!.paramDescriptor(forKeyword: AEKeyword(keyDirectObject))!.stringValue!)!

        if url.scheme == "finicky" || url.scheme == "finickys" {
            if var urlComponents = URLComponents(url: url, resolvingAgainstBaseURL: true) {
                if url.scheme == "finicky" {
                    urlComponents.scheme = "http"
                }

                if url.scheme == "finickys" {
                    urlComponents.scheme = "https"
                }

                url = urlComponents.url!
            }
        }
        shortUrlResolver.resolveUrl(url, callback: { (URL) -> Void in
            self.callUrlHandlers(opener: opener, url: URL)
        })
    }

    func callUrlHandlers(opener: Application, url: URL) {
        if let appDescriptor = configLoader.determineOpeningApp(url: url, opener: opener) {
            if let appToStart = getActiveApp(browsers: appDescriptor.browsers) {
                var success = false
                if let bundleId = appToStart.bundleId {
                    if NSWorkspace.shared.absolutePathForApplication(withBundleIdentifier: bundleId) != nil {
                        openUrlWithBrowser(appDescriptor.url, browserOpts: appToStart)
                        success = true
                    }
                } else if let appPath = appToStart.appPath {
                    if BrowserOpts.isAppDirectory(appPath) {
                        openUrlWithBrowser(appDescriptor.url, browserOpts: appToStart)
                        success = true
                    }
                }
                if !success {
                    let description = "Finicky was unable to find the application \"\(appToStart)\""
                    print(description)
                    logToConsole(description)
                    showNotification(title: "Unable to find application", informativeText: description, error: true)
                }
            }
        }
    }

    func userNotificationCenter(_: NSUserNotificationCenter, shouldPresent _: NSUserNotification) -> Bool {
        return true
    }

    func userNotificationCenter(_: NSUserNotificationCenter, didActivate _: NSUserNotification) {
        showTestConfigWindow(nil)
    }

    func application(_: NSApplication, openFiles filenames: [String]) {
        toggleDockIcon(showIcon: false)
        for filename in filenames {
            let opener = Application(pid: getpid())
            callUrlHandlers(opener: opener, url: URL(fileURLWithPath: filename))
        }
    }

    func application(_: NSApplication, willContinueUserActivityWithType userActivityType: String) -> Bool {
        return userActivityType == NSUserActivityTypeBrowsingWeb
    }

    func application(_: NSApplication, continue userActivity: NSUserActivity, restorationHandler _: @escaping ([NSUserActivityRestoring]) -> Void) -> Bool {
        if userActivity.activityType != NSUserActivityTypeBrowsingWeb {
            return false
        }

        guard let url = userActivity.webpageURL else {
            return false
        }

        let opener = Application(pid: getpid())
        callUrlHandlers(opener: opener, url: url)
        return true
    }

    func application(_: NSApplication, didFailToContinueUserActivityWithType _: String, error _: Error) {}
}
