import AppKit
import Foundation

public enum AppDescriptorType: String {
    case bundleId
    case appName
    case appPath
    case none
}

enum BrowserError: Error {
    case cantFindBrowser(msg: String)
}

public struct BrowserOpts : CustomStringConvertible {
    public var name: String
    public var openInBackground: Bool
    public var bundleId: String?
    public var appPath: String?
    
    public var description: String {
        if let bundleId = self.bundleId {
            return bundleId
        } else if let appPath = self.appPath {
            return appPath
        } else {
            return self.name
        }
    }

    public init(name: String, appType: AppDescriptorType, openInBackground: Bool?) throws {
        self.name = name
        self.openInBackground = openInBackground ?? false

        if appType == AppDescriptorType.bundleId {
            self.bundleId = name
        } else if appType == AppDescriptorType.appPath {
            self.appPath = name
        } else if let path = NSWorkspace.shared.fullPath(forApplication: name) {
            if let bundle = Bundle(path: path) {
                self.bundleId = bundle.bundleIdentifier!
                self.appPath = path
            } else {
                throw BrowserError.cantFindBrowser(msg: "Couldn't find application \"\(name)\"")
            }
        } else {
            throw BrowserError.cantFindBrowser(msg: "Couldn't find application \"\(name)\"")
        }
    }
    
    public static func isAppDirectory(_ path: String) -> Bool {
        var isDirectory = ObjCBool(true)
        let exists = FileManager.default.fileExists(atPath: path, isDirectory: &isDirectory)
        return exists && isDirectory.boolValue
    }
}

public struct AppDescriptor {
    public var url: URL
    public var browsers: [BrowserOpts]

    public init(browsers: [BrowserOpts], url: URL) {
        self.browsers = browsers
        self.url = url
    }
}
