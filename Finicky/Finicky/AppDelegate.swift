import Cocoa
import Foundation
import AppKit

@NSApplicationMain
class AppDelegate: NSObject, NSApplicationDelegate, NSUserNotificationCenterDelegate, NSTextFieldDelegate, NSTextViewDelegate {

    @IBOutlet var statusItemMenu: NSMenu!
    @IBOutlet weak var testConfigWindow: NSWindow!
    @IBOutlet weak var yourTextField: NSTextField!
    @IBOutlet weak var textView: NSTextView!

    @objc var statusItem: NSStatusItem!
    var configLoader: FinickyConfig!
    var shortUrlResolver: FNShortUrlResolver!
    @objc var isActive: Bool = true

    func applicationDidFinishLaunching(_ aNotification: Notification) {

        yourTextField.delegate = self
        let bundleId = "net.kassett.Finicky"
        LSSetDefaultHandlerForURLScheme("http" as CFString, bundleId as CFString)
        LSSetDefaultHandlerForURLScheme("https" as CFString, bundleId as CFString)

        NSUserNotificationCenter.default.delegate = self
        let img: NSImage! = NSImage(named: NSImage.Name(rawValue: "statusitem"))
        img.isTemplate = true

        let bar = NSStatusBar.system
        // Workaround for some bug: -1 instead of NSVariableStatusItemLength
        statusItem = bar.statusItem(withLength: CGFloat(-1))
        statusItem.menu = statusItemMenu
        statusItem.highlightMode = true
        statusItem.image = img
        toggleDockIcon(showIcon: false)

        func toggleIconCallback(show: Bool) {
            guard statusItem != nil else { return }
            statusItem.isVisible = !show
        }

        func setShortUrlProviders(shortUrlProviders: [String]?) {
            shortUrlResolver = FNShortUrlResolver(shortUrlProviders: shortUrlProviders)
        }

        configLoader = FinickyConfig(toggleIconCallback: toggleIconCallback, logToConsoleCallback: logToConsole, setShortUrlProviders: setShortUrlProviders)
        configLoader.reload(showSuccess: false)
    }

    @IBAction func reloadConfig(_ sender: NSMenuItem) {
        configLoader.reload(showSuccess: true)
    }

    @IBAction func showAboutPanel(_ sender: NSMenuItem) {
        NSApp.activate(ignoringOtherApps: true)
        NSApp.orderFrontStandardAboutPanel(sender)
    }

    @IBAction func showTestConfigWindow(_ sender: Any?) {
        NSApp.activate(ignoringOtherApps: true)
        self.testConfigWindow.center()
        self.testConfigWindow.orderFront(sender)
    }

    func logToConsole(_ message: String) {
        let date = Date()
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm:ss"
        let dateString = formatter.string(from: date)
        self.textView.string = dateString + " - " + message + "\n\n" + self.textView.string.prefix(20000).trimmingCharacters(in: .whitespacesAndNewlines)
    }

    @IBAction func testUrl(_ sender: NSTextField) {
        let value = sender.stringValue

        if (!value.starts(with: "https://") && !value.starts(with: "http://")) {
            logToConsole("Finicky only understand https:// and http:// urls")
            return
        }
        if let url = URL.init(string: value) {
            shortUrlResolver.resolveUrl(url, callback: {(URL) -> Void in
                self.performTest(url: URL)
            })
        }
    }

    func performTest(url: URL) {
        if let appDescriptor = configLoader.determineOpeningApp(url: url, sourceBundleIdentifier: "net.kassett.finicky") {
            var description = ""

            if let openInBackground = appDescriptor.openInBackground {
                description = """
                Would open \(AppDescriptorType.bundleId == appDescriptor.appType ? "bundleId" : "")\(appDescriptor.name) \(openInBackground ? "application in the background" : "") URL: \(appDescriptor.url)
                """
            } else {
                description = """
                Would open \(AppDescriptorType.bundleId == appDescriptor.appType ? "bundleId" : "")\(appDescriptor.name) URL: \(appDescriptor.url)
                """
            }
            logToConsole(description)
        }
    }

