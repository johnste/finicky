import AppKit
import Foundation
import JavaScriptCore

public typealias Callback<T> = (T) -> Void
public typealias Callback2<T, U> = (T, U) -> Void
public typealias OptionsCb = (_ hideIcon: Bool,
                              _ shortUrlProviders: [String]?,
                              _ checkForUpdate: Bool) -> Void
public typealias ConfigPathProvider = () -> String?

/*
 FinickyConfig deals with everything related to the config file.

 It does several things:
    - Run validation of the config file
    - Reads global app settings and uses a callback to let the app get them
    - Deals with everything related to the javascript config file
    - Runs the process that determines the browser we should start
 */
open class FinickyConfig {

    private struct Constants {
        static let defaultConfigPath: NSString = "~/.finicky.js"
    }

    static var defaultConfigLocation: URL {
        let path = Constants.defaultConfigPath.resolvingSymlinksInPath
        return URL(fileURLWithPath: path)
    }

    var ctx: JSContext!
    var configAPIString: String
    var configAPI: JSValue?
    var hasError: Bool = false
    var configObject: JSValue?

    var dispatchSource: DispatchSourceFileSystemObject?
    var fileDescriptor: Int32 = -1
    var lastModificationDate: Date?
    var logToConsole: Callback2<String, Bool>
    var configureAppOptions: OptionsCb
    var updateStatus: Callback<Status>?
    var configPathProvider: ConfigPathProvider

    public enum Status {
        case unavailable
        case invalid
        case valid
    }

    public init(configureAppCb: @escaping OptionsCb,
                logCb: @escaping Callback2<String, Bool>,
                updateStatusCb: @escaping Callback<Status>,
                configPath: @escaping ConfigPathProvider) {
        configAPIString = loadJS("finickyConfigAPI.js")

        configureAppOptions = configureAppCb
        logToConsole = logCb
        updateStatus = updateStatusCb
        configPathProvider = configPath
        listenToChanges(showInitialSuccess: false)
    }

