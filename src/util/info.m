#import "info.h"
#import <Cocoa/Cocoa.h>

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