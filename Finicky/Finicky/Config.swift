import AppKit
import Foundation
import JavaScriptCore

var FNConfigPath: String = "~/.finicky.js"

public typealias Callback<T> = (T) -> Void

open class FinickyConfig {
    var ctx: JSContext!
    var configAPIString: String
    var configAPI: JSValue?
    var hasError: Bool = false
    var configObject: JSValue?

    var dispatchSource: DispatchSourceFileSystemObject?
    var fileDescriptor: Int32 = -1
    var lastModificationDate: Date?
    var toggleIconCallback: Callback<Bool>?
    var logToConsole: Callback<String>?
    var setShortUrlProviders: Callback<[String]?>?
    var updateStatus: Callback<Bool>?

    public init() {
        configAPIString = loadJS("finickyConfigAPI.js")
    }

    public convenience init(toggleIconCallback: @escaping Callback<Bool>, logToConsoleCallback: @escaping Callback<String>, setShortUrlProviders: @escaping Callback<[String]?>, updateStatus: @escaping Callback<Bool>) {
        self.init()
        self.toggleIconCallback = toggleIconCallback
        logToConsole = logToConsoleCallback
        self.setShortUrlProviders = setShortUrlProviders
        self.updateStatus = updateStatus
        listenToChanges(showInitialSuccess: false)
    }

    func waitForFile() {
        let filename: String = (FNConfigPath as NSString).resolvingSymlinksInPath
        Timer.scheduledTimer(withTimeInterval: 2.0, repeats: true, block: { timer in
            let fileDescriptor = open(filename, O_EVTONLY)
            if fileDescriptor != -1 {
                timer.invalidate()
                self.listenToChanges(showInitialSuccess: true)
            }
        })
    }

    func resetFileDescriptor() {
        if fileDescriptor != -1 {
            close(fileDescriptor)
            fileDescriptor = -1
            dispatchSource = nil
        }
    }

    func listenToChanges(showInitialSuccess: Bool = false) {
        print("Start listening to file changes")

        resetFileDescriptor()

        let filename: String = (FNConfigPath as NSString).resolvingSymlinksInPath
        fileDescriptor = open(filename, O_EVTONLY)

        reload(showSuccess: showInitialSuccess)

        guard fileDescriptor != -1 else {
            print("Couldn't find or read the file. Error: \(String(describing: strerror(errno)))")
            waitForFile()
            updateStatus?(false)
            return
        }

        lastModificationDate = getModificationDate(atPath: filename)

        dispatchSource =
            DispatchSource.makeFileSystemObjectSource(fileDescriptor: fileDescriptor, eventMask: [.attrib, .delete], queue: DispatchQueue.main)

        dispatchSource?.setEventHandler { [weak self] in
            print("Detected file change")
            if let modificationDate = getModificationDate(atPath: filename) {
                if !(self!.lastModificationDate != nil) || modificationDate > self!.lastModificationDate! {
                    print("Reloading config")
                    self!.lastModificationDate = modificationDate
                    self!.reload(showSuccess: true)
                }
            } else {
                self!.updateStatus?(false)
                self!.waitForFile()
            }
        }

        dispatchSource?.setCancelHandler {
            self.resetFileDescriptor()
        }

        dispatchSource?.resume()
    }

    @discardableResult
    open func createJSContext() -> JSContext {
        let ctx: JSContext = JSContext()

        ctx.exceptionHandler = {
            (_: JSContext!, exception: JSValue!) in
            self.hasError = true
            self.updateStatus?(false)

            let stacktrace = exception.objectForKeyedSubscript("stack").toString()
            let lineNumber = exception.objectForKeyedSubscript("line").toString()
            let columnNumber = exception.objectForKeyedSubscript("column").toString()
            let message = "Error parsing config: \"\(String(describing: exception!))\" \nStack: \(stacktrace!):\(lineNumber!):\(columnNumber!)"
            let shortMessage = "Configuration: \(String(describing: exception!))"
            print(message)
            showNotification(title: "Configuration", informativeText: String(describing: exception!), error: true)
            self.logToConsole?(shortMessage)
        }

        ctx.evaluateScript("const module = {}")
        configAPI = ctx.evaluateScript(configAPIString)

        return ctx
    }

    @discardableResult
    open func parseConfig(_: JSValue) -> Bool {
        if hasError {
            return false
        }

        let validConfig = ctx.evaluateScript("finickyConfigApi.validateConfig")?.call(withArguments: [configObject!])

        if let isBoolean = validConfig?.isBoolean {
            if isBoolean {
                print("Valid config ✅")
            } else {
                print("Invalid config 🚫")
            }

            if isBoolean {
                let invalid = !(validConfig?.toBool())!
                updateStatus?(!invalid)
                if invalid {
                    let message = "Invalid config"
                    print(message)
                    showNotification(title: message, error: true)

                    logToConsole?(message)

                    return false
                } else {
                    return true
                }
            } else {
                updateStatus?(false)
            }
        }

        return false
    }

