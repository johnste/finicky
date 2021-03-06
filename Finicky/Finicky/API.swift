import AppKit
import Foundation
import JavaScriptCore

@objc protocol FinickyAPIExports: JSExport {
    static func log(_ message: String?) -> Void
    static func notify(_ title: JSValue, _ subtitle: JSValue) -> Void
    static func getBattery() -> NSDictionary?
    static func getKeys() -> [String: Bool]
    static func getSystemInfo() -> [String: String]
}

/*
 FinickyAPI represents the internal API methods available javascript context.
 The _actual_ API available in .finicks.js is created in the config-api and uses these methods.
 */
@objc open class FinickyAPI: NSObject, FinickyAPIExports {
    fileprivate static var context: JSContext!
    fileprivate static var logToConsole: ((_ message: String, _ clearConsole: Bool) -> Void)?

    static func log(_ message: String?) {
        if message != nil {
            NSLog(message!)
            logToConsole?(message!, false)
        }
    }

    static func getBattery() -> NSDictionary? {
        if let batteryStatus = getBatteryStatus() {
            return [
                "isCharging": batteryStatus.isCharging,
                "isPluggedIn": batteryStatus.isPluggedIn,
                "chargePercentage": batteryStatus.chargePercentage,
            ]
        }

        return nil
    }

    static func getKeys() -> [String: Bool] {
        return [
            "shift": NSEvent.modifierFlags.contains(.shift),
            "option": NSEvent.modifierFlags.contains(.option),
            "command": NSEvent.modifierFlags.contains(.command),
            "control": NSEvent.modifierFlags.contains(.control),
            "capsLock": NSEvent.modifierFlags.contains(.capsLock),
            "function": NSEvent.modifierFlags.contains(.function),
        ]
    }

    static func getSystemInfo() -> [String: String] {
        return [
            "localizedName": Host.current().localizedName ?? "",
            "name": Host.current().name ?? "",
        ]
    }

    static func notify(_ title: JSValue, _ informativeText: JSValue) {
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

    @objc class func setLog(_ logToConsole: @escaping (_ message: String, _ clearConsole: Bool) -> Void) {
        self.logToConsole = logToConsole
    }
}