    func waitForFile() {
        Timer.scheduledTimer(withTimeInterval: 2.0, repeats: true, block: { timer in
            guard let configPath = self.configPathProvider(),
                  open(configPath, O_EVTONLY) != -1 else {
                return
            }
            timer.invalidate()
            self.listenToChanges(showInitialSuccess: true)
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

        guard let configPath = self.configPathProvider() else {
            logToConsole("No config file specified", false)
            showNotification(at: .default, title: "No config file specified", subtitle: "You need to create a new one", informativeText: "Config > Create new...", error: true)
            return
        }

        fileDescriptor = open(configPath, O_EVTONLY)
        reload(from: configPath, showSuccess: showInitialSuccess)

        guard fileDescriptor != -1 else {
            print("Couldn't find or read the file. Error: \(String(describing: strerror(errno)))")
            waitForFile()
            updateStatus?(.unavailable)
            return
        }

        lastModificationDate = getModificationDate(atPath: configPath)

        dispatchSource =
            DispatchSource.makeFileSystemObjectSource(fileDescriptor: fileDescriptor, eventMask: [.attrib, .delete], queue: DispatchQueue.main)

        dispatchSource?.setEventHandler { [weak self] in
            guard let self = self else { return }
            print("Detected file change")
            if let modificationDate = getModificationDate(atPath: configPath) {
                if !(self.lastModificationDate != nil) || modificationDate > self.lastModificationDate! {
                    print("Reloading config")
                    self.lastModificationDate = modificationDate
                    self.reload(from: configPath, showSuccess: true)
                }
            } else {
                self.updateStatus?(.unavailable)
                self.waitForFile()
            }
        }

        dispatchSource?.setCancelHandler { [weak self] in
            self?.resetFileDescriptor()
        }

        dispatchSource?.resume()
    }

    @discardableResult
    open func createJSContext() -> JSContext {
        let ctx : JSContext = JSContext()

        ctx.exceptionHandler = {
            (_: JSContext!, exception: JSValue!) in
            self.hasError = true
            self.updateStatus?(.invalid)

            let stacktrace = exception.objectForKeyedSubscript("stack").toString()
            let lineNumber = exception.objectForKeyedSubscript("line").toString()
            let columnNumber = exception.objectForKeyedSubscript("column").toString()
            let message = "Error parsing config: \"\(String(describing: exception!))\" \nStack: \(stacktrace!):\(lineNumber!):\(columnNumber!)"
            let shortMessage = "Configuration: \(String(describing: exception!))"
            print(message)
            showNotification(title: "Configuration", informativeText: String(describing: exception!), error: true)
            self.logToConsole(shortMessage, false)
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
                updateStatus?(invalid ? .invalid : .valid)
                if invalid {
                    let message = "Invalid config"
                    print(message)
                    showNotification(title: message, error: true)
                    logToConsole(message, false)

                    return false
                } else {
                    return true
                }
            } else {
                updateStatus?(.invalid)
            }
        }

        return false
    }

    func reload(from configPath: String, showSuccess: Bool) {
        print("Reloading config")
        hasError = false
        var config: String?

        usleep(100 * 1000) // Sleep for a few millisconds to make sure file is available (See https://github.com/johnste/finicky/issues/140)


        logToConsole("Trying to read config file from: \(configPath)", false)
        
        do {
            config = try String(contentsOfFile: configPath, encoding: .utf8)
        } catch let error as NSError {
            config = nil
            print("\(error.localizedDescription)", terminator: "")
        }

        if config == nil {
            hasError = true
            let message = "Config file could not be read or found"
            showNotification(title: message, subtitle: "Click here for help", informativeText: "Path: \(configPath)", error: true)
            logToConsole("Config file could not be read or found. * Example configuration: \n" + """
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
            """, false)

            return
        }

        setupAPI()
        ctx.evaluateScript(config)
        configObject = ctx.evaluateScript("module.exports")

        if config != nil {
            let success = parseConfig(configObject!)
            if success {
                configureAppOptions(
                    getSimpleOption(name: "hideIcon", defaultValue: false),
                    getShortUrlProviders(),
                    getSimpleOption(name: "checkForUpdate", defaultValue: true)
                )

                if showSuccess {
                    showNotification(title: "Reloaded config successfully")
                    logToConsole("Reloaded config successfully", false)
                }
            }
        }
    }

    func getSimpleOption<T>(name: String, defaultValue: T) -> T {
        if name.count == 0 {
            print("Tried to get an option with no name")
            return defaultValue
        }

        let path = "module.exports.options && module.exports.options." + name

        if let result = ctx.evaluateScript(path) {
            if JSValueIsUndefined(ctx.jsGlobalContextRef, result.jsValueRef) {
                return defaultValue
            }

            if T.self == Bool.self {
                let bool = (result.toBool() as? T)
                return bool ?? defaultValue
            } else {
                print("This type is not yet supported")
            }
        }

        return defaultValue
    }

    func getShortUrlProviders() -> [String]? {
        guard var urlShorteners = ctx.evaluateScript("module.exports.options && module.exports.options.urlShorteners || null") else {
            return defaultUrlShorteners
        }

        if urlShorteners.isNull {
            return defaultUrlShorteners
        }

        if urlShorteners.isArray {
            let list = urlShorteners.toArray() as! [String]?
            return list
        }

        urlShorteners = (ctx.evaluateScript("module.exports.options.urlShorteners")?.call(withArguments: [defaultUrlShorteners]))!

        if urlShorteners.isArray {
            let list = urlShorteners.toArray() as! [String]?
            return list
        }

        return defaultUrlShorteners
    }

    func determineOpeningApp(url: URL, opener: Application) -> AppDescriptor? {
        if let appValue = getConfiguredAppValue(url: url, opener: opener) {
            if !appValue.isObject {
                return nil
            }

            if let browsersArray = appValue.forProperty("browsers")?.toArray() {
                let browsers = browsersArray.compactMap { (raw) -> BrowserOpts? in

                    let dict = raw as! NSMutableDictionary
                    let appType = AppDescriptorType(rawValue: dict["appType"] as! String)
                    let openInBackground: Bool? = dict["openInBackground"] as? Bool
                    let browserName = dict["name"] as! String
                    let browserProfile: String? = dict["profile"] as? String
                    let args: [String] = dict["args"] as? [String] ?? []

                    if browserName == "" {
                        return nil
                    }

                    do {
                        // Default to opening the application in the bg if Finicky is not activated.
                        let browser = try BrowserOpts(
                            name: browserName,
                            appType: appType!,
                            openInBackground: openInBackground,
                            profile: browserProfile,
                            args: args
                        )
                        return browser
                    } catch _ as BrowserError {
                        showNotification(title: "Couldn't find browser \"\(browserName)\"")
                        logToConsole("Couldn't find browser \"\(browserName)\"", false)
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
                        logToConsole("Couldn't generate url from handler \(newUrl), falling back to original url", false)
                    }
                }

                return AppDescriptor(browsers: browsers, url: finalUrl)
            }
        }

        return nil
    }

    func getConfiguredAppValue(url: URL, opener: Application) -> JSValue? {
        let optionsDict = [
            "opener": opener.serialize() as Any,
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
        FinickyAPI.setLog(logToConsole)
        FinickyAPI.setContext(ctx)
        ctx.setObject(FinickyAPI.self, forKeyedSubscript: "finickyInternalAPI" as NSCopying & NSObjectProtocol)
        ctx.evaluateScript("var finicky = finickyConfigApi.createAPI();")
    }
}