    func reload(showSuccess: Bool) {
        print("Reloading config")
        hasError = false
        var config: String?

        do {
            let filename: String = (FNConfigPath as NSString).resolvingSymlinksInPath
            config = try String(contentsOfFile: filename, encoding: String.Encoding.utf8)
        } catch let error as NSError {
            config = nil
            print("\(error.localizedDescription)", terminator: "")
        }

        if config == nil {
            hasError = true
            let message = "Config file could not be read or found"
            showNotification(title: message, subtitle: "Click here for help", error: true)
            logToConsole?("Config file could not be read or found. * Example configuration: \n" + """
                /**
                * Save as ~/.finicky.js
                */
                module.exports = {
                    defaultBrowser: "Safari",
                    handlers: [
                        {
                            match: finicky.matchHostnames(["youtube.com", "facebook.com"]),
                            browser: "Google Chrome"
                        }
                    ]
                };
                // For more examples, see the Finicky github page https://github.com/johnste/finicky
            """)

            return
        }

        setupAPI()

        ctx.evaluateScript(config)
        configObject = ctx.evaluateScript("module.exports")

        if config != nil {
            let success = parseConfig(configObject!)
            if success {
                toggleIconCallback?(getHideIcon())

                setShortUrlProviders?(getShortUrlProviders())

                if showSuccess {
                    showNotification(title: "Reloaded config successfully")

                    logToConsole?("Reloaded config successfully")
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
        let list = urlShorteners as! [String]?
        if list?.count == 0 {
            return nil
        }
        return list
    }

    open func determineOpeningApp(url: URL, sourceBundleIdentifier: String? = nil) -> AppDescriptor? {
        if let appValue = getConfiguredAppValue(url: url, sourceBundleIdentifier: sourceBundleIdentifier) {
            if !appValue.isObject {
                return nil
            }

            if let browsersArray = appValue.forProperty("browsers")?.toArray() {
                let browsers = browsersArray.compactMap { (raw) -> BrowserOpts? in

                    let dict = raw as! NSMutableDictionary
                    let appType = AppDescriptorType(rawValue: dict["appType"] as! String)
                    let openInBackground: Bool? = dict["openInBackground"] as? Bool
                    let browserName = dict["name"] as! String

                    if browserName == "" {
                        return nil
                    }

                    do {
                        let browser = try BrowserOpts(name: browserName, appType: appType!, openInBackground: openInBackground)
                        return browser
                    } catch _ as BrowserError {
                        showNotification(title: "Couldn't find browser \"\(browserName)\"")
                        logToConsole?("Couldn't find browser \"\(browserName)\"")
                        return nil
                    } catch let msg {
                        print("Unknown error resolving browser: \(msg)")
                        showNotification(title: "Unknown error resolving browser", subtitle: msg.localizedDescription)
                        return nil
                    }
                }

                var finalUrl = url

                if let newUrl = appValue.forProperty("url")?.toString() {
                    if let rewrittenUrl = URL(string: newUrl) {
                        finalUrl = rewrittenUrl
                    } else {
                        logToConsole?("Couldn't generate url from handler \(newUrl), falling back to original url")
                    }
                }

                return AppDescriptor(browsers: browsers, url: finalUrl)
            }
        }

        return nil
    }

    func getConfiguredAppValue(url: URL, sourceBundleIdentifier: String?) -> JSValue? {
        let optionsDict = [
            "sourceBundleIdentifier": sourceBundleIdentifier as Any,
            "keys": getModifierKeyFlags(),
        ] as [AnyHashable: Any]
        let result: JSValue? = ctx.evaluateScript("finickyConfigApi.processUrl")?.call(withArguments: [configObject!, url.absoluteString, optionsDict])
        return result
    }

    func getModifierKeyFlags() -> [String: Bool] {
        return [
            "shift": NSEvent.modifierFlags.contains(.shift),
            "option": NSEvent.modifierFlags.contains(.option),
            "command": NSEvent.modifierFlags.contains(.command),
            "control": NSEvent.modifierFlags.contains(.control),
            "capsLock": NSEvent.modifierFlags.contains(.capsLock),
            "function": NSEvent.modifierFlags.contains(.function),
        ]
    }

    open func setupAPI() {
        ctx = createJSContext()

        if logToConsole != nil {
            FinickyAPI.setLog(logToConsole!)
        }
        FinickyAPI.setContext(ctx)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "finickyInternalAPI" as NSCopying & NSObjectProtocol)
        ctx.evaluateScript("var finicky = finickyConfigApi.createAPI();")
    }
}
