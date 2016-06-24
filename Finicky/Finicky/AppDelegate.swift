//
//  AppDelegate.swift
//  Finicky
//
//  Created by John Sterling on 04/06/15.
//  Copyright (c) 2015 John Sterling. All rights reserved.
//

import Cocoa
import Foundation
import AppKit

@NSApplicationMain
class AppDelegate: NSObject, NSApplicationDelegate {

    @IBOutlet weak var window: NSWindow!
    @IBOutlet var statusItemMenu: NSMenu!

    var statusItem: NSStatusItem!
    var configLoader: FNConfigLoader!
    var shortUrlResolver: FNShortUrlResolver!
    var urlsToLoad = Array<String>()
    var isActive: Bool = true

    static var defaultBrowser: String! = "com.google.Chrome"

    func applicationDidFinishLaunching(aNotification: NSNotification) {
        let bundleId = "net.kassett.Finicky"
        LSSetDefaultHandlerForURLScheme("http", bundleId)
        LSSetDefaultHandlerForURLScheme("https", bundleId)

        let img: NSImage! = NSImage(named: "statusitem")
        img.template = true

        let bar = NSStatusBar.systemStatusBar()        
        // Workaround for some bug: -1 instead of NSVariableStatusItemLength
        statusItem = bar.statusItemWithLength(CGFloat(-1))
        statusItem.menu = statusItemMenu
        statusItem.highlightMode = true
        statusItem.image = img
        toggleDockIcon(showIcon: false)
    }

    @IBAction func reloadConfig(sender: NSMenuItem) {
        configLoader.reload()
    }

    @IBAction func showAboutPanel(sender: NSMenuItem) {
        NSApp.orderFrontStandardAboutPanel(sender)
    }

    func toggleDockIcon(showIcon state: Bool) -> Bool {
        var result: Bool
        if state {
            result = NSApp.setActivationPolicy(NSApplicationActivationPolicy.Regular)
        }
        else {
            result = NSApp.setActivationPolicy(NSApplicationActivationPolicy.Accessory)
        }
        return result
    }

    func handleGetURLEvent(event: NSAppleEventDescriptor?, withReplyEvent: NSAppleEventDescriptor?) {
        let url : NSURL = NSURL(string: event!.paramDescriptorForKeyword(AEKeyword(keyDirectObject))!.stringValue!)!
        let pid = event!.attributeDescriptorForKeyword(AEKeyword(keySenderPIDAttr))!.int32Value
        let sourceBundleIdentifier = NSRunningApplication(processIdentifier: pid)?.bundleIdentifier

        let callback = callUrlHandlers(sourceBundleIdentifier)

        if shortUrlResolver.isShortUrl(url) {
            shortUrlResolver.resolveUrl(url, callback: callback)
        } else {
            callback(url: url)
        }
    }

    func getActiveApp(bundleIds: Array<String>) -> String {
        for bundleId in bundleIds {
            let apps = NSRunningApplication.runningApplicationsWithBundleIdentifier(bundleId)
            if !apps.isEmpty {
                let app : NSRunningApplication = apps[0]
                let bundleIdentifier = app.bundleIdentifier
                if bundleIdentifier != nil {
                    return bundleIdentifier!
                }
            }
        }

        // If we are here, no apps are running, so we return the first bundleIds in the array instead.
        return bundleIds.first!
    }

    func callUrlHandlers(sourceBundleIdentifier: String?)(url: NSURL) {
        let flags = getFlags()
        var bundleIdentifier : String! = AppDelegate.defaultBrowser
        var newUrl : NSURL = url
        var openInBackground : Bool? = nil

        let strategy = FinickyAPI.callUrlHandlers(newUrl, sourceBundleIdentifier: sourceBundleIdentifier, flags: flags)
        print("opening %@ from %@ as %@ in %@", url, bundleIdentifier, strategy["url"], strategy["bundleIdentifier"]);
        if strategy["url"] != nil {
            newUrl = NSURL(string: strategy["url"]! as! String)!

            // If the bundle identifier is a string, open the url with that app. If it's an array, find the first running
            // app, and open the url with that. If none of the apps are running, use the first available one instead.
            if let bundleId : String! = strategy["bundleIdentifier"] as? String! {
                if bundleId != nil && !bundleId.isEmpty {
                    bundleIdentifier = strategy["bundleIdentifier"]! as! String
                }
            } else if let bundleIds = strategy["bundleIdentifier"] as? Array<String>! {
                bundleIdentifier = getActiveApp(bundleIds)
            }
            
            if strategy["openInBackground"] != nil {
                openInBackground = (strategy["openInBackground"]! as! Bool)
            }

            if bundleIdentifier != nil && !bundleIdentifier.isEmpty {
                openUrlWithBrowser(newUrl, bundleIdentifier:bundleIdentifier, openInBackground: openInBackground)
            }
        }
    }

    func openUrlWithBrowser(url: NSURL, bundleIdentifier: String, openInBackground: Bool?) {
        var eventDescriptor: NSAppleEventDescriptor? = NSAppleEventDescriptor()
        var errorInfo : NSDictionary? = nil
        var appleEventManager:NSAppleEventManager = NSAppleEventManager.sharedAppleEventManager()
        let urls = [url]

        var launchInBackground = !isActive
        if openInBackground != nil {
            launchInBackground = openInBackground!
        }
        
        NSWorkspace.sharedWorkspace().openURLs(
            urls,
            withAppBundleIdentifier: bundleIdentifier,
            options: launchInBackground ? NSWorkspaceLaunchOptions.WithoutActivation : NSWorkspaceLaunchOptions.Default,
            additionalEventParamDescriptor: nil,
            launchIdentifiers: nil
        )
    }

    func getFlags() -> Dictionary<String, Bool> {
        return [
            "cmd": NSEvent.modifierFlags().intersect(.CommandKeyMask) != [],
            "ctrl": NSEvent.modifierFlags().intersect(.ControlKeyMask) != [],
            "shift": NSEvent.modifierFlags().intersect(.ShiftKeyMask) != [],
            "alt": NSEvent.modifierFlags().intersect(.AlternateKeyMask) != []
        ]
    }
    
    func application(sender: NSApplication, openFiles filenames: [String]) {
        for filename in filenames {
            callUrlHandlers(nil)(url: NSURL(fileURLWithPath: filename ))
        }
    }

    func applicationWillFinishLaunching(aNotification: NSNotification) {
        configLoader = FNConfigLoader()
        configLoader.reload()
        shortUrlResolver = FNShortUrlResolver()
        let appleEventManager:NSAppleEventManager = NSAppleEventManager.sharedAppleEventManager()
        appleEventManager.setEventHandler(self, andSelector: "handleGetURLEvent:withReplyEvent:", forEventClass: AEEventClass(kInternetEventClass), andEventID: AEEventID(kAEGetURL))
    }

    func applicationDidBecomeActive(aNotification: NSNotification) {
        isActive = true
    }
    
    func applicationDidResignActive(aNotification: NSNotification) {
        isActive = false
    }

}

