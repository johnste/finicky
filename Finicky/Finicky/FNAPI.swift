    //
//  FNAPI.swift
//  Finicky
//
//  Created by John Sterling on 12/06/15.
//  Copyright (c) 2015 John sterling. All rights reserved.
//

import Foundation
import JavaScriptCore

@objc protocol FinickyAPIExports : JSExport {
    static func setDefaultBrowser(browser: String?) -> Void
    static func onUrl(handler: JSValue) -> Void
}

@objc class FinickyAPI : NSObject, FinickyAPIExports {
    
    private static var urlHandlers = Array<JSValue>()
    
    class func setDefaultBrowser(browser: String?) -> Void {
        AppDelegate.defaultBrowser = browser
    }
    
    
    class func onUrl(handler: JSValue) -> Void {
        urlHandlers.append(handler)
    }
    
    class func reset() -> Void {
        urlHandlers.removeAll(keepCapacity: true)
    }
    
    /**
        Get strategy from registered handlers
    
        @param originalUrl The original url that triggered finicky to start
    
        @param sourceBundleIdentifier Bundle identifier of the application that triggered the url to open
    
        @return A dictionary keyed with "url" and "bundleIdentifier" with 
            the new url and bundle identifier to spawn
    */

    class func callUrlHandlers(originalUrl: String, sourceBundleIdentifier: String) -> Dictionary<String, String> {
        var strategy : Dictionary<String, String> = [
            "url": originalUrl,
            "bundleIdentifier": ""
        ]
        
        var options : Dictionary<String, String> = [
            "sourceBundleIdentifier": sourceBundleIdentifier
        ]
        
        for handler in urlHandlers {
            let url = strategy["url"]!
            var val = handler.callWithArguments([url, options])
            
            if val != nil {
                let options = val.toDictionary()
                if options != nil {
                    if (options["url"] != nil) {
                        strategy["url"] = (options["url"] as! String)
                    }
            
                    if (options["bundleIdentifier"] != nil) {
                        strategy["bundleIdentifier"] = (options["bundleIdentifier"] as! String)
                    }
                }
            }
        }        
        return strategy
    }
    
}