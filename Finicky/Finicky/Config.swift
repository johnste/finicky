import Foundation
import JavaScriptCore

var FNConfigPath: String = "~/.finicky.js"

public enum AppDescriptorType: String {
    case bundleId
    case appName
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

open class FinickyConfig {
    var ctx: JSContext!
    var validateConfigJS : String?;
    var processUrlJS : String?;
    var validateLibJS : String?;
    var hasError: Bool = false;

    var dispatchSource: DispatchSourceFileSystemObject?
    var fileDescriptor: Int32 = -1
    var fileManager = FileManager.init();
    var lastModificationDate: Date? = nil;
    var toggleIconCallback:((_ hide: Bool) -> Void)?;
    var logToConsole:((_ message: String) -> Void)?;
    var setShortUrlProviders:((_ urlShorteners: [String]?) -> Void)?;

    public init() {
        listenToChanges();

        if let path = Bundle.main.path(forResource: "validate.js", ofType: nil ) {
            validateLibJS = try! String(contentsOfFile: path, encoding: String.Encoding.utf8)
        }

        if let path = Bundle.main.path(forResource: "validateConfig.js", ofType: nil ) {
            validateConfigJS = try! String(contentsOfFile: path, encoding: String.Encoding.utf8)
        }

        if let path = Bundle.main.path(forResource: "processUrl.js", ofType: nil ) {
            processUrlJS = try! String(contentsOfFile: path, encoding: String.Encoding.utf8)
        }
    }

    public convenience init(toggleIconCallback: @escaping (_ hide: Bool) -> Void, logToConsoleCallback: @escaping (_ message: String) -> Void , setShortUrlProviders: @escaping (_ shortUrlProviders: [String]?) -> Void) {
        self.init();
        self.toggleIconCallback = toggleIconCallback
        self.logToConsole = logToConsoleCallback
        self.setShortUrlProviders = setShortUrlProviders

    }

    func getModificationDate(atPath: String) -> Date? {
        do {
            let attributes = try self.fileManager.attributesOfItem(atPath: atPath)
            let modificationDate = attributes[FileAttributeKey.modificationDate] as? Date
            guard modificationDate != nil else { return nil }
            return modificationDate!
        } catch (let msg) {
            print("Error message: \(msg)")
            return nil
        }
    }

    func listenToChanges() {

        let filename: String = (FNConfigPath as NSString).resolvingSymlinksInPath

        guard dispatchSource == nil && fileDescriptor == -1 else { return }

        fileDescriptor = open(filename, O_EVTONLY)
        if (fileDescriptor <= 0) {
            print(strerror(errno));
        }

        guard fileDescriptor != -1 else { return }

        lastModificationDate = getModificationDate(atPath: filename)

        dispatchSource =
            DispatchSource.makeFileSystemObjectSource(fileDescriptor: fileDescriptor, eventMask: .attrib, queue: DispatchQueue.main)

        dispatchSource?.setEventHandler { [weak self] in
            if let modificationDate = self?.getModificationDate(atPath: filename) {
                if !(self!.lastModificationDate != nil) || modificationDate > self!.lastModificationDate! {
                    self!.lastModificationDate = modificationDate
                    self!.reload(showSuccess: true)
                }
            }
        }

        dispatchSource?.setCancelHandler {
            close(self.fileDescriptor)
            self.fileDescriptor = -1
            self.dispatchSource = nil
        }

        dispatchSource?.resume()
    }

    open func createContext() -> JSContext {
        ctx = JSContext()

        ctx.exceptionHandler = {
            context, exception in
                self.hasError = true;
                let message = "Error parsing config: \"\(String(describing: exception!))\"";
                print(message)
                showNotification(title: "Error parsing config", informativeText: String(describing: exception!), error: true)
                if (self.logToConsole != nil) {
                    self.logToConsole!(message)
                }
        }

        ctx.evaluateScript("const module = {}")

        ctx.evaluateScript(validateLibJS!);

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
                if (!(validConfig?.toBool())!) {
                    let message = "Invalid config"
                    print(message)
                    showNotification(title: message, error: true)
                    if (self.logToConsole != nil) {
                        self.logToConsole!(message)
                    }
                    return false;
                } else {
                    return true;
                }
            }
        }

