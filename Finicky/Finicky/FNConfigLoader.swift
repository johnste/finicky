//
//  ConfigLoader.swift
//  Finicky
//
//  Created by John Sterling on 07/06/15.
//  Copyright (c) 2015 John Sterling. All rights reserved.
//

import Foundation
import JavaScriptCore

var FNConfigPath: String = "~/.finicky.js"

public class FNConfigLoader {

    var configPaths: NSMutableSet
    var configWatcher: FNPathWatcher?
    var monitor : FNPathWatcher!
    var ctx: JSContext!

    public init() {
        self.configPaths = NSMutableSet()
    }

    func resetConfigPaths() {
        FinickyAPI.reset()
        configPaths.removeAllObjects()
        configPaths.addObject(FNConfigPath)
    }

    func setupConfigWatcher() {

        let url = NSURL(fileURLWithPath: (FNConfigPath as NSString).stringByExpandingTildeInPath)
        monitor = FNPathWatcher(url: url, handler: {
            self.reload()
        })
        monitor.start()
    }
    
    public func createContext() -> JSContext {
        ctx = JSContext()
        
        ctx.exceptionHandler = {
            context, exception in
            print("JS Error: \(exception)")
        }
        
        self.setupAPI(ctx)
        return ctx
    }
    
    public func parseConfig(config: String) {
        ctx.evaluateScript(config)
    }

    func reload() {
        self.resetConfigPaths()
        var error:NSError?
        let filename: String = (FNConfigPath as NSString).stringByStandardizingPath
        var config: String?
        do {
            config = try String(contentsOfFile: filename, encoding: NSUTF8StringEncoding)
        } catch let error1 as NSError {
            error = error1
            config = nil
        }

        if config == nil {
            print("Config file could not be read or found")
            return
        }

        if let theError = error {
            print("\(theError.localizedDescription)", terminator: "")
        }

        ctx = createContext()
        if config != nil {
            parseConfig(config!)
        }
        setupConfigWatcher()
    }
    
    public func setupAPI(ctx: JSContext) {
        FinickyAPI.setContext(ctx)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "api")
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "finicky")
    }
}