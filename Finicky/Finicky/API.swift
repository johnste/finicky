import Foundation
import JavaScriptCore

@objc protocol FinickyAPIExports : JSExport {
    static func log(_ message: String?) -> Void
    static func notify(_ title: JSValue, _ subtitle: JSValue) -> Void
}

@objc open class FinickyAPI : NSObject, FinickyAPIExports {

    fileprivate static var context : JSContext! = nil

    static func log(_ message: String?) -> Void {
        if message != nil {
            NSLog(message!)
        }
    }

    static func notify(_ title: JSValue, _ informativeText: JSValue) -> Void {
        if title.isString {
            if informativeText.isString {
                showNotification(title: title.toString(), informativeText: informativeText.toString())
            } else {
                showNotification(title: title.toString())
            }
        }
    }

    @objc open class func reset() -> Void {
    }

    @objc class func setContext(_ context: JSContext) {
        self.context = context
    }

}
