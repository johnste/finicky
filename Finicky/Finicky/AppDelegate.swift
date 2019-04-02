import Cocoa
import Foundation
import AppKit

@NSApplicationMain
class AppDelegate: NSObject, NSApplicationDelegate, NSUserNotificationCenterDelegate, NSTextFieldDelegate {

    @IBOutlet var statusItemMenu: NSMenu!
    @IBOutlet weak var testConfigWindow: NSWindow!
    @IBOutlet weak var yourTextField: NSTextField!
    @IBOutlet weak var label: NSTextField!

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
    }

    @IBAction func reloadConfig(_ sender: NSMenuItem) {
        configLoader.reload(showSuccess: true)
    }

    @IBAction func showAboutPanel(_ sender: NSMenuItem) {
        NSApp.activate(ignoringOtherApps: true)
        NSApp.orderFrontStandardAboutPanel(sender)
    }

    @IBAction func showTestConfigWindow(_ sender: NSMenuItem) {
        NSApp.activate(ignoringOtherApps: true)
        self.testConfigWindow.center()
        self.testConfigWindow.orderFront(sender)
    }

    override func controlTextDidChange(_ obj: Notification) {
        let object = obj.object as! NSTextField
        let value = object.stringValue

        if (!value.starts(with: "https://") && !value.starts(with: "http://")) {
            self.label.stringValue = ""
            return
        }
        if let url = URL.init(string: value) {
            if let appDescriptor = configLoader.determineOpeningApp(url: url) {
                var description = ""

                if let options = appDescriptor.options {
                    description = """
                    Will open:
                    Application \(AppDescriptorType.bundleId == appDescriptor.type ? "bundleId" : "") "\(appDescriptor.value)"
                    URL: "\(appDescriptor.url)"
                    Will \(options.openInBackground ? "open" : "not open") application in the background
                    """
                } else {
                    description = """
                        Will open:
                        Application \(AppDescriptorType.bundleId == appDescriptor.type ? "bundleId" : "") "\(appDescriptor.value)"
                        URL: "\(appDescriptor.url)"
                        """
                }
                self.label.stringValue = description
            } else {
                self.label.stringValue = ""
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
        if let appDescriptor = configLoader.determineOpeningApp(url: url) {
            var bundleId : String?

            if (appDescriptor.type == AppDescriptorType.bundleId) {
                bundleId = appDescriptor.value
            } else {
                if let path = NSWorkspace.shared.fullPath(forApplication: appDescriptor.value) {
                    if let bundle = Bundle(path: path) {
                        bundleId = bundle.bundleIdentifier
                    }
                }
            }

            if bundleId != nil {
                openUrlWithBrowser(appDescriptor.url, bundleIdentifier:bundleId!, options: appDescriptor.options )
            } else {
                print ("Finicky was unable to find the application \"" + appDescriptor.value + "\"")
                showNotification(title: "Unable to find application", informativeText: "Finicky was unable to find the application \"" + appDescriptor.value + "\"")
            }
        }
    }

    func userNotificationCenter(_ center: NSUserNotificationCenter, shouldPresent notification: NSUserNotification) -> Bool {
        return true
    }

    func openUrlWithBrowser(_ url: URL, bundleIdentifier: String, options: AppDescriptorOptions?) {
        let urls = [url]

        // Launch in background by default if finicky isn't active to avoid something..
        var openInBackground = !isActive
        if options != nil {
            openInBackground = (options?.openInBackground)!
        }

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
        configLoader = FinickyConfig()
        configLoader.reload(showSuccess: false)
        shortUrlResolver = FNShortUrlResolver()
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

