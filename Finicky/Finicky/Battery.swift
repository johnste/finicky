import Foundation
import IOKit.ps
import JavaScriptCore

/*
 Battery utilities
 */

enum BatteryError: Error { case error }

@objc protocol BatteryProtocol: JSExport {
    var chargePercentage: Int { get }
    var isCharging: Bool { get }
    var isPluggedIn: Bool { get }
}

@objc class BatteryStatus: NSObject, BatteryProtocol {
    let chargePercentage: Int
    let isCharging: Bool
    let isPluggedIn: Bool

    public init(chargePercentage: Int, isCharging: Bool, isPluggedIn: Bool) {
        self.chargePercentage = chargePercentage
        self.isCharging = isCharging
        self.isPluggedIn = isPluggedIn
    }
}

func getBatteryStatus() -> BatteryStatus? {
    do {
        // Take a snapshot of all the power source info
        guard let snapshot = IOPSCopyPowerSourcesInfo()?.takeRetainedValue()
        else { throw BatteryError.error }

        // Pull out a list of power sources
        guard let sources: NSArray = IOPSCopyPowerSourcesList(snapshot)?.takeRetainedValue()
        else { throw BatteryError.error }

        // For each power source...
        for ps in sources {
            // Fetch the information for a given power source out of our snapshot
            guard let info: NSDictionary = IOPSGetPowerSourceDescription(snapshot, ps as CFTypeRef)?.takeUnretainedValue()
            else { throw BatteryError.error }

            // Pull out the name and current capacity
            if let capacity = info[kIOPSCurrentCapacityKey] as? Int,
               let isCharging = info[kIOPSIsChargingKey] as? Bool,
               let powerSource = info[kIOPSPowerSourceStateKey] as? String
            {
                let isPluggedIn = powerSource == kIOPSACPowerValue
                return BatteryStatus(chargePercentage: capacity, isCharging: isCharging, isPluggedIn: isPluggedIn)
            }
        }
    } catch {
        return nil
    }

    return nil
}
