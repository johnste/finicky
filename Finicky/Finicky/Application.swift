import AppKit
import Foundation

class Application {
    var application: NSRunningApplication?
    var path: String?
    var pid: Int32

    init(pid: Int32) {
        let application = NSRunningApplication(processIdentifier: pid)
        path = getPidPath(pid: pid)
        self.pid = pid
        self.application = application
    }

    func serialize() -> [String: Any] {
        return [
            "pid": pid,
            "path": path as Any,
            "bundleId": (application?.bundleIdentifier ?? nil) as Any,
            "name": (application?.localizedName ?? nil) as Any,
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
