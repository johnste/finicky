import AppKit
//
//  App.swift
//  Finicky
//
//  Created by John Sterling on 2021-01-29.
//  Copyright © 2021 John sterling. All rights reserved.
//
// Represents an application

import Foundation

class Application {
    
    var application: NSRunningApplication?;
    var path: String?;
    var pid: Int32;
    
    init(pid: Int32) {
        let application = NSRunningApplication(processIdentifier: pid)
        self.path = getPidPath(pid: pid)
        self.pid = pid
        self.application = application
    }
    
    func serialize() -> [String: Any] {
        return [
            "pid": self.pid,
            "path": self.path!,
            "bundleId": (self.application?.bundleIdentifier! ?? nil) as Any,
            "name": (self.application?.localizedName! ?? nil) as Any,
        ]
    }

}

func getPidPath(pid: Int32) -> String? {
    let pathBuffer = UnsafeMutablePointer<UInt8>.allocate(capacity: Int(MAXPATHLEN))
    defer {
        pathBuffer.deallocate()
    }
    let pathLength = proc_pidpath(pid, pathBuffer, UInt32(MAXPATHLEN))

    if pathLength > 0 {
        let path = String(cString: pathBuffer)
        return path
    }
    return nil
}
