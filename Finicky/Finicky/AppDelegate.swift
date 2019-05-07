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
        _ = toggleDockIcon(showIcon: false)

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
        formatter.dateFormat = "yyyy-MM-dd hh:mm:ss"
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
            if let appDescriptor = configLoader.determineOpeningApp(url: url, sourceBundleIdentifier: "net.kassett.finicky") {
                var description = ""

                if let openInBackground = appDescriptor.openInBackground {
                    description = """
                    Would open \(AppDescriptorType.bundleId == appDescriptor.appType ? "bundleId" : "") "\(appDescriptor.name)" \(openInBackground ? "application in the background" : "") URL: "\(appDescriptor.url)"
                    """
                } else {
                    description = """
                        Would open \(AppDescriptorType.bundleId == appDescriptor.appType ? "bundleId" : "") "\(appDescriptor.name)" URL: "\(appDescriptor.url)"
                        """
                }
                logToConsole(description)
            }
        }
    }

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

        if shortUrlResolver.isShortUrl(url) {
            shortUrlResolver.resolveUrl(url, callback: {(URL) -> Void in
                self.callUrlHandlers(sourceBundleIdentifier, url: url)
            })
        } else {
            self.callUrlHandlers(sourceBundleIdentifier, url: url)
        }
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
                print ("Finicky was unable to find the application \"" + appDescriptor.name + "\"")
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
        let urls = [url]

        // Launch in background by default if finicky isn't active to avoid something..
        let openInBackground = openInBackground ?? !isActive

        if !openInBackground {
            NSWorkspace.shared.launchApplication(
                withBundleIdentifier: bundleIdentifier,
                options: NSWorkspace.LaunchOptions.default,
                additionalEventParamDescriptor: nil,
                launchIdentifier: nil
            )
        }

        NSWorkspace.shared.open(
            urls,
            withAppBundleIdentifier: bundleIdentifier,
            options: openInBackground ? NSWorkspace.LaunchOptions.withoutActivation : NSWorkspace.LaunchOptions.default,
            additionalEventParamDescriptor: nil,
            launchIdentifiers: nil
        )
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

