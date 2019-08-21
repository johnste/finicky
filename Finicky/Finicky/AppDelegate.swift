import AppKit
import Cocoa
import Foundation

@NSApplicationMain
class AppDelegate: NSObject, NSApplicationDelegate, NSUserNotificationCenterDelegate, NSTextFieldDelegate, NSTextViewDelegate {
    @IBOutlet var statusItemMenu: NSMenu!
    @IBOutlet var testConfigWindow: NSWindow!
    @IBOutlet var yourTextField: NSTextField!
    @IBOutlet var textView: NSTextView!

    @objc var statusItem: NSStatusItem!
    var configLoader: FinickyConfig!
    var shortUrlResolver: FNShortUrlResolver = FNShortUrlResolver()
    @objc var isActive: Bool = true

    func applicationWillFinishLaunching(_: Notification) {
        yourTextField.delegate = self
        let bundleId = "net.kassett.Finicky"
        LSSetDefaultHandlerForURLScheme("http" as CFString, bundleId as CFString)
        LSSetDefaultHandlerForURLScheme("https" as CFString, bundleId as CFString)

        NSUserNotificationCenter.default.delegate = self
        let img: NSImage! = NSImage(named: "statusitem")
        img.isTemplate = true

        let invalidImg: NSImage! = NSImage(named: "statusitemerror")
        invalidImg.isTemplate = true

        let bar = NSStatusBar.system
        // Workaround for some bug: -1 instead of NSVariableStatusItemLength
        statusItem = bar.statusItem(withLength: CGFloat(-1))
        statusItem.menu = statusItemMenu
        statusItem.highlightMode = true
        statusItem.image = invalidImg
        toggleDockIcon(showIcon: false)

        func toggleIconCallback(show: Bool) {
            guard statusItem != nil else { return }
            statusItem.isVisible = !show
        }

        func setShortUrlProviders(shortUrlProviders: [String]?) {
            shortUrlResolver = FNShortUrlResolver(shortUrlProviders: shortUrlProviders)
        }

        func updateStatus(valid: Bool) {
            if valid {
                statusItem.image = img
            } else {
                statusItem.image = invalidImg
            }
        }

        configLoader = FinickyConfig(toggleIconCallback: toggleIconCallback, logToConsoleCallback: logToConsole, setShortUrlProviders: setShortUrlProviders, updateStatus: updateStatus)

        let appleEventManager: NSAppleEventManager = NSAppleEventManager.shared()
        appleEventManager.setEventHandler(self, andSelector: #selector(AppDelegate.handleGetURLEvent(_:withReplyEvent:)), forEventClass: AEEventClass(kInternetEventClass), andEventID: AEEventID(kAEGetURL))
    }

    @IBAction func reloadConfig(_: NSMenuItem) {
        configLoader.listenToChanges(showInitialSuccess: true)
    }

    @IBAction func checkUpdates(_: NSMenuItem) {
        checkForUpdate { (version: Version?) -> Void in
            if version == nil {
                showNotification(title: "No new version available")
                DispatchQueue.main.async {
                    self.logToConsole("No new version available")
                }
                return
            }

            if let version = version {
                print("There is a newer version out there: \(version.title) \(version.version)")
                showNotification(title: "New version available: \(version.title) \(version.version)")
                DispatchQueue.main.async {
                    self.logToConsole("There is a newer version out there: \(version.title) \(version.version). More information is available on https://github.com/johnste/finicky/releases")
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

    func logToConsole(_ message: String) {
        let date = Date()
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
        let dateString = formatter.string(from: date)
        textView.string = dateString + " - " + message + "\n\n" + textView.string.prefix(20000).trimmingCharacters(in: .whitespacesAndNewlines)
    }

    @IBAction func testUrl(_ sender: NSTextField) {
        let value = sender.stringValue

        if !value.starts(with: "https://"), !value.starts(with: "http://") {
            logToConsole("Finicky only understand https:// and http:// urls")
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
        if let appDescriptor = configLoader.determineOpeningApp(url: url, sourceBundleIdentifier: "net.kassett.finicky") {
            var description = """

            Would open url: \(appDescriptor.url)

            """

            if appDescriptor.browsers.count == 1 {
                if let browser = appDescriptor.browsers.first {
                    description += "Browser:\n"
                    description += "    \(browser.name) \(browser.openInBackground ?? false ? "(opens in background)" : "")\n"
                }
            } else if appDescriptor.browsers.count == 0 {
                description += "Won't open any browser"
            } else {
                description += "First active browser of:\n"
                for (index, browser) in appDescriptor.browsers.enumerated() {
                    description += "    [\(index)]: \(browser.name) \(browser.openInBackground ?? false ? "(opens in background)" : "")\n"
                }
            }

            logToConsole(description)
        }
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
        let url: URL = URL(string: event!.paramDescriptor(forKeyword: AEKeyword(keyDirectObject))!.stringValue!)!
        let pid = event!.attributeDescriptor(forKeyword: AEKeyword(keySenderPIDAttr))!.int32Value
        let sourceBundleIdentifier = NSRunningApplication(processIdentifier: pid)?.bundleIdentifier

        shortUrlResolver.resolveUrl(url, callback: { (URL) -> Void in
            self.callUrlHandlers(sourceBundleIdentifier, url: URL)
        })
    }

    func getActiveApp(browsers: [BrowserOpts]) -> BrowserOpts? {
        if browsers.count == 0 {
            return nil
        }

        if browsers.count == 1 {
            return browsers.first
        }

        for browser in browsers {
            let apps = NSRunningApplication.runningApplications(withBundleIdentifier: browser.bundleId)
            if !apps.isEmpty {
                let app: NSRunningApplication = apps[0]
                let bundleIdentifier = app.bundleIdentifier
                if bundleIdentifier != nil {
                    return browser
                }
            }
        }

        // If we are here, no apps are running, so we return the first bundleIds in the array instead.
        return browsers.first
    }

    @objc func callUrlHandlers(_ sourceBundleIdentifier: String?, url: URL) {
        if let appDescriptor = configLoader.determineOpeningApp(url: url, sourceBundleIdentifier: sourceBundleIdentifier) {
            if let appToStart = getActiveApp(browsers: appDescriptor.browsers) {
                if NSWorkspace.shared.absolutePathForApplication(withBundleIdentifier: appToStart.bundleId) != nil {
                    openUrlWithBrowser(appDescriptor.url, bundleIdentifier: appToStart.bundleId, openInBackground: appToStart.openInBackground)
                } else {
                    let description = "Finicky was unable to find the application \"" + appToStart.name + "\""
                    print(description)
                    logToConsole(description)
                    showNotification(title: "Unable to find application", informativeText: "Finicky was unable to find the application \"" + appToStart.name + "\"", error: true)
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

    func openUrlWithBrowser(_ url: URL, bundleIdentifier: String, openInBackground: Bool?) {
        // Launch in background by default if finicky isn't active to avoid something that causes some bug to happen...
        let openInBackground = openInBackground ?? !isActive
        print("Opening " + bundleIdentifier + " at: " + url.absoluteString)
        let command = getBrowserCommand(bundleIdentifier, url: url, openInBackground: openInBackground)
        shell(command)
    }

    func application(_: NSApplication, openFiles filenames: [String]) {
        for filename in filenames {
            callUrlHandlers(nil, url: URL(fileURLWithPath: filename))
        }
    }

    func applicationDidBecomeActive(_: Notification) {
        isActive = true
    }

    func applicationDidResignActive(_: Notification) {
        isActive = false
    }
}
