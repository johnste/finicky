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
    static func setDefaultBrowser(_ browser: String?) -> Void
    static func log(_ message: String?) -> Void
    static func onUrl(_ handler: JSValue) -> Void
}

@objc open class FinickyAPI : NSObject, FinickyAPIExports {

    fileprivate static var urlHandlers = Array<JSValue>()
    fileprivate static var context : JSContext! = nil

    class func setDefaultBrowser(_ browser: String?) -> Void {
        AppDelegate.defaultBrowser = browser
    }

    static func log(_ message: String?) -> Void {
        if message != nil {
            NSLog(message!)
        }
    }

    open class func onUrl(_ handler: JSValue) -> Void {
        urlHandlers.append(handler)
    }

    @objc open class func reset() -> Void {
        urlHandlers.removeAll(keepingCapacity: true)
    }
    
    @objc class func setContext(_ context: JSContext) {
        self.context = context
    }

    /**
        Get strategy from registered handlers

        @param originalUrl The original url that triggered finicky to start

        @param sourceBundleIdentifier Bundle identifier of the application that triggered the url to open

        @return A dictionary keyed with "url" and "bundleIdentifier" with
            the new url and bundle identifier to spawn
    */

    @objc open class func callUrlHandlers(_ originalUrl: URL, sourceBundleIdentifier: String?, flags : Dictionary<String, Bool>) -> Dictionary<String, AnyObject> {
        var strategy : Dictionary<String, AnyObject> = [
            "url": originalUrl.absoluteString as AnyObject
        ]
        
        var sourceBundleId : AnyObject
        if sourceBundleIdentifier != nil {
            sourceBundleId = sourceBundleIdentifier! as AnyObject
        } else {
            sourceBundleId = JSValue(nullIn: context)
        }
    
        let options : Dictionary<String, AnyObject> = [
            "sourceBundleIdentifier": sourceBundleId,
            "flags": flags as AnyObject
        ]

        var previousStrategy : [AnyHashable: Any]! = strategy
        for handler in urlHandlers {
            let url = strategy["url"]! as! String
            let val = handler.call(withArguments: [url, options, previousStrategy])

            if !(val?.isUndefined)! {
                let handlerStrategy = val?.toDictionary()
                if handlerStrategy != nil {
                    if handlerStrategy?["url"] != nil {
                        strategy["url"] = (handlerStrategy?["url"] as! String as AnyObject)
                    }

                    let bundleId: AnyObject? = handlerStrategy?["bundleIdentifier"] as AnyObject
                    if bundleId != nil {
                        strategy["bundleIdentifier"] = bundleId!
                    }
                    
                    if handlerStrategy?["openInBackground"] != nil {
                        strategy["openInBackground"] = (handlerStrategy?["openInBackground"] as! Bool as AnyObject)
                    }

                    if handlerStrategy?["last"] != nil {
                        break
                    }
                }
                previousStrategy = strategy
            }
        }
        return strategy
    }
}
