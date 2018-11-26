import Cocoa
import Foundation
import AppKit

@NSApplicationMain
class AppDelegate: NSObject, NSApplicationDelegate, NSUserNotificationCenterDelegate {

    @IBOutlet weak var window: NSWindow!
    @IBOutlet var statusItemMenu: NSMenu!

    @objc var statusItem: NSStatusItem!
    var configLoader: FinickyConfig!
    var shortUrlResolver: FNShortUrlResolver!
    @objc var urlsToLoad = Array<String>()
    @objc var isActive: Bool = true

    // @objc static var defaultBrowser: String = "com.google.Chrome"

    func applicationDidFinishLaunching(_ aNotification: Notification) {
        let bundleId = "net.kassett.Finicky"
        LSSetDefaultHandlerForURLScheme("http" as CFString, bundleId as CFString)
        LSSetDefaultHandlerForURLScheme("https" as CFString, bundleId as CFString)

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
        configLoader.reload()        
    }

    @IBAction func showAboutPanel(_ sender: NSMenuItem) {
        NSApp.orderFrontStandardAboutPanel(sender)
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
            callUrlHandlers(sourceBundleIdentifier, url: url)
        }
    }

    @objc func callUrlHandlers(_ sourceBundleIdentifier: String?, url: URL) {
        let flags = getFlags()
        // var bundleIdentifier : String = AppDelegate.defaultBrowser
        var newUrl : URL = url
        var openInBackground : Bool? = nil
        

        let app = configLoader.determineOpeningApp(url: url)

        print(app)

//        let strategy = FinickyAPI.callUrlHandlers(newUrl, sourceBundleIdentifier: sourceBundleIdentifier, flags: flags)
//        print("opening \"\(url as Any)\" from \(String(describing: sourceBundleIdentifier!)) as \(strategy["url"]! as Any) in \(strategy["bundleIdentifier"] as Any)");
//        if strategy["url"] != nil {
//            newUrl = URL(string: strategy["url"]! as! String)!
//
//            if let bundleId : String = strategy["bundleIdentifier"] as? String {
//                if !bundleId.isEmpty {
//                    bundleIdentifier = (strategy["bundleIdentifier"]! as! String)
//                }
//            }
//
//            if strategy["openInBackground"] != nil {
//                openInBackground = (strategy["openInBackground"]! as! Bool)
//            }
//
//            if !bundleIdentifier.isEmpty {
//
//                let s = NSWorkspace.shared.fullPath(forApplication: "Sublime Text")
//                let y = Bundle(path: s!)
//                NSLog(y!.bundleIdentifier!)
//
//                openUrlWithBrowser(newUrl, bundleIdentifier:bundleIdentifier, openInBackground: openInBackground)
//            }
//        }
    }

    func userNotificationCenter(_ center: NSUserNotificationCenter, shouldPresent notification: NSUserNotification) -> Bool {
        return true
    }

    func openUrlWithBrowser(_ url: URL, bundleIdentifier: String, openInBackground: Bool?) {
        let urls = [url]

        // Launch in background by default if finicky isn't active to avoid something..
        var launchInBackground = !isActive
        if openInBackground != nil {
            launchInBackground = openInBackground!
        }
        
        if !launchInBackground {
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
            options: launchInBackground ? NSWorkspace.LaunchOptions.withoutActivation : NSWorkspace.LaunchOptions.default,
            additionalEventParamDescriptor: nil,
            launchIdentifiers: nil
        )
    }

    @objc func getFlags() -> Dictionary<String, Bool> {
        return [
            "cmd": NSEvent.modifierFlags.intersection(.command) != [],
            "ctrl": NSEvent.modifierFlags.intersection(.control) != [],
            "shift": NSEvent.modifierFlags.intersection(.shift) != [],
            "alt": NSEvent.modifierFlags.intersection(.option) != []
        ]
    }

    func application(_ sender: NSApplication, openFiles filenames: [String]) {
        for filename in filenames {
            callUrlHandlers(nil, url: URL(fileURLWithPath: filename ))
        }
    }

    func applicationWillFinishLaunching(_ aNotification: Notification) {
        configLoader = FinickyConfig()
        configLoader.reload()
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

