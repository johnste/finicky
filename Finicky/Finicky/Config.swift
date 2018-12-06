import Foundation
import JavaScriptCore

var FNConfigPath: String = "~/.finicky.js"

open class FinickyConfig {

    var configPaths: NSMutableSet
    var ctx: JSContext!
    var defaultBrowser : String?;
    var javascript : String?;

    public init() {
        self.configPaths = NSMutableSet()

        if let path = Bundle.main.path(forResource: "getConfiguredApp.js", ofType: nil ) {
            do {
                javascript = try String(contentsOfFile: path, encoding: String.Encoding.utf8)
            } catch {

            }
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
            showNotification(title: "Parse error when reading config file", subtitle: String(describing: exception))
        }

        ctx.evaluateScript("var module = {}")

        self.setupAPI(ctx)
        return ctx
    }

    open func parseConfig(_ config: String) {
        ctx.evaluateScript(config)

        defaultBrowser = ctx.evaluateScript("module.exports.defaultApp").toString()
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

    open func determineOpeningApp(url: URL) -> String {
        var app = getConfiguredApp(url: url)

        if (app == nil) {
            app = defaultBrowser
        }

        return "hej"
    }

    func getConfiguredApp(url: URL) -> String? {

        if (javascript == nil) {
            return nil;
        }

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

        let result = ctx.evaluateScript(javascript!)?.call(withArguments: [url.absoluteString, urlDict])

        if ((result?.isString)!) {
            return (result?.toString())!
        }

        return nil
    }

    open func setupAPI(_ ctx: JSContext) {
        FinickyAPI.setContext(ctx)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "finicky" as NSCopying & NSObjectProtocol)
    }
}
