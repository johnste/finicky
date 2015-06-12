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
    
    static var config: [String : Array<String>]!
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
        

        configLoader = FNConfigLoader()
        configLoader.reload()
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
        //println("yay");
        
        var url = event!.paramDescriptorForKeyword(AEKeyword(keyDirectObject))!.stringValue
        
        //println(url!);
        
        var browserIdentifier : String! = nil
        
        for (browser, patterns) in AppDelegate.config {
            for pattern in patterns {
                if url!.rangeOfString(pattern, options: NSStringCompareOptions.CaseInsensitiveSearch) != nil {
                    browserIdentifier = browser
                    break
                }
            }
            if browserIdentifier != nil {
                break
            }
        }
        //println(browserIdentifier)
        if browserIdentifier == nil {
            browserIdentifier  = AppDelegate.defaultBrowser
        }
        //println(browserIdentifier)
        var appleEventManager:NSAppleEventManager = NSAppleEventManager.sharedAppleEventManager()
        var urls = [NSURL(string: url!)!]
        NSWorkspace.sharedWorkspace().openURLs(urls, withAppBundleIdentifier: browserIdentifier, options: NSWorkspaceLaunchOptions.Default, additionalEventParamDescriptor: nil, launchIdentifiers: nil)
        
    }
    
    func applicationWillFinishLaunching(aNotification: NSNotification) {
        var appleEventManager:NSAppleEventManager = NSAppleEventManager.sharedAppleEventManager()
        appleEventManager.setEventHandler(self, andSelector: "handleGetURLEvent:withReplyEvent:", forEventClass: AEEventClass(kInternetEventClass), andEventID: AEEventID(kAEGetURL))
        
    }
    
    func applicationWillTerminate(aNotification: NSNotification) {
        // Insert code here to tear down your application
    }


}

