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
    static func defaultBrowser(browser: String?) -> Void
    static func config(data: [NSObject : AnyObject]!) -> String
    static func onUrl(handler: JSValue) -> Void
}

@objc class FinickyAPI : NSObject, FinickyAPIExports {
    
    private static var urlHandlers = Array<JSValue>()
    
    class func defaultBrowser(browser: String?) -> Void {
        AppDelegate.defaultBrowser = browser
    }
    
    class func config(data: [NSObject : AnyObject]!) -> String {
        return "deprecated"
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
    
        @return A dictionary keyed with "url" and "bundleIdentifier" with 
            the new url and bundle identifier to spawn
    */
    func
    class func callUrlHandlers(originalUrl: String) -> Dictionary<String, String> {
        var strategy : Dictionary<String, String> = [
            "url": originalUrl,
            "bundleIdentifier": ""
        ]
        
        for handler in urlHandlers {
            let url = strategy["url"]!
            var val = handler.callWithArguments([url])
            
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