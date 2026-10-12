#import <Cocoa/Cocoa.h>
#import "picker.h"

// Builds and runs the alert. Must be called on the main thread.
static int runPickerOnMainThread(NSString *title, NSString *message, NSArray<NSString *> *labels) {
    // Finicky always has an NSApplication by this point, but guarding here lets
    // the picker be exercised from a plain binary with no app bundle.
    if (NSApp == nil) {
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    }

    NSAlert *alert = [[NSAlert alloc] init];
    [alert setMessageText:title];
    [alert setInformativeText:message];
    [alert setAlertStyle:NSAlertStyleInformational];

    for (NSString *label in labels) {
        [alert addButtonWithTitle:label];
    }
    // Titling the last button "Cancel" gives it Escape automatically.
    [alert addButtonWithTitle:@"Cancel"];

    // Finicky normally runs without a dock icon, so the alert needs an explicit
    // activation to come to the front of whichever app the link was clicked in.
    [NSApp activateIgnoringOtherApps:YES];

    NSModalResponse response = [alert runModal];
    NSInteger index = response - NSAlertFirstButtonReturn;
    if (index < 0 || index >= (NSInteger)[labels count]) {
        return -1;
    }
    return (int)index;
}

int showBrowserPicker(const char *title, const char *message, const char **labels, int count) {
    @autoreleasepool {
        NSMutableArray<NSString *> *items = [NSMutableArray arrayWithCapacity:count];
        for (int i = 0; i < count; i++) {
            NSString *label = labels[i] ? [NSString stringWithUTF8String:labels[i]] : @"";
            [items addObject:label];
        }

        NSString *titleStr = title ? [NSString stringWithUTF8String:title] : @"Open link";
        NSString *messageStr = message ? [NSString stringWithUTF8String:message] : @"";

        __block int result = -1;
        if ([NSThread isMainThread]) {
            result = runPickerOnMainThread(titleStr, messageStr, items);
        } else {
            // URLs are handled on a background goroutine, but AppKit is main-thread only.
            dispatch_sync(dispatch_get_main_queue(), ^{
                result = runPickerOnMainThread(titleStr, messageStr, items);
            });
        }
        return result;
    }
}
