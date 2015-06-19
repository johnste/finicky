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
    static func test(testString: String?) -> Void
    static func defaultBrowser(browser: String?) -> Void
    static func config(data: [NSObject : AnyObject]!) -> Void
}

@objc class FinickyAPI : NSObject, FinickyAPIExports {
    
    class func test(testString: String?) -> Void {
        println("Test: \(testString)")
    }
    
    class func defaultBrowser(browser: String?) -> Void {
        AppDelegate.defaultBrowser = browser
    }
    
    class func config(data: [NSObject : AnyObject]!) -> Void {
        //println("Test: \(data)")
        let options : NSRegularExpressionOptions = NSRegularExpressionOptions.CaseInsensitive
        var config = [String: Array<NSRegularExpression>]()
        var error:NSError?
        for(key, value) in data {
            
            let browser = key as! String
            var patterns = Array<NSRegularExpression>()
            
            let rules : Array = (value as! NSArray) as Array
            for rule in rules {
                let pattern = rule as! String
                var regex = NSRegularExpression(pattern: pattern, options: options, error: &error)
            
                if let theError = error {
                    print("\(theError.localizedDescription)")
                } else {
                    patterns.append(regex!)
                }
            }
            config.updateValue(patterns, forKey: browser)
        }
        
        AppDelegate.config = config
        
    }
}