    @discardableResult
    @objc func toggleDockIcon(showIcon state: Bool) -> Bool {
        var result: Bool
        if state {
            result = NSApp.setActivationPolicy(NSApplication.ActivationPolicy.regular)
        }
        else {
            result = NSApp.setActivationPolicy(NSApplication.ActivationPolicy.accessory)
        }
        return result
    }

    @objc func handleGetURLEvent(_ event: NSAppleEventDescriptor?, withReplyEvent: NSAppleEventDescriptor?) {
        let url : URL = URL(string: event!.paramDescriptor(forKeyword: AEKeyword(keyDirectObject))!.stringValue!)!
        let pid = event!.attributeDescriptor(forKeyword: AEKeyword(keySenderPIDAttr))!.int32Value
        let sourceBundleIdentifier = NSRunningApplication(processIdentifier: pid)?.bundleIdentifier

        shortUrlResolver.resolveUrl(url, callback: {(URL) -> Void in
            self.callUrlHandlers(sourceBundleIdentifier, url: URL)
        })
    }

    @objc func callUrlHandlers(_ sourceBundleIdentifier: String?, url: URL) {
        if let appDescriptor = configLoader.determineOpeningApp(url: url, sourceBundleIdentifier: sourceBundleIdentifier) {
            var bundleId : String?

            if (appDescriptor.appType == AppDescriptorType.bundleId) {
                bundleId = appDescriptor.name
            } else {
                if let path = NSWorkspace.shared.fullPath(forApplication: appDescriptor.name) {
                    if let bundle = Bundle(path: path) {
                        bundleId = bundle.bundleIdentifier
                    }
                }
            }

            if bundleId != nil {
                openUrlWithBrowser(appDescriptor.url, bundleIdentifier:bundleId!, openInBackground: appDescriptor.openInBackground )
            } else {
                let description = "Finicky was unable to find the application \"" + appDescriptor.name + "\"";
                print(description)
                logToConsole(description)
                showNotification(title: "Unable to find application", informativeText: "Finicky was unable to find the application \"" + appDescriptor.name + "\"", error: true)
            }
        }
    }

    func userNotificationCenter(_ center: NSUserNotificationCenter, shouldPresent notification: NSUserNotification) -> Bool {
        return true
    }

    func userNotificationCenter(_ center: NSUserNotificationCenter, didActivate notification: NSUserNotification) -> Void {
        showTestConfigWindow(nil)
    }


    func openUrlWithBrowser(_ url: URL, bundleIdentifier: String, openInBackground: Bool?) {
        // Launch in background by default if finicky isn't active to avoid something that causes some bug to happen...
        // Too long ago to remember what actually happened
        let openInBackground = openInBackground ?? !isActive

        print("opening " + url.absoluteString)
        if (openInBackground) {
            shell("open",  url.absoluteString , "-b", bundleIdentifier, "-g")
        } else {
            shell("open",url.absoluteString , "-b", bundleIdentifier)
        }
    }

    func application(_ sender: NSApplication, openFiles filenames: [String]) {
        for filename in filenames {
            self.callUrlHandlers(nil, url: URL(fileURLWithPath: filename ))
        }
    }

    func applicationWillFinishLaunching(_ aNotification: Notification) {
        let appleEventManager:NSAppleEventManager = NSAppleEventManager.shared()
        appleEventManager.setEventHandler(self, andSelector: #selector(AppDelegate.handleGetURLEvent(_:withReplyEvent:)), forEventClass: AEEventClass(kInternetEventClass), andEventID: AEEventID(kAEGetURL))
    }

    func applicationDidBecomeActive(_ aNotification: Notification) {
        isActive = true
    }

    func applicationDidResignActive(_ aNotification: Notification) {
        isActive = false
    }
}

