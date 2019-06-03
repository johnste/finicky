import Foundation
import JavaScriptCore

var FNConfigPath: String = "~/.finicky.js"

public typealias Callback<T> = (T) -> Void

open class FinickyConfig {
    var ctx: JSContext!
    var validateConfigJS : String;
    var processUrlJS : String;
    var fastidiousLibJS : String;
    var jsAPIJS: String;
    var hasError: Bool = false;

    var dispatchSource: DispatchSourceFileSystemObject?
    var fileDescriptor: Int32 = -1
    var lastModificationDate: Date? = nil;
    var toggleIconCallback: Callback<Bool>?;
    var logToConsole: Callback<String>?;
    var setShortUrlProviders: Callback<[String]?>?
    var updateStatus: Callback<Bool>?;

    public init() {
        fastidiousLibJS = loadJS("fastidious.js")
        validateConfigJS = loadJS("validateConfig.js")
        processUrlJS = loadJS("processUrl.js")
        jsAPIJS = loadJS("jsAPI.js")
    }

    public convenience init(toggleIconCallback: @escaping Callback<Bool>, logToConsoleCallback: @escaping Callback<String> , setShortUrlProviders: @escaping Callback<[String]?>, updateStatus: @escaping Callback<Bool>) {
        self.init();
        self.toggleIconCallback = toggleIconCallback
        self.logToConsole = logToConsoleCallback
        self.setShortUrlProviders = setShortUrlProviders
        self.updateStatus = updateStatus
        listenToChanges(showInitialSuccess: false);
    }

    func waitForFile() {
        let filename: String = (FNConfigPath as NSString).resolvingSymlinksInPath
        Timer.scheduledTimer(withTimeInterval: 2.0, repeats: true, block: { timer in
            let fileDescriptor = open(filename, O_EVTONLY)
            if (fileDescriptor != -1) {
                timer.invalidate()
                self.listenToChanges(showInitialSuccess: true)
            }
        })
    }

    func resetFileDescriptor() {
        print("Cancel fileDescriptor")
        if (self.fileDescriptor != -1) {
            close(self.fileDescriptor)
            self.fileDescriptor = -1
            self.dispatchSource = nil
        }
    }

    func listenToChanges(showInitialSuccess : Bool = false) {
        print("Start listening to file changes")

        resetFileDescriptor()

        let filename: String = (FNConfigPath as NSString).resolvingSymlinksInPath
        fileDescriptor = open(filename, O_EVTONLY)

        self.reload(showSuccess: showInitialSuccess)

        guard fileDescriptor != -1 else {
            print("Couldn't find or read the file. Error: \(String(describing: strerror(errno)))")
            waitForFile()
            self.updateStatus!(false)
            return
        }

        lastModificationDate = getModificationDate(atPath: filename)

        dispatchSource =
            DispatchSource.makeFileSystemObjectSource(fileDescriptor: fileDescriptor, eventMask: [.attrib , .delete] , queue: DispatchQueue.main)

        dispatchSource?.setEventHandler { [weak self] in
            print("Detected file change")
            if let modificationDate = getModificationDate(atPath: filename) {
                if !(self!.lastModificationDate != nil) || modificationDate > self!.lastModificationDate! {
                    print("Reloading config")
                    self!.lastModificationDate = modificationDate
                    self!.reload(showSuccess: true)
                }
            } else {
                self!.updateStatus!(false)
                self!.waitForFile()
            }
        }

        dispatchSource?.setCancelHandler {
            self.resetFileDescriptor();
        }

        dispatchSource?.resume()
    }

    @discardableResult
    open func createJSContext() -> JSContext {
        let ctx : JSContext = JSContext()

        ctx.exceptionHandler = {
            (context: JSContext!, exception: JSValue!) in
                self.hasError = true
                self.updateStatus!(false)
        
                let stacktrace = exception.objectForKeyedSubscript("stack").toString()
                let lineNumber = exception.objectForKeyedSubscript("line").toString()
                let columnNumber = exception.objectForKeyedSubscript("column").toString()
                let message = "Error parsing config: \"\(String(describing: exception!))\" \nStack: \(stacktrace!):\(lineNumber!):\(columnNumber!)";
                let shortMessage = "Configuration: \(String(describing: exception!))";
                print(message)
                showNotification(title: "Configuration", informativeText: String(describing: exception!), error: true)
                if (self.logToConsole != nil) {
                    self.logToConsole!(shortMessage)
                }
        }

        ctx.evaluateScript("const module = {}")
        ctx.evaluateScript(fastidiousLibJS);

        print ("Created new context")
        return ctx
    }

    @discardableResult
    open func parseConfig(_ config: String) -> Bool {
        ctx.evaluateScript(config)

        if (self.hasError) {
            return false;
        }

        let validConfig = ctx.evaluateScript(validateConfigJS)?.call(withArguments: [])

        if let isBoolean = validConfig?.isBoolean {
            print("Valid config: \(isBoolean)")
            if (isBoolean) {
                let invalid = !(validConfig?.toBool())!
                updateStatus!(!invalid)
                if (invalid) {
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
            } else {
                updateStatus!(false)
            }
        }

        return false;
    }

    func reload(showSuccess: Bool) {
        print("Reload config. showSuccess: \(showSuccess)")
        self.hasError = false;
        var config: String?

        do {
            let filename: String = (FNConfigPath as NSString).resolvingSymlinksInPath
            config = try String(contentsOfFile: filename, encoding: String.Encoding.utf8)
        } catch let error as NSError {
            config = nil
            print("\(error.localizedDescription)", terminator: "")
        }

        if config == nil {
            self.hasError = true;
            let message = "Config file could not be read or found"
            showNotification(title: message, subtitle: "Click here for assistance", error: true)
            print(message)
            if (self.logToConsole != nil) {
                self.logToConsole!(message + ". * Example configuration: \n" + """
                    /**
                    * Save as ~/.finicky.js
                    */
                    module.exports = {
                        defaultBrowser: "Safari",
                        handlers: [
                            {
                                match: finicky.matchDomains(["youtube.com", "facebook.com"]),
                                browser: "Google Chrome"
                            }
                        ]
                    };
                    // For more examples, see the Finicky github page https://github.com/johnste/finicky
                """)
            }
            return
        }

        setupAPI()

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

    func getHideIcon() -> Bool {
        let hideIcon = ctx.evaluateScript("module.exports.options && module.exports.options.hideIcon")?.toBool()
        return hideIcon ?? false
    }

    func getShortUrlProviders() -> [String]? {
        let urlShorteners = ctx.evaluateScript("module.exports.options && module.exports.options.urlShorteners || []")?.toArray()
        let list = urlShorteners as! [String]?;
        if (list?.count == 0) {
            return nil;
        }
        return list;
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
            "urlString": url.absoluteString,
            "url": FinickyAPI.getUrlParts(url.absoluteString),
            ] as [AnyHashable : Any]
        let result = ctx.evaluateScript(processUrlJS)?.call(withArguments: [optionsDict])
        return result
    }

    open func setupAPI() {
        self.ctx = createJSContext()

        if (self.logToConsole != nil) {
            FinickyAPI.setLog(self.logToConsole!)
        }
        FinickyAPI.setContext(ctx)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "finicky" as NSCopying & NSObjectProtocol)
        ctx.evaluateScript(jsAPIJS)
    }
}
