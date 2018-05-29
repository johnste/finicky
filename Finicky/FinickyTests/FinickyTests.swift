//
//  FinickyTests.swift
//  FinickyTests
//
//  Created by John Sterling on 04/06/15.
//  Copyright (c) 2015 John sterling. All rights reserved.
//

import Cocoa
import XCTest
import Finicky
import JavaScriptCore

class FinickyTests: XCTestCase {
    
    var ctx : JSContext!
    var configLoader: FNConfigLoader = FNConfigLoader()
    let exampleUrl = URL(string: "http://example.com")!
    
    override func setUp() {
        super.setUp()
        configLoader.createContext()
        FinickyAPI.reset()
        // Put setup code here. This method is called before the invocation of each test method in the class.
    }
    
    override func tearDown() {
        // Put teardown code here. This method is called after the invocation of each test method in the class.
        super.tearDown()
    }
    
    func getFlags(_ cmd: Bool = false, ctrl: Bool = false, shift: Bool = false, alt: Bool = false) -> Dictionary<String, Bool> {
        return [
            "cmd": cmd,
            "ctrl": ctrl,
            "shift": shift,
            "alt": alt
        ]
    }
 
    func testDefaults() {
        configLoader.parseConfig(
            "finicky.onUrl(function(url, opts) {}); " +
            "finicky.onUrl(function(url, opts) {}); "
        )
        
        let strategy = FinickyAPI.callUrlHandlers(exampleUrl, sourceBundleIdentifier: "", flags: getFlags())
        let bundleId = strategy["bundleIdentifier"] as! String!
        let strategyUrl : String! = strategy["url"] as! String!
        let bool = strategy["openInBackground"] as! Bool?
        XCTAssertNil(bundleId,  "Bundle ID should not have been set")
        XCTAssertEqual(strategyUrl, "http://example.com", "URL should not have been changed")
        XCTAssertNil(bool, "openInBackground should not have been set")
    }
    
    func testSetAll() {
        configLoader.parseConfig(
        "   finicky.onUrl(function(url, opts) { finicky.log('test'); " +
        "        return { " +
        "           bundleIdentifier: 'com.google.Chrome.canary', " +
        "           url: 'http://example.org', " +
        "           openInBackground: false " +
        "        }" +
        "   });"
        )
        
        let strategy = FinickyAPI.callUrlHandlers(exampleUrl, sourceBundleIdentifier: "", flags: getFlags())
        let bundleId = strategy["bundleIdentifier"] as! String!
        let strategyUrl : String! = strategy["url"] as! String!
        let bool = strategy["openInBackground"] as! Bool!
        XCTAssertEqual(bundleId, "com.google.Chrome.canary", "Bundle ID should have been set")
        XCTAssertEqual(strategyUrl, "http://example.org", "URL should have been changed")
        XCTAssertEqual(bool, false, "openInBackground should have been set to false")
    }
    
    func testLastFlag() {
        configLoader.parseConfig(
            "   finicky.onUrl(function(url, opts) { " +
            "        return { " +
            "           bundleIdentifier: 'com.example.test', " +
            "           last: true, " +
            "        }" +
            "   });" +
            "   finicky.onUrl(function(url, opts) { " +
            "        return { " +
            "           bundleIdentifier: 'com.example.test2', " +
            "        }" +
            "   });"
        )
        
        let strategy = FinickyAPI.callUrlHandlers(exampleUrl, sourceBundleIdentifier: "", flags: getFlags())
        let bundleId = strategy["bundleIdentifier"] as! String!
        XCTAssertEqual(bundleId, "com.example.test", "Last flag should be respected")
    }
}
