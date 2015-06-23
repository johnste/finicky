//
//  AppDelegate.swift
//  Finicky
//
//  Created by John Sterling on 04/06/15.
//  Copyright (c) 2015 John Sterling. All rights reserved.
//

import Cocoa
import Foundation

@NSApplicationMain
class AppDelegate: NSObject, NSApplicationDelegate {

    @IBOutlet weak var window: NSWindow!
    @IBOutlet var statusItemMenu: NSMenu!
    
    var statusItem: NSStatusItem!
    var configLoader: FNConfigLoader!
    
    static var config: [String : Array<NSRegularExpression>]! = nil
    static var defaultBrowser: String! = "com.google.Chrome"
    
    func applicationDidFinishLaunching(aNotification: NSNotification) {
        var bundleId = "net.kassett.Finicky"
        LSSetDefaultHandlerForURLScheme("http", bundleId)
        LSSetDefaultHandlerForURLScheme("https", bundleId)
        
        var img: NSImage! = NSImage(named: "statusitem")
        img.setTemplate(true)
        
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
        var url = event!.paramDescriptorForKeyword(AEKeyword(keyDirectObject))!.stringValue
        
        var browserIdentifier : String! = nil
        
        for (browser, patterns) in AppDelegate.config {
            for pattern in patterns {
                let match: NSTextCheckingResult? = pattern.firstMatchInString(url!, options: nil, range: NSMakeRange(0, count(url!)))
                    
                if match != nil {
                    browserIdentifier = browser
                    break
                }
            }
            
            if browserIdentifier != nil {
                break
            }
        }
        
        if browserIdentifier == nil {
            browserIdentifier  = AppDelegate.defaultBrowser
        }
        
        url = FinickyAPI.callUrlHandlers(url!)
        
        openUrlWithBrowser(url!, browser:browserIdentifier)
        
    }
    
    func openUrlWithBrowser(url: String, browser: String) {
        
        var eventDescriptor: NSAppleEventDescriptor? = NSAppleEventDescriptor()

        var errorInfo : NSDictionary? = nil
    
        var appleEventManager:NSAppleEventManager = NSAppleEventManager.sharedAppleEventManager()
        
        var urls = [NSURL(string: url)!]
        NSWorkspace.sharedWorkspace().openURLs(urls, withAppBundleIdentifier: browser, options: NSWorkspaceLaunchOptions.Default, additionalEventParamDescriptor: nil, launchIdentifiers: nil)
    }
    
    func applicationWillFinishLaunching(aNotification: NSNotification) {
        configLoader = FNConfigLoader()
        configLoader.reload()
        
        var appleEventManager:NSAppleEventManager = NSAppleEventManager.sharedAppleEventManager()
        appleEventManager.setEventHandler(self, andSelector: "handleGetURLEvent:withReplyEvent:", forEventClass: AEEventClass(kInternetEventClass), andEventID: AEEventID(kAEGetURL))
    }
    
    func applicationWillTerminate(aNotification: NSNotification) {
        // Insert code here to tear down your application
    }
}

