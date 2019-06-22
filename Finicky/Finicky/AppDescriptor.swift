import Foundation

public enum AppDescriptorType: String {
    case bundleId
    case appName
    case none
}

public struct AppDescriptor {
    public var name: String
    public var appType: AppDescriptorType
    public var url: URL
    public var openInBackground: Bool?

    public init(name: String, appType: AppDescriptorType, url: URL, openInBackground: Bool?) {
        self.name = name
        self.appType = appType
        self.url = url
        self.openInBackground = openInBackground
    }
}
