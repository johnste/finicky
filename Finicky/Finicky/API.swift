import Foundation
import JavaScriptCore

@objc protocol FinickyAPIExports : JSExport {
    static func setDefaultBrowser(_ browser: String) -> Void

    static func log(_ message: String?) -> Void
    static func notify(_ title: JSValue, _ subtitle: JSValue) -> Void

    static func isDomain(_ url: String, _ domain: JSValue ) -> JSValue

    static func onUrl(_ handler: JSValue) -> Void
    static func onDomain(_ domain: JSValue, _ bundleIdentifier: JSValue) -> Void
}

struct URLHandler {
    let callback : JSValue?
    let predicate : ((_ url: String) -> Bool)?
    let bundleIdentifier : String?

    init(callback: JSValue) {
        self.callback = callback
        self.predicate = nil
        self.bundleIdentifier = nil
    }

    init(bundleIdentifier: String, predicate: @escaping (_ url: String) -> Bool) {
        self.bundleIdentifier = bundleIdentifier
        self.predicate = predicate
        self.callback = nil
    }
}

@objc open class FinickyAPI : NSObject, FinickyAPIExports {

    fileprivate static var urlHandlers = Array<URLHandler>()
    fileprivate static var context : JSContext! = nil

    class func setDefaultBrowser(_ browser: String) -> Void {
        // AppDelegate.defaultBrowser = browser
    }

    static func log(_ message: String?) -> Void {
        if message != nil {
            NSLog(message!)
        }
    }

    static func notify(_ title: JSValue, _ subtitle: JSValue) -> Void {
        if title.isString {
            if subtitle.isString {
                showNotification(title: title.toString(), subtitle: subtitle.toString())
            } else {
                showNotification(title: title.toString())
            }
        }

    }

    open class func onUrl(_ handler: JSValue) -> Void {
        urlHandlers.append(URLHandler(callback: handler))
    }

    @objc open class func reset() -> Void {
        urlHandlers.removeAll(keepingCapacity: true)
    }

    @objc class func setContext(_ context: JSContext) {
        self.context = context
    }

    open class func isDomain(_ url: String, _ maybeDomains: JSValue ) -> JSValue {
        let isMatch = self.matchesDomain(url, maybeDomains)
        return JSValue(bool: isMatch, in: self.context)
    }

    class func match(value: JSValue, match: String ) -> Bool {
        if value.isString {
            let matchesHost = value.toString() == match
            return matchesHost
        }

        return false
    }

    open class func matchesDomain(_ url: String, _ maybeDomains: JSValue ) -> Bool {
        let urlObj = URL(string: url)

        if urlObj == nil || urlObj!.host == nil{
            return false
        }

        let host = urlObj!.host!

        if maybeDomains.isArray {
            for maybe in maybeDomains.toArray() {
                if match(value: maybe as! JSValue, match: host) {
                    return true
                }
            }
        } else {
            return match(value: maybeDomains, match: host)
        }

        return false
    }

    open class func onDomain(_ domain: JSValue, _ bundleIdentifier: JSValue) -> Void {

        func predicate(_ url: String) -> Bool {
            return self.matchesDomain(url, domain)
        }

        urlHandlers.append(URLHandler(bundleIdentifier: bundleIdentifier.toString(), predicate: predicate))
    }


    /**
        Get strategy from registered handlers

        @param originalUrl The original url that triggered finicky to start

        @param sourceBundleIdentifier Bundle identifier of the application that triggered the url to open

        @return A dictionary keyed with "url" and "bundleIdentifier" with
            the new url and bundle identifier to spawn
    */

    @objc open class func callUrlHandlers(_ originalUrl: URL, sourceBundleIdentifier: String?, flags : Dictionary<String, Bool>) -> Dictionary<String, AnyObject> {
        var strategy : Dictionary<String, AnyObject> = [
            "url": originalUrl.absoluteString as AnyObject
        ]

        var sourceBundleId : AnyObject
        if sourceBundleIdentifier != nil {
            sourceBundleId = sourceBundleIdentifier! as AnyObject
        } else {
            sourceBundleId = JSValue(nullIn: context)
        }

        let options : Dictionary<String, AnyObject> = [
            "sourceBundleIdentifier": sourceBundleId,
            "flags": flags as AnyObject
        ]


        for handler in urlHandlers {

            let url = strategy["url"]! as! String

            if let predicate = handler.predicate {
                if predicate(url) && handler.bundleIdentifier != nil {
                    strategy["bundleIdentifier"] = handler.bundleIdentifier! as AnyObject
                    return strategy
                }
            }

            if let callback = handler.callback {
                let val = callback.call(withArguments: [url, options])

                if (val?.isUndefined)! {
                    continue
                }

                let handlerStrategy = val?.toDictionary()
                if handlerStrategy == nil {
                    continue
                }

                if handlerStrategy?["url"] != nil {
                    strategy["url"] = (handlerStrategy?["url"] as! String as AnyObject)
                }

                let bundleId: AnyObject? = handlerStrategy?["bundleIdentifier"] as AnyObject
                if bundleId != nil {
                    strategy["bundleIdentifier"] = bundleId!
                }

                if handlerStrategy?["openInBackground"] != nil {
                    strategy["openInBackground"] = (handlerStrategy?["openInBackground"] as! Bool as AnyObject)
                }

                return strategy
            }
        }
        return strategy
    }
}
