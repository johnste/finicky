import Foundation
import JavaScriptCore

@objc protocol FinickyAPIExports: JSExport {
    static func log(_ message: String?) -> Void
    static func notify(_ title: JSValue, _ subtitle: JSValue) -> Void
}

@objc open class FinickyAPI: NSObject, FinickyAPIExports {
    fileprivate static var context: JSContext!
    fileprivate static var logToConsole: ((_ message: String, _ clearConsole: Bool) -> Void)?

    static func log(_ message: String?) {
        if message != nil {
            NSLog(message!)
            logToConsole?(message!, false)
        }
    }

    static func notify(_ title: JSValue, _ informativeText: JSValue) {
        if title.isString {
            if informativeText.isString {
                showNotification(title: title.toString(), informativeText: informativeText.toString())
            } else {
                showNotification(title: title.toString())
            }
        }
    }

    @objc class func setContext(_ context: JSContext) {
        self.context = context
    }

    @objc class func setLog(_ logToConsole: @escaping (_ message: String, _ clearConsole: Bool) -> Void) {
        self.logToConsole = logToConsole
    }
}