        return false;
    }

    func reload(showSuccess: Bool) {
        self.hasError = false;
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
            let message = "Config file could not be read or found"
            showNotification(title: message, subtitle: "Click here to show example config file", error: true)
            print(message)
            if (self.logToConsole != nil) {
                self.logToConsole!(message + "\n\n" + """
                    // --------------------------------------------------------------
                    // Example config, save as ~/.finicky.js
                    module.exports = {
                        defaultBrowser: "Safari",
                        handlers: [
                            {
                                match: /^https?:\\/\\/(youtube|facebook|twitter|linkedin|keep\\.google)\\.com/,
                                app: "Google Chrome"
                            }
                        ]
                    };
                    // --------------------------------------------------------------
                """)
            }
            return
        }

        if let theError = error {
            print("\(theError.localizedDescription)", terminator: "")
        }

        ctx = createContext()
        if config != nil {
            let success = parseConfig(config!)
            if (success) {
                if (self.toggleIconCallback != nil) {
                    self.toggleIconCallback!(getHideIcon())
                }
                if (self.setShortUrlProviders != nil) {
                    self.setShortUrlProviders!(getShortUrlProviders());
                }

                if (showSuccess) {
                    showNotification(title: "Reloaded config successfully")
                    if (self.logToConsole != nil) {
                        self.logToConsole!("Reloaded config successfully")
                    }
                }
            }
        }
    }

    open func getHideIcon() -> Bool {
        let hideIcon = ctx.evaluateScript("module.exports.options && module.exports.options.hideIcon")?.toBool()

        if hideIcon == nil {
            return false
        }
        return hideIcon!;
    }

    open func getShortUrlProviders() -> [String]? {
        let urlShorteners = ctx.evaluateScript("module.exports.options && module.exports.options.urlShorteners || []")?.toArray()
        return urlShorteners as! [String]?;
    }

    open func determineOpeningApp(url: URL, sourceBundleIdentifier: String? = nil) -> AppDescriptor? {
        let appValue = getConfiguredAppValue(url: url, sourceBundleIdentifier: sourceBundleIdentifier)
       
        if ((appValue?.isObject)!) {
            let dict = appValue?.toDictionary()
            let appType = AppDescriptorType(rawValue: dict!["appType"] as! String)

            var finalUrl = url

            if let newUrl = dict!["url"] as? String {
                if let rewrittenUrl = URL.init(string: newUrl) {
                    finalUrl = rewrittenUrl
                } else if (self.logToConsole != nil){
                    self.logToConsole!("Couldn't generate url from handler \(newUrl), falling back to original url")
                }
            }

            if (appType == nil) {
                let message = "Unrecognized app type \"\(String(describing: appType))\""
                showNotification(title: message, error: true)
                if (self.logToConsole != nil){
                    self.logToConsole!(message)
                }
            } else {
                let openInBackground = dict!["openInBackground"] as? Bool;
                return AppDescriptor(name: dict!["name"] as! String, appType: appType!, url: finalUrl, openInBackground: openInBackground)
            }
        }

        return nil
    }

    func getConfiguredAppValue(url: URL, sourceBundleIdentifier: String?) -> JSValue? {
        let optionsDict = [
            "sourceBundleIdentifier": sourceBundleIdentifier as Any,
            ] as [AnyHashable : Any]
        let result = ctx.evaluateScript(processUrlJS!)?.call(withArguments: [url.absoluteString, optionsDict])
        return result
    }

    open func setupAPI(_ ctx: JSContext) {
        if (self.logToConsole != nil) {
            FinickyAPI.setLog(logToConsole!)
        }
        FinickyAPI.setContext(ctx)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "finicky" as NSCopying & NSObjectProtocol)
    }
}
