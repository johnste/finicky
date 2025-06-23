#import "info.h"
#import <Cocoa/Cocoa.h>
#import <IOKit/ps/IOPSKeys.h>
#import <IOKit/ps/IOPowerSources.h>
#import <stdlib.h>
#import <string.h>

ModifierKeys getModifierKeys() {
    NSEventModifierFlags flags = [NSEvent modifierFlags];
    ModifierKeys keys = {
        .shift = (flags & NSEventModifierFlagShift) != 0,
        .option = (flags & NSEventModifierFlagOption) != 0,
        .command = (flags & NSEventModifierFlagCommand) != 0,
        .control = (flags & NSEventModifierFlagControl) != 0,
        .capsLock = (flags & NSEventModifierFlagCapsLock) != 0,
        .fn = (flags & NSEventModifierFlagFunction) != 0
    };
    return keys;
}

SystemInfo getSystemInfo() {
    NSHost *currentHost = [NSHost currentHost];
    NSString *localizedNameStr = [currentHost localizedName] ?: @"";
    NSString *nameStr = [currentHost name] ?: @"";

    SystemInfo info = {
        .localizedName = [localizedNameStr UTF8String],
        .name = [nameStr UTF8String]
    };
    return info;
}

PowerInfo getPowerInfo() {
    PowerInfo info = {
        .isConnected = false,
        .isCharging = false,
        .percentage = -1
    };

    CFTypeRef powerSourcesInfo = IOPSCopyPowerSourcesInfo();
    if (!powerSourcesInfo) {
        return info;
    }

    CFArrayRef powerSources = IOPSCopyPowerSourcesList(powerSourcesInfo);
    if (!powerSources) {
        CFRelease(powerSourcesInfo);
        return info;
    }

    CFIndex count = CFArrayGetCount(powerSources);
    if (count == 0) {
        CFRelease(powerSources);
        CFRelease(powerSourcesInfo);
        return info;
    }

    // Get the first power source (usually the internal battery)
    CFDictionaryRef powerSource = CFArrayGetValueAtIndex(powerSources, 0);
    if (!powerSource) {
        CFRelease(powerSources);
        CFRelease(powerSourcesInfo);
        return info;
    }

    // Check if it's a battery
    CFStringRef powerSourceType = CFDictionaryGetValue(powerSource, CFSTR(kIOPSTypeKey));
    if (!powerSourceType || !CFEqual(powerSourceType, CFSTR(kIOPSInternalBatteryType))) {
        CFRelease(powerSources);
        CFRelease(powerSourcesInfo);
        return info;
    }

    // Get battery information
    CFBooleanRef isChargingRef = CFDictionaryGetValue(powerSource, CFSTR(kIOPSIsChargingKey));
    if (isChargingRef) {
        info.isCharging = CFBooleanGetValue(isChargingRef);
    }

    CFStringRef powerSourceStateRef = CFDictionaryGetValue(powerSource, CFSTR(kIOPSPowerSourceStateKey));
    if (powerSourceStateRef) {
        info.isConnected = CFEqual(powerSourceStateRef, CFSTR(kIOPSACPowerValue));
    }

    CFNumberRef percentageRef = CFDictionaryGetValue(powerSource, CFSTR(kIOPSCurrentCapacityKey));
    if (percentageRef) {
        CFNumberGetValue(percentageRef, kCFNumberIntType, &info.percentage);
    }

    CFRelease(powerSources);
    CFRelease(powerSourcesInfo);

    return info;
}

_Bool isAppRunning(const char* identifier) {
    if (!identifier) {
        return 0;
    }

    NSString *identifierStr = [NSString stringWithUTF8String:identifier];
    NSArray *runningApps = [[NSWorkspace sharedWorkspace] runningApplications];

    for (NSRunningApplication *app in runningApps) {
        // Check bundle ID
        if ([[app bundleIdentifier] isEqualToString:identifierStr]) {
            return 1;
        }

        // Check app name
        NSString *appName = [app localizedName];
        if ([appName isEqualToString:identifierStr]) {
            return 1;
        }
    }

    return 0;
}

const char* getNSHomeDirectory(void) {
    NSString *homeDirString = NSHomeDirectory();
    return [homeDirString UTF8String];
}

const char* getNSCacheDirectory(void) {
    NSArray *paths = NSSearchPathForDirectoriesInDomains(NSCachesDirectory, NSUserDomainMask, YES);
    if (paths.count > 0) {
        NSString *cacheDirString = [paths objectAtIndex:0];
        return [cacheDirString UTF8String];
    }
    return NULL;
}
