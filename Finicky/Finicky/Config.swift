import Foundation
import JavaScriptCore

var FNConfigPath: String = "~/.finicky.js"

public enum AppDescriptorType: String {
    case bundleId
    case appName
}

public struct AppDescriptor {
    public var value: String
    public var type: AppDescriptorType
    public var url: URL
    public var openInBackground: Bool?

    public init(value: String, type: AppDescriptorType, url: URL, openInBackground: Bool?) {
        self.value = value
        self.type = type
        self.url = url
        self.openInBackground = openInBackground
    }
}

open class FinickyConfig {
    var configPaths: NSMutableSet
    var ctx: JSContext!
    var validateConfigJS : String?;
    var processUrlJS : String?;
    var hasError: Bool;

    public init() {
        self.configPaths = NSMutableSet()
        self.hasError = false;

        if let path = Bundle.main.path(forResource: "validateConfig.js", ofType: nil ) {
            do {
                validateConfigJS = try String(contentsOfFile: path, encoding: String.Encoding.utf8)
            }
            catch {
                validateConfigJS = nil
            }
        }

        if let path = Bundle.main.path(forResource: "processUrl.js", ofType: nil ) {
            do {
                processUrlJS = try String(contentsOfFile: path, encoding: String.Encoding.utf8)}
            catch {
                processUrlJS = nil
            }
        }
    }

    func resetConfigPaths() {
        self.hasError = false;
        FinickyAPI.reset()
        configPaths.removeAllObjects()
        configPaths.add(FNConfigPath)
    }

    open func createContext() -> JSContext {
        ctx = JSContext()

        ctx.exceptionHandler = {
            context, exception in
                self.hasError = true;
                print("Error parsing config: \"\(String(describing: exception!))\"")
                showNotification(title: "Error parsing config", informativeText: String(describing: exception!))
        }

        ctx.evaluateScript("const module = {}")
        self.setupAPI(ctx)
        return ctx
    }

    open func parseConfig(_ config: String) -> Bool {
        ctx.evaluateScript(config)

        if (self.hasError) {
            return false;
        }

        let validConfig = ctx.evaluateScript(validateConfigJS!)?.call(withArguments: [])
        if let isBoolean = validConfig?.isBoolean {
            if (isBoolean) {
                let result = validConfig?.toBool()
                if (!result!) {
                    print("Invalid config")
                    showNotification(title: "Invalid config")
                    return false;
                }
            }
        }

        return true;
    }

    func reload(showSuccess: Bool) {
        self.resetConfigPaths()
        var error:NSError?
        let filename: String = (FNConfigPath as NSString).standardizingPath
        var config: String?
        do {
            config = try String(contentsOfFile: filename, encoding: String.Encoding.utf8)
        } catch let configError as NSError {
            error = configError
            config = nil
        }

        if config == nil {
            showNotification(title: "Config file could not be read or found")
            print("Config file could not be read or found")
            return
        }

        if let theError = error {
            print("\(theError.localizedDescription)", terminator: "")
        }

        ctx = createContext()
        if config != nil {
            let success = parseConfig(config!)
            if (success && showSuccess) {
                showNotification(title: "Reloaded config successfully")
            }
        }
    }

    open func determineOpeningApp(url: URL) -> AppDescriptor? {
        let appValue = getConfiguredAppValue(url: url)

        if ((appValue?.isObject)!) {
            let dict = appValue?.toDictionary()

            let type = AppDescriptorType(rawValue: dict!["type"] as! String)


            var finalUrl = url
            if let newUrl = dict!["url"] as? String {
                finalUrl = URL.init(string: newUrl) ?? url
            }

            if (type == nil) {
                showNotification(title: "Unrecognized app type \"\(String(describing: type))\"")
            } else {
                return AppDescriptor(value: dict!["value"] as! String, type: type!, url: finalUrl, openInBackground: dict!["openInBackground"] as! Bool?)
            }
        }

        return nil
    }

    func createUrlDict(url: URL) -> Dictionary<String, Any> {
        let _protocol = url.scheme ?? nil
        let username = url.user ?? nil
        let password = url.password ?? nil
        let host = url.host ?? nil
        let port = url.port ?? nil
        let pathname = url.path
        let search = url.query ?? nil
        let hash = url.fragment ?? nil

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

    func getConfiguredAppValue(url: URL) -> JSValue? {
        let urlDict = createUrlDict(url: url)
        let result = ctx.evaluateScript(processUrlJS!)?.call(withArguments: [url.absoluteString, urlDict])
        return result
    }

    open func setupAPI(_ ctx: JSContext) {
        FinickyAPI.setContext(ctx)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "finicky" as NSCopying & NSObjectProtocol)
    }
}
