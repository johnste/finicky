
import Foundation
import Carbon

enum Browser: String {
    case Chrome = "com.google.Chrome"
    case ChromeCanary = "com.google.Chrome.canary"
    case Brave = "com.brave.Browser"
    case BraveDev = "com.brave.Browser.dev"
    case Safari = "com.apple.Safari"
    case Firefox = "org.mozilla.firefox"
    case FirefoxDeveloperEdition = "org.mozilla.firefoxdeveloperedition"
    case Opera = "com.operasoftware.Opera"
}

public func startBrowser(app: BrowserOpts, url: URL, defaultOpenInBackground: Bool) {

    if (app.bundleId == Browser.Chrome.rawValue) {
        let chromeScript = """
            on openIncognito(newurl)
                if application "Google Chrome" is running then
                    tell application "Google Chrome"
                        set incognitoWindows to (every window whose mode is "incognito")

                        if (count of incognitoWindows) > 0 then
                            set incognitoWindow to first window whose mode is "incognito"
                            tell incognitoWindow to make new tab at after (get active tab) with properties {URL:newurl}
                        else
                            set incognitoWindow to make new window with properties {mode:"incognito"}
                            tell incognitoWindow to set URL of active tab to newurl
                        end if
                    end tell
                else
                    do shell script "open -a \"Google Chrome\" --args {URL: newurl} --incognito"
                end if
            end openIncognito
        """
        executeScript(chromeScript, url: url.absoluteString)
        return
    }

    var command = ["open", url.absoluteString, "-b", app.bundleId]

    if app.openInBackground ?? defaultOpenInBackground {
        command.append("-g")
    }

    shell(command)
}

func executeScript(_ script: String, url: String) {
    let parameters = NSAppleEventDescriptor.list()
    parameters.insert(NSAppleEventDescriptor(string: url), at: 0)

    let scriptObject = NSAppleScript(source: script)
    var error: NSDictionary?
    scriptObject?.compileAndReturnError(&error)


    let event = NSAppleEventDescriptor(
        eventClass: AEEventClass(kASAppleScriptSuite),
        eventID: AEEventID(kASSubroutineEvent),
        targetDescriptor: nil,
        returnID: AEReturnID(kAutoGenerateReturnID),
        transactionID: AETransactionID(kAnyTransactionID)
    )
    event.setDescriptor(NSAppleEventDescriptor(string: "openIncognito"), forKeyword: AEKeyword(keyASSubroutineName))
    event.setDescriptor(parameters, forKeyword: AEKeyword(keyDirectObject))

    print("hello")
    let result = scriptObject?.executeAppleEvent(event, error: &error)


//    if let scriptObject = NSAppleScript(source: script) {
//
//        if let output: NSAppleEventDescriptor = scriptObject.executeAndReturnError(
//                                                                           &error) {
//            print(output.stringValue!)
//        } else if (error != nil) {
//            showNotification(title: "error: \(String(describing: error))")
//        }
//    }
}


