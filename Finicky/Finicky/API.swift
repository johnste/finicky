import Foundation
import JavaScriptCore

@objc protocol FinickyAPIExports : JSExport {
    static func log(_ message: String?) -> Void
    static func notify(_ title: JSValue, _ subtitle: JSValue) -> Void
    static func getUrlParts(_ url: String) -> Dictionary<String, Any>
}

@objc open class FinickyAPI : NSObject, FinickyAPIExports {

    fileprivate static var context : JSContext! = nil
    fileprivate static var logToConsole:((_ message: String) -> Void)? = nil;

    static func log(_ message: String?) -> Void {
        if message != nil {
            NSLog(message!)
            logToConsole!(message!)
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

    @objc class func setContext(_ context: JSContext) {
        self.context = context
    }

    @objc class func setLog(_ logToConsole: @escaping (_ message: String) -> Void) {
        self.logToConsole = logToConsole
    }

    @objc public class func getUrlParts(_ urlString: String) -> Dictionary<String, Any> {
        let url: URL! = URL.init(string: urlString)

        guard url != nil else { return [:] }

        let _protocol = url.scheme ?? ""
        let username = url.user ?? ""
        let password = url.password ?? ""
        let host = url.host ?? ""
        let port = url.port ?? nil
        let pathname = url.path
        let search = url.query ?? ""
        let hash = url.fragment ?? ""

        let urlDict = [
            "hash": hash as Any,
            "host": host as Any,
            "protocol": _protocol as Any,
            "port": port as Any,
            "username": username as Any,
            "password": password as Any,
            "pathname": pathname,
            "search": search as Any,
        ]

        return urlDict
    }
}
