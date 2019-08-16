import AppKit
import Foundation

public enum AppDescriptorType: String {
    case bundleId
    case appName
    case none
}

enum BrowserError: Error {
    case cantFindBrowser(msg: String)
}

public struct BrowserOpts {
    public var name: String
    public var openInBackground: Bool?
    public var bundleId: String

    public init(name: String, appType: AppDescriptorType, openInBackground: Bool?) throws {
        self.name = name
        self.openInBackground = openInBackground

        if appType == AppDescriptorType.bundleId {
            bundleId = name
        } else if let path = NSWorkspace.shared.fullPath(forApplication: name) {
            if let bundle = Bundle(path: path) {
                bundleId = bundle.bundleIdentifier!
            } else {
                throw BrowserError.cantFindBrowser(msg: "Couldn't find application \"\(name)\"")
            }
        } else {
            throw BrowserError.cantFindBrowser(msg: "Couldn't find application \"\(name)\"")
        }
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
