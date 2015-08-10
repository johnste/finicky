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

        let url = NSURL(fileURLWithPath: FNConfigPath.stringByExpandingTildeInPath)!
        monitor = FNPathWatcher(url: url, handler: {
            self.reload()
        })
        monitor.start()
    }
    
    public func createContext() -> JSContext {
        ctx = JSContext()
        
        ctx.exceptionHandler = {
            context, exception in
            println("JS Error: \(exception)")
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
        let filename: String = FNConfigPath.stringByStandardizingPath
        var config: String? = String(contentsOfFile: filename, encoding: NSUTF8StringEncoding, error: &error)

        if config == nil {
            println("Config file could not be read or found")
            return
        }

        if let theError = error {
            print("\(theError.localizedDescription)")
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