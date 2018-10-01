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

open class FNConfigLoader {

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
        configPaths.add(FNConfigPath)
    }

    func setupConfigWatcher() {

        let url = URL(fileURLWithPath: (FNConfigPath as NSString).expandingTildeInPath)
        monitor = FNPathWatcher(url: url, handler: {
            self.reload()
        })
        monitor.start()
    }
    
    open func createContext() -> JSContext {
        ctx = JSContext()
        
        ctx.exceptionHandler = {
            context, exception in
            print("JS Error: \(exception)")
        }
        
        self.setupAPI(ctx)
        return ctx
    }
    
    open func parseConfig(_ config: String) {
        ctx.evaluateScript(config)
    }

    func reload() {
        self.resetConfigPaths()
        var error:NSError?
        let filename: String = (FNConfigPath as NSString).standardizingPath
        var config: String?
        do {
            config = try String(contentsOfFile: filename, encoding: String.Encoding.utf8)
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
    
    open func setupAPI(_ ctx: JSContext) {
        FinickyAPI.setContext(ctx)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "api" as NSCopying & NSObjectProtocol)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "finicky" as NSCopying & NSObjectProtocol)
    }
}
