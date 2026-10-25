#import <Cocoa/Cocoa.h>
#import <ApplicationServices/ApplicationServices.h>
#import <unistd.h>
#import "safari_menu.h"

static bool titleIs(AXUIElementRef element, CFStringRef title) {
    CFTypeRef value = NULL;
    if (AXUIElementCopyAttributeValue(element, kAXTitleAttribute, &value) != kAXErrorSuccess || value == NULL) {
        return false;
    }
    bool match = CFGetTypeID(value) == CFStringGetTypeID() &&
                 CFStringCompare((CFStringRef)value, title, 0) == kCFCompareEqualTo;
    CFRelease(value);
    return match;
}

// Returns a retained reference to the first child of element titled title, or
// NULL. A NULL title matches the first child.
static AXUIElementRef copyChild(AXUIElementRef element, CFStringRef title) {
    CFTypeRef children = NULL;
    if (AXUIElementCopyAttributeValue(element, kAXChildrenAttribute, &children) != kAXErrorSuccess || children == NULL) {
        return NULL;
    }
    AXUIElementRef found = NULL;
    for (CFIndex i = 0; i < CFArrayGetCount((CFArrayRef)children); i++) {
        AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex((CFArrayRef)children, i);
        if (title == NULL || titleIs(child, title)) {
            found = (AXUIElementRef)CFRetain(child);
            break;
        }
    }
    CFRelease(children);
    return found;
}

static CFIndex windowCount(AXUIElementRef app) {
    CFIndex count = 0;
    if (AXUIElementGetAttributeValueCount(app, kAXWindowsAttribute, &count) != kAXErrorSuccess) {
        return -1;
    }
    return count;
}

int safariOpenWindowFromMenu(const char *menuTitle, double timeoutSeconds) {
    @autoreleasepool {
        // Pressing the menu happens in this process so that Finicky's own
        // Accessibility permission applies. Driving it from an osascript child
        // fails, because the check is made against osascript instead.
        // Asking with the prompt option makes macOS offer the grant instead of
        // failing silently on first use.
        NSDictionary *options = @{(id)kAXTrustedCheckOptionPrompt: @YES};
        if (!AXIsProcessTrustedWithOptions((CFDictionaryRef)options)) {
            return SAFARI_MENU_NOT_TRUSTED;
        }

        NSArray *running = [NSRunningApplication runningApplicationsWithBundleIdentifier:@"com.apple.Safari"];
        if ([running count] == 0) {
            return SAFARI_MENU_NOT_RUNNING;
        }
        NSRunningApplication *safari = [running objectAtIndex:0];
        // The url is handed over afterwards with `open`, which lands in the
        // profile of Safari's active window, so the new window has to be it.
        [safari activateWithOptions:0];

        CFStringRef title = CFStringCreateWithCString(NULL, menuTitle, kCFStringEncodingUTF8);
        AXUIElementRef app = AXUIElementCreateApplication([safari processIdentifier]);
        int result = SAFARI_MENU_NO_ITEM;

        CFTypeRef menuBar = NULL;
        AXUIElementRef fileItem = NULL;
        AXUIElementRef fileMenu = NULL;
        AXUIElementRef item = NULL;

        // Menu titles are localized; this matches English Safari, and a missing
        // item is reported rather than guessed at.
        if (AXUIElementCopyAttributeValue(app, kAXMenuBarAttribute, &menuBar) == kAXErrorSuccess && menuBar != NULL) {
            fileItem = copyChild((AXUIElementRef)menuBar, CFSTR("File"));
        }
        if (fileItem != NULL) {
            fileMenu = copyChild(fileItem, NULL);
        }
        if (fileMenu != NULL) {
            item = copyChild(fileMenu, title);
        }

        if (item != NULL) {
            // Without a starting count a new window cannot be told apart, so
            // the press is skipped rather than reported as a success.
            CFIndex before = windowCount(app);
            result = SAFARI_MENU_NO_WINDOWS;
            if (before >= 0) {
                AXUIElementPerformAction(item, kAXPressAction);
                result = SAFARI_MENU_TIMEOUT;
            }
            for (double waited = 0; before >= 0 && waited < timeoutSeconds; waited += 0.1) {
                if (windowCount(app) > before) {
                    result = SAFARI_MENU_OK;
                    break;
                }
                usleep(100000);
            }
        }

        if (item != NULL) CFRelease(item);
        if (fileMenu != NULL) CFRelease(fileMenu);
        if (fileItem != NULL) CFRelease(fileItem);
        if (menuBar != NULL) CFRelease(menuBar);
        CFRelease(app);
        CFRelease(title);
        return result;
    }
}
