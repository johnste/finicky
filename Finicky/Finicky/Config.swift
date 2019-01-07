import Foundation
import JavaScriptCore

var FNConfigPath: String = "~/.finicky.js"


enum AppDescriptorType: String {
    case bundleId
    case appName
}

public struct AppDescriptor {
    var value: String
    var type: AppDescriptorType?


    init(value: String, type: String) {
        self.value = value
        self.type = AppDescriptorType(rawValue: type)

        if (self.type == nil) {
            showNotification(title: "Unrecognized app type \"\(String(describing: type))\"")
        }
    }
}

open class FinickyConfig {

    var configPaths: NSMutableSet
    var ctx: JSContext!
    var validateConfigJS : String?;
    var processUrlJS : String?;

    public init() throws {
        self.configPaths = NSMutableSet()

        if let path = Bundle.main.path(forResource: "validateConfig.js", ofType: nil ) {
            validateConfigJS = try String(contentsOfFile: path, encoding: String.Encoding.utf8)
        }

        if let path = Bundle.main.path(forResource: "processUrl.js", ofType: nil ) {
            processUrlJS = try String(contentsOfFile: path, encoding: String.Encoding.utf8)
        }
    }

    func resetConfigPaths() {
        FinickyAPI.reset()
        configPaths.removeAllObjects()
        configPaths.add(FNConfigPath)
    }

    open func createContext() -> JSContext {
        ctx = JSContext()

        ctx.exceptionHandler = {
            context, exception in
                print("Error when reading config file: \"\(String(describing: exception!))\"")
                showNotification(title: "Error when reading config file", subtitle: String(describing: exception!))
        }

        ctx.evaluateScript("const module = {}")

        self.setupAPI(ctx)
        return ctx
    }

    open func parseConfig(_ config: String) {
        ctx.evaluateScript(config)

        let validConfig = ctx.evaluateScript(validateConfigJS!)?.call(withArguments: [])
        if let isBoolean = validConfig?.isBoolean {
            if (isBoolean) {
                let result = validConfig?.toBool()
                if (!result!) {
                    print("Invalid config")
                    showNotification(title: "Invalid config")
                }
            }

        }

    }

    func reload() {
        self.resetConfigPaths()
        var error:NSError?
        let filename: String = (FNConfigPath as NSString).standardizingPath
        var config: String?
        do {
            config = try String(contentsOfFile: filename, encoding: String.Encoding.utf8)
        } catch let error1 as NSError {
            error = error1
            config = nil
        }

        if config == nil {
            print("Config file could not be read or found")
            return
        }

        if let theError = error {
            print("\(theError.localizedDescription)", terminator: "")
        }

        ctx = createContext()
        if config != nil {
            parseConfig(config!)
        }
    }

    open func determineOpeningApp(url: URL) -> AppDescriptor? {
        let appValue = getConfiguredAppValue(url: url)

        if ((appValue?.isObject)!) {
            let dict = appValue?.toDictionary()
            return AppDescriptor(value: dict!["value"] as! String, type: dict!["type"] as! String)
        }

        return nil
    }

    func getConfiguredAppValue(url: URL) -> JSValue? {
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
            "pathname": pathname as Any,
            "search": search as Any,
        ]        

        let result = ctx.evaluateScript(processUrlJS!)?.call(withArguments: [url.absoluteString, urlDict])

        return result
    }

    open func setupAPI(_ ctx: JSContext) {
        FinickyAPI.setContext(ctx)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "finicky" as NSCopying & NSObjectProtocol)
    }
}
