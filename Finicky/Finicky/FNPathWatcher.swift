//
//  FNPathWatcher.swift
//  Finicky
//
//  Created by John Sterling on 07/07/15.
//  Copyright (c) 2015 John sterling. All rights reserved.
//

import Foundation

open class FNPathWatcher {

    enum State {
        case on, off
    }

    fileprivate let source: DispatchSource
    fileprivate let descriptor: CInt
    fileprivate var state: State = .off

    /// Creates a folder monitor object with monitoring enabled.
    public init(url: URL, handler: ()->Void) {

        state = .off
        descriptor = open((url as NSURL).fileSystemRepresentation, O_EVTONLY)
        let queue = DispatchQueue.global()
        source = DispatchSource.makeFileSystemObjectSource(fileDescriptor: descriptor, eventMask:
            DispatchSource.FileSystemEvent(rawValue: UInt(
                UInt8(DispatchSource.FileSystemEvent.delete.rawValue) |
                UInt8(DispatchSource.FileSystemEvent.write.rawValue) |
                UInt8(DispatchSource.FileSystemEvent.extend.rawValue) |
                UInt8(DispatchSource.FileSystemEvent.attrib.rawValue) |
                UInt8(DispatchSource.FileSystemEvent.link.rawValue) |
                UInt8(DispatchSource.FileSystemEvent.rename.rawValue) |
                UInt8(DispatchSource.FileSystemEvent.revoke.rawValue)
            )), queue: queue) /*Migrator FIXME: Use DispatchSourceFileSystemObject to avoid the cast*/ as! DispatchSource

        source.setEventHandler(handler: nil)
        //dispatch_source_set_cancel_handler({})
        start()
    }

    /// Starts sending notifications if currently stopped
    open func start() {
        if state == .off {
            state = .on
            source.resume()
        }
    }

    /// Stops sending notifications if currently enabled
    open func stop() {
        if state == .on {
            state = .off
            source.suspend()
        }
    }

    deinit {
        close(descriptor)
        source.cancel()
    }
}
