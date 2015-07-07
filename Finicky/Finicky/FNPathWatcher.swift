//
//  FNPathWatcher.swift
//  Finicky
//
//  Created by John Sterling on 07/07/15.
//  Copyright (c) 2015 John sterling. All rights reserved.
//

import Foundation

public class FNPathWatcher {
    
    enum State {
        case On, Off
    }
    
    private let source: dispatch_source_t
    private let descriptor: CInt
    private var state: State = .Off
    
    /// Creates a folder monitor object with monitoring enabled.
    public init(url: NSURL, handler: ()->Void) {
        
        state = .Off
        descriptor = open(url.fileSystemRepresentation, O_EVTONLY)
        let queue = dispatch_get_global_queue(DISPATCH_QUEUE_PRIORITY_DEFAULT, 0)
        source = dispatch_source_create(
            DISPATCH_SOURCE_TYPE_VNODE,
            UInt(descriptor),
            DISPATCH_VNODE_DELETE | DISPATCH_VNODE_WRITE | DISPATCH_VNODE_EXTEND | DISPATCH_VNODE_ATTRIB | DISPATCH_VNODE_LINK | DISPATCH_VNODE_RENAME | DISPATCH_VNODE_REVOKE,
            queue
        )

        dispatch_source_set_event_handler(source, handler)
        //dispatch_source_set_cancel_handler({})
        start()
    }
    
    /// Starts sending notifications if currently stopped
    public func start() {
        if state == .Off {
            state = .On
            dispatch_resume(source)
        }
    }
    
    /// Stops sending notifications if currently enabled
    public func stop() {
        if state == .On {
            state = .Off
            dispatch_suspend(source)
        }
    }
    
    deinit {
        close(descriptor)
        dispatch_source_cancel(source)
    }
}