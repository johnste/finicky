import AppKit
import Foundation

/*
 Things related to representing the final browser options that should be executed
 */

public enum AppDescriptorType: String {
    case bundleId
    case appName
    case appPath
    case none
}

enum BrowserError: Error {
    case cantFindBrowser(msg: String)
}

public struct BrowserOpts: CustomStringConvertible {
    public var name: String
    public var openInBackground: Bool
    public var bundleId: String?
    public var appPath: String?
    public var profile: String?
    public var args: [String]

    public var description: String {
        if let bundleId = self.bundleId {
            return bundleId
        } else if let appPath = self.appPath {
            return appPath
        } else {
            return name
        }
    }

    public init(
        name: String,
        appType: AppDescriptorType,
        openInBackground: Bool?,
        profile: String?,
        args: [String]
    ) throws {
        self.name = name
        if openInBackground == nil {
            // Open the browser in the background depending on if Finicky recieved focus
            // But we need to sleep for a few millisconds to attempt to work around focus issue
            // See https://github.com/johnste/finicky/issues/126
            usleep(120 * 1000)

            self.openInBackground = !NSApplication.shared.isActive
        } else {
            self.openInBackground = openInBackground!
        }

        self.profile = profile
        self.args = args

        if appType == AppDescriptorType.bundleId {
            bundleId = name
        } else if appType == AppDescriptorType.appPath {
            appPath = name
        } else if let path = NSWorkspace.shared.fullPath(forApplication: name) {
            if let bundle = Bundle(path: path) {
                bundleId = bundle.bundleIdentifier!
                appPath = path
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